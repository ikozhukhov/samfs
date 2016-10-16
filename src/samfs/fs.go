package samfs

import (
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	pb "github.com/smihir/samfs/src/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type SamFsOptions struct {
	server string
	port   string
}

type SamFs struct {
	pathfs.FileSystem
	Mount     string
	cacheLock sync.RWMutex
	fileCache map[string]*SamFsFileData
	options   *SamFsOptions

	nfsClient  pb.NFSClient
	clientConn *grpc.ClientConn

	rootfh pb.FileHandle

	owner fuse.Owner
}

func NewSamFs(opts *SamFsOptions) (*SamFs, error) {
	samFs := &SamFs{
		options:   opts,
		fileCache: make(map[string]*SamFsFileData),
	}
	conn, err := grpc.DialContext(context.Background(), opts.server+":"+opts.port,
		grpc.WithInsecure(), grpc.WithBackoffMaxDelay(120*time.Second))
	if err != nil {
		return nil, err
	}
	samFs.nfsClient = pb.NewNFSClient(conn)
	samFs.clientConn = conn
	u, uErr := user.Current()
	if uErr != nil {
		return nil, err
	}
	uid, uIdErr := strconv.ParseUint(u.Uid, 10, 32)
	if uIdErr != nil {
		return nil, uIdErr
	}
	gid, gIdErr := strconv.ParseUint(u.Gid, 10, 32)
	if gIdErr != nil {
		return nil, gIdErr
	}

	samFs.owner.Uid = uint32(uid)
	samFs.owner.Gid = uint32(gid)
	glog.Infof("running samfs with uid: %d, gid: %d", uid, gid)

	return samFs, nil
}

func (c *SamFs) SetDebug(debug bool) {
}

func (c *SamFs) getFileHandle(name string) (*pb.FileHandle, fuse.Status) {

	if name == "" {
		return &c.rootfh, fuse.OK
	}
	parentFh := &c.rootfh

	path := strings.Split(name, "/")
	for _, fname := range path {
		resp, err := c.nfsClient.Lookup(context.Background(), &pb.LocalDirectoryRequest{
			DirectoryFileHandle: parentFh,
			Name:                fname,
		})
		if err != nil {
			glog.Errorf(`failed to lookup file "%s" :: %s`, fname, err.Error())
			// error code 5 implies the file does not exist
			if grpc.Code(err) == 5 {
				return nil, fuse.ENOENT
			}
			// TODO(mihir): maybe check for more error codes and translate to appropriate
			// fuse statuses
			return nil, fuse.EIO
		}
		parentFh = resp.FileHandle
	}
	return parentFh, fuse.OK
}

func (c *SamFs) getParentHandle(name string) (*pb.FileHandle, fuse.Status) {

	if name == "" {
		// there is no parent of root as far as samfs is concerned
		return nil, fuse.OK
	}

	path := strings.Split(name, "/")

	// if directory is directly under root, then the parent is root
	if len(path) == 1 {
		return &c.rootfh, fuse.OK
	}

	// remove the name of the current file from the slice, so that we get the
	// path to the parent
	myName := path[len(path)-1]
	parentPath := strings.TrimSuffix(name, "/"+myName)
	glog.Infof("getParentHandle: name %s, parentPath %s", name, parentPath)

	return c.getFileHandle(parentPath)
}

// Attributes.  This function is the main entry point, through
// which FUSE discovers which files and directories exist.
//
// If the filesystem wants to implement hard-links, it should
// return consistent non-zero FileInfo.Ino data.  Using
// hardlinks incurs a performance hit.
func (c *SamFs) GetAttr(name string, fContext *fuse.Context) (*fuse.Attr,
	fuse.Status) {

	glog.Infof(`GetAttr called on "%s"`, name)
	fh, fhErr := c.getFileHandle(name)
	if fhErr != fuse.OK {
		return nil, fhErr
	}
	resp, err := c.nfsClient.GetAttr(context.Background(), &pb.FileHandleRequest{
		FileHandle: fh,
	})
	if err != nil {
		glog.Errorf(`failed to get attributes of file "%s" :: %s`, name, err.Error())
		return nil, fuse.EIO
	}

	fAttr := ProtoToFuseAttr(resp)
	fAttr.Owner = c.owner

	return fAttr, fuse.OK
}

func (c *SamFs) Truncate(path string, size uint64,
	fContext *fuse.Context) fuse.Status {

	glog.Infof("Truncate called on  %s", path)
	return fuse.EINVAL
}

func (c *SamFs) Utimens(name string, atime *time.Time, mtime *time.Time,
	fContext *fuse.Context) fuse.Status {
	glog.Infof("Utimens called on %s", name)

	return fuse.OK
}

func (c *SamFs) Chown(name string, uid uint32, gid uint32,
	fContext *fuse.Context) fuse.Status {

	glog.Info("Chown called")
	return fuse.OK
}

func (c *SamFs) Chmod(name string, mode uint32,
	fContext *fuse.Context) fuse.Status {

	glog.Info("Chmod called")
	return fuse.OK
}

func (c *SamFs) Access(name string, mode uint32,
	fContext *fuse.Context) fuse.Status {

	glog.Info("Access called")
	return fuse.OK
}

func (c *SamFs) Link(orig string, newName string,
	fContext *fuse.Context) fuse.Status {

	glog.Info("Link called")
	return fuse.EINVAL
}

func (c *SamFs) Rmdir(path string, fContext *fuse.Context) fuse.Status {
	glog.Infof("Rmdir called %s", path)

	fh, fhErr := c.getParentHandle(path)
	if fhErr != fuse.OK {
		return fhErr
	}

	splitPath := strings.Split(path, "/")
	name := splitPath[len(splitPath)-1]
	_, err := c.nfsClient.Rmdir(context.Background(), &pb.LocalDirectoryRequest{
		DirectoryFileHandle: fh,
		Name:                name,
	})
	if err != nil {
		glog.Errorf(`failed to remove directory "%s" :: %s`, path, err.Error())
		return fuse.EIO
	}
	return fuse.OK
}

func (c *SamFs) Mkdir(path string, mode uint32,
	fContext *fuse.Context) fuse.Status {

	glog.Infof("Mkdir called for %s", path)
	fh, fhErr := c.getParentHandle(path)
	if fhErr != fuse.OK {
		return fhErr
	}

	splitPath := strings.Split(path, "/")
	name := splitPath[len(splitPath)-1]
	_, err := c.nfsClient.Mkdir(context.Background(), &pb.LocalDirectoryRequest{
		DirectoryFileHandle: fh,
		Name:                name,
	})
	if err != nil {
		glog.Errorf(`failed to create directory "%s" :: %s`, path, err.Error())
		return fuse.EIO
	}
	return fuse.OK
}

func (c *SamFs) Rename(oldName string, newName string,
	fContext *fuse.Context) fuse.Status {

	glog.Info("Rename called")
	return fuse.OK
}

func (c *SamFs) Unlink(name string, fContext *fuse.Context) fuse.Status {
	glog.Infof("Unlink called on %s", name)

	fh, fhErr := c.getParentHandle(name)
	if fhErr != fuse.OK {
		return fhErr
	}

	splitPath := strings.Split(name, "/")
	justName := splitPath[len(splitPath)-1]
	_, err := c.nfsClient.Remove(context.Background(), &pb.LocalDirectoryRequest{
		DirectoryFileHandle: fh,
		Name:                justName,
	})
	if err != nil {
		glog.Errorf(`failed to remove file "%s" :: %s`, name, err.Error())
		return fuse.EIO
	}
	return fuse.OK
}

func (c *SamFs) GetXAttr(name string, attribute string,
	fContext *fuse.Context) ([]byte, fuse.Status) {

	glog.Info("GetXAttr called")
	return []byte{}, fuse.OK
}

func (c *SamFs) RemoveXAttr(name string, attr string,
	fContext *fuse.Context) fuse.Status {

	glog.Info("RemoveXAttr called")
	return fuse.OK
}

func (c *SamFs) SetXAttr(name string, attr string, data []byte, flags int,
	fContext *fuse.Context) fuse.Status {

	glog.Info("SetXAttr called")
	return fuse.OK
}

func (c *SamFs) ListXAttr(name string, fContext *fuse.Context) ([]string,
	fuse.Status) {

	glog.Info("ListXAttr called")
	return []string{}, fuse.OK
}

func (c *SamFs) OnMount(nodefs *pathfs.PathNodeFs) {
	glog.Info("OnMount called")
	resp, err := c.nfsClient.Mount(context.Background(), &pb.MountRequest{})
	if err != nil {
		glog.Fatalf("failed to mount the remote filesystem :: %s", err.Error())
		c.clientConn.Close()
		return
	}
	c.rootfh = *resp.FileHandle
}

func (c *SamFs) OnUnmount() {
	c.clientConn.Close()
	glog.Info("unmount okay")
}

func (c *SamFs) Open(name string, flags uint32,
	fContext *fuse.Context) (nodefs.File, fuse.Status) {

	glog.Infof("Open called on %s", name)
	fh, fhErr := c.getFileHandle(name)
	if fhErr != fuse.OK {
		glog.Errorf(`failed to open file "%s"`, name)
		return nil, fhErr
	}
	fdata := NewFileData(name, c, fh)
	fsFh := NewFileHandle(fdata)
	return &nodefs.WithFlags{
		File: fsFh,
		// NOTE(mihir): if there is some problem wrt fuse, uncomment the
		// line below!
		//FuseFlags: fuse.FOPEN_DIRECT_IO,
	}, fuse.OK
}

func (c *SamFs) OpenDir(name string, fContext *fuse.Context) ([]fuse.DirEntry,
	fuse.Status) {

	glog.Infof(`OpenDir called on "%s"`, name)
	fh, fhErr := c.getFileHandle(name)
	if fhErr != fuse.OK {
		return nil, fhErr
	}
	resp, err := c.nfsClient.Readdir(context.Background(), &pb.FileHandleRequest{
		FileHandle: fh,
	})
	if err != nil {
		glog.Errorf(`failed to read directory "%s" :: %s`, name, err.Error())
		return nil, fuse.EBUSY
	}

	d := make([]fuse.DirEntry, len(resp.Entries))
	for i, e := range resp.Entries {
		d[i].Mode = e.Mode
		d[i].Name = e.Name
	}
	return d, fuse.OK
}

func (c *SamFs) Create(name string, flags uint32, mode uint32,
	fContext *fuse.Context) (nodefs.File, fuse.Status) {

	glog.Infof("Create called %s", name)
	fh, fhErr := c.getParentHandle(name)
	if fhErr != fuse.OK {
		glog.Errorf(`failed to create file "%s"`, name)
		return nil, fhErr
	}

	splitPath := strings.Split(name, "/")
	justName := splitPath[len(splitPath)-1]
	resp, err := c.nfsClient.Create(context.Background(), &pb.LocalDirectoryRequest{
		DirectoryFileHandle: fh,
		Name:                justName,
	})
	if err != nil {
		glog.Errorf(`failed to create file "%s" :: %s`, name, err.Error())
		return nil, fuse.EIO
	}
	fdata := NewFileData(name, c, resp.FileHandle)
	fsFh := NewFileHandle(fdata)
	return fsFh, fuse.OK
}

func (c *SamFs) Symlink(pointedTo string, linkName string,
	fContext *fuse.Context) fuse.Status {

	glog.Info("Symlink called")
	return fuse.OK
}

func (c *SamFs) Readlink(name string, fContext *fuse.Context) (string,
	fuse.Status) {

	glog.Info("Readlink called")
	return "", fuse.OK
}

func (c *SamFs) StatFs(name string) *fuse.StatfsOut {
	glog.Info("StatFs called")
	return &fuse.StatfsOut{}
}
