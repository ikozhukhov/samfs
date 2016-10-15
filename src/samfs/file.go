package samfs

import (
	//"bytes"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	pb "github.com/smihir/samfs/src/proto"
	"golang.org/x/net/context"
)

type SamFsFileHandle struct {
	at       int64
	closed   bool
	fileData *SamFsFileData
}

type SamFsFileData struct {
	sync.Mutex
	serverFh *pb.FileHandle
	Fs       *SamFs
	Refs     int32
	Name     string
	Data     []byte
	Dirty    bool
	Attr     *fuse.Attr
}

func NewFileHandle(f *SamFsFileData) *SamFsFileHandle {
	f.Lock()
	f.Refs++
	f.Unlock()
	return &SamFsFileHandle{
		at:       0,
		closed:   false,
		fileData: f,
	}
}

func NewFileData(path string, fs *SamFs, serverFh *pb.FileHandle) *SamFsFileData {

	return &SamFsFileData{
		Refs:     0,
		Fs:       fs,
		Name:     path,
		Dirty:    false,
		serverFh: serverFh,
	}
}

func (c *SamFsFileHandle) String() string {
	glog.Info("String called")
	return c.fileData.Name
}

func (c *SamFsFileHandle) Chmod(mode uint32) fuse.Status {
	glog.Info("Chmod called")
	return fuse.OK
}

func (c *SamFsFileHandle) Chown(uid uint32, gid uint32) fuse.Status {
	glog.Info("Chown called")
	return fuse.OK
}

func (c *SamFsFileHandle) Read(buf []byte, off int64) (fuse.ReadResult,
	fuse.Status) {

	glog.Info("Read called on %s off: %d, size %d", c.fileData.Name, off, len(buf))
	name := c.fileData.Name
	fh := c.fileData.serverFh
	resp, err := c.fileData.Fs.nfsClient.Read(context.Background(),
		&pb.ReadRequest{
			FileHandle: fh,
			Offset:     off,
			Size:       int64(len(buf)),
		})
	if err != nil {
		glog.Errorf(`failed to write to file "%s" :: %s`, name, err.Error())
		var nullData []byte
		return fuse.ReadResultData(nullData), fuse.EIO
	}
	return fuse.ReadResultData(resp.Data), fuse.OK
}

func (c *SamFsFileHandle) Write(data []byte, offset int64) (uint32,
	fuse.Status) {

	glog.Info("Write called on %s", c.fileData.Name)
	name := c.fileData.Name
	fh := c.fileData.serverFh
	_, err := c.fileData.Fs.nfsClient.Write(context.Background(),
		&pb.WriteRequest{
			FileHandle: fh,
			Offset:     offset,
			Size:       int64(len(data)),
			Data:       data,
		})
	if err != nil {
		glog.Errorf(`failed to write to file "%s" :: %s`, name, err.Error())
		return 0, fuse.EIO
	}
	return uint32(len(data)), fuse.OK
}

func (c *SamFsFileHandle) Flush() fuse.Status {
	glog.Infof("Flush called on %s", c.fileData.Name)
	return fuse.OK
}

func (c *SamFsFileHandle) Allocate(off uint64, size uint64,
	mode uint32) fuse.Status {

	glog.Info("Allocate called")
	return fuse.OK
}

func (c *SamFsFileHandle) Release() {
	glog.Infof("Release called on %s", c.fileData.Name)
	c.fileData.Lock()
	c.fileData.Refs--
	c.fileData.Unlock()
	c.closed = true
	return
}

func (c *SamFsFileHandle) Fsync(flags int) fuse.Status {
	glog.Infof("Fsync called %s", c.fileData.Name)
	return fuse.OK
}

func (c *SamFsFileHandle) GetAttr(out *fuse.Attr) fuse.Status {
	glog.Infof("GetAttr(file) called %s", c.fileData.Name)

	name := c.fileData.Name
	fh := c.fileData.serverFh
	resp, err := c.fileData.Fs.nfsClient.GetAttr(context.Background(),
		&pb.FileHandleRequest{
			FileHandle: fh,
		})
	if err != nil {
		glog.Errorf(`failed to get attributes of file "%s" :: %s`, name, err.Error())
		return fuse.EIO
	}

	fAttr := ProtoToFuseAttr(resp)
	// TODO(mihir): keep the owner as the user who started the client process, not
	// the owner of the process who is calling the command which in turn calls this
	// function
	// fAttr.Owner = fContext.Owner

	out = fAttr

	return fuse.OK
}

func (c *SamFsFileHandle) InnerFile() nodefs.File {
	glog.Info("InnerFile called")
	return c
}

func (c *SamFsFileHandle) SetInode(i *nodefs.Inode) {
	glog.Infof("SetInode called %s", c.fileData.Name)
}

func (c *SamFsFileHandle) Truncate(size uint64) fuse.Status {
	glog.Info("Truncate called")
	return fuse.ENOENT
}

func (c *SamFsFileHandle) Utimens(atime *time.Time,
	mtime *time.Time) fuse.Status {

	glog.Infof("Utimens(file) called %s", c.fileData.Name)
	return fuse.OK
}
