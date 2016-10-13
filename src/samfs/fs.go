package samfs

import (
	"sync"
	//"syscall"
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
}

func NewSamFs(opts *SamFsOptions) (*SamFs, error) {
	samFs := &SamFs{
		options:   opts,
		fileCache: make(map[string]*SamFsFileData),
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	conn, err := grpc.DialContext(ctx, opts.server+":"+opts.port,
		grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	samFs.nfsClient = pb.NewNFSClient(conn)
	samFs.clientConn = conn

	return samFs, nil
}

func (c *SamFs) SetDebug(debug bool) {
}

// Attributes.  This function is the main entry point, through
// which FUSE discovers which files and directories exist.
//
// If the filesystem wants to implement hard-links, it should
// return consistent non-zero FileInfo.Ino data.  Using
// hardlinks incurs a performance hit.
func (c *SamFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr,
	fuse.Status) {

	glog.Info("GetAttr called")
	return nil, fuse.OK
}

func (c *SamFs) Truncate(path string, size uint64,
	context *fuse.Context) fuse.Status {

	glog.Info("Truncate called")
	return fuse.EINVAL
}

func (c *SamFs) Utimens(name string, atime *time.Time, mtime *time.Time,
	context *fuse.Context) fuse.Status {
	glog.Info("Utimens called")

	return fuse.OK
}

func (c *SamFs) Chown(name string, uid uint32, gid uint32,
	context *fuse.Context) fuse.Status {

	glog.Info("Chown called")
	return fuse.OK
}

func (c *SamFs) Chmod(name string, mode uint32,
	context *fuse.Context) fuse.Status {

	glog.Info("Chmod called")
	return fuse.OK
}

func (c *SamFs) Access(name string, mode uint32,
	context *fuse.Context) fuse.Status {

	glog.Info("Access called")
	return fuse.OK
}

func (c *SamFs) Link(orig string, newName string,
	context *fuse.Context) fuse.Status {

	glog.Info("Link called")
	return fuse.OK
}

func (c *SamFs) Rmdir(path string, context *fuse.Context) fuse.Status {
	glog.Info("Rmdir called")
	return fuse.OK
}

func (c *SamFs) Mkdir(path string, mode uint32,
	context *fuse.Context) fuse.Status {

	glog.Info("Mkdir called")
	return fuse.OK
}

func (c *SamFs) Rename(oldName string, newName string,
	context *fuse.Context) fuse.Status {

	glog.Info("Rename called")
	return fuse.OK
}

func (c *SamFs) Unlink(name string, context *fuse.Context) fuse.Status {
	glog.Info("Unlink called")
	return fuse.OK
}

func (c *SamFs) GetXAttr(name string, attribute string,
	context *fuse.Context) ([]byte, fuse.Status) {

	glog.Info("GetXAttr called")
	return []byte{}, fuse.OK
}

func (c *SamFs) RemoveXAttr(name string, attr string,
	context *fuse.Context) fuse.Status {

	glog.Info("RemoveXAttr called")
	return fuse.OK
}

func (c *SamFs) SetXAttr(name string, attr string, data []byte, flags int,
	context *fuse.Context) fuse.Status {

	glog.Info("SetXAttr called")
	return fuse.OK
}

func (c *SamFs) ListXAttr(name string, context *fuse.Context) ([]string,
	fuse.Status) {

	glog.Info("GetAttr called")
	return []string{}, fuse.OK
}

func (c *SamFs) OnMount(nodefs *pathfs.PathNodeFs) {
	c.clientConn.Close()
	glog.Info("mount ok")
}

func (c *SamFs) OnUnmount() {
	glog.Info("unmount okay")
}

func (c *SamFs) Open(name string, flags uint32,
	context *fuse.Context) (nodefs.File, fuse.Status) {

	glog.Info("Open called")
	return nil, fuse.OK
}

func (c *SamFs) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry,
	fuse.Status) {

	glog.Info("OpenDir called")
	return nil, fuse.OK
}

func (c *SamFs) Create(name string, flags uint32, mode uint32,
	context *fuse.Context) (nodefs.File, fuse.Status) {

	glog.Info("Create called")
	return nil, fuse.OK
}

func (c *SamFs) Symlink(pointedTo string, linkName string,
	context *fuse.Context) fuse.Status {

	glog.Info("Symlink called")
	return fuse.OK
}

func (c *SamFs) Readlink(name string, context *fuse.Context) (string,
	fuse.Status) {

	glog.Info("Readlink called")
	return "", fuse.OK
}

func (c *SamFs) StatFs(name string) *fuse.StatfsOut {
	glog.Info("StatFs called")
	return &fuse.StatfsOut{}
}
