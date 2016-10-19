package samfs

import (
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

type CacheEntry struct {
	Data            *[]byte
	Offset          int64
	ServerSessionID int64
}

type Cache struct {
	entries    []*CacheEntry
	numEntries uint32
}

type SamFsFileData struct {
	sync.Mutex
	serverFh *pb.FileHandle
	Fs       *SamFs
	Refs     int32
	Name     string
	DCache   Cache
	Dirty    bool
	Attr     *fuse.Attr
}

var _ nodefs.File = &SamFsFileHandle{}

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

func (cc *Cache) AddEntry(e *CacheEntry) {
	cc.entries = append(cc.entries, e)
	cc.numEntries += 1
}

func (cc *Cache) Invalidate() {
	if cc.numEntries == 0 {
		return
	}
	// copy is required so that gc can remove the older entries
	// otherwise our array will keep on growing
	cc.entries = []*CacheEntry{}
	cc.numEntries = 0
}

func (c *SamFsFileHandle) String() string {
	glog.V(3).Info("String called")
	return c.fileData.Name
}

func (c *SamFsFileHandle) Chmod(mode uint32) fuse.Status {
	glog.V(3).Info("Chmod called")
	return fuse.OK
}

func (c *SamFsFileHandle) Chown(uid uint32, gid uint32) fuse.Status {
	glog.V(3).Info("Chown called")
	return fuse.OK
}

func (c *SamFsFileHandle) Read(buf []byte, off int64) (fuse.ReadResult,
	fuse.Status) {

	glog.V(3).Infof("Read called on %s off: %d, size %d", c.fileData.Name, off, len(buf))
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

	glog.V(3).Infof("Write called on %s", c.fileData.Name)

	resp, err := c.write(data, offset)

	if err != nil {
		glog.Errorf(`failed to write to file "%s" :: %s`, c.fileData.Name,
			err.Error())
		return 0, fuse.EIO
	}
	if c.fileData.DCache.numEntries != 0 &&
		c.fileData.DCache.entries[c.fileData.DCache.numEntries-1].ServerSessionID !=
			resp.ServerSessionID {
		glog.Warning("server state change detected")
		c.fileData.Dirty = true
	}
	c.fileData.Lock()
	c.fileData.DCache.AddEntry(&CacheEntry{
		Data:            &data,
		Offset:          offset,
		ServerSessionID: resp.ServerSessionID,
	})
	c.fileData.Unlock()
	return uint32(len(data)), fuse.OK
}

func (c *SamFsFileHandle) Flush() fuse.Status {
	glog.V(3).Infof("Flush called on %s", c.fileData.Name)
	if c.fileData.DCache.numEntries == 0 {
		return fuse.OK
	} else {
		return c.Fsync(0)
	}
}

func (c *SamFsFileHandle) Allocate(off uint64, size uint64,
	mode uint32) fuse.Status {

	glog.V(3).Info("Allocate called")
	return fuse.OK
}

func (c *SamFsFileHandle) Release() {
	glog.V(3).Infof("Release called on %s", c.fileData.Name)
	if c.fileData.DCache.numEntries != 0 {
		_ = c.Fsync(0)
	}
	c.fileData.Lock()
	c.fileData.Refs--
	c.fileData.Unlock()
	c.closed = true
	return
}

func (c *SamFsFileHandle) Fsync(flags int) fuse.Status {
	glog.V(3).Infof("Fsync called %s", c.fileData.Name)
	if c.fileData.Dirty == true {
		glog.Errorf("Fsync called on %s with dirty cache", c.fileData.Name)
	}

	fh := c.fileData.serverFh
	var resp *pb.StatusReply
	var err error
	if c.fileData.Dirty != true {
		resp, err = c.fileData.Fs.nfsClient.Commit(context.Background(),
			&pb.CommitRequest{
				FileHandle: fh,
			})
		if err != nil {
			glog.Errorf(`failed to commit to file "%s" :: %s`, c.fileData.Name,
				err.Error())
			return fuse.EIO
		}
	}

	// crash detected between writes or between fsync and writes
	if (c.fileData.DCache.numEntries != 0 &&
		c.fileData.DCache.entries[c.fileData.DCache.numEntries-1].ServerSessionID !=
			resp.ServerSessionID) || (c.fileData.Dirty == true) {
		glog.Warning("server state change detected during fsync, replay all writes")

		for _, de := range c.fileData.DCache.entries {
			resp, err := c.write(*de.Data, de.Offset)
			if err != nil {
				glog.Errorf(`failed to write during recovery to file "%s" :: %s`,
					c.fileData.Name, err.Error())
				return fuse.EIO
			}

			de.ServerSessionID = resp.ServerSessionID
		}
		resp, err := c.fileData.Fs.nfsClient.Commit(context.Background(),
			&pb.CommitRequest{
				FileHandle: fh,
			})
		if err != nil {
			glog.Errorf(`failed to recursively commit to file "%s" :: %s`,
				c.fileData.Name, err.Error())
			return fuse.EIO
		}
		for _, de := range c.fileData.DCache.entries {
			if de.ServerSessionID != resp.ServerSessionID {
				glog.Error("recursive crash during fsync")
				return fuse.EIO
			}
		}
	}

	c.fileData.Lock()
	c.fileData.DCache.Invalidate()
	c.fileData.Unlock()

	return fuse.OK
}

func (c *SamFsFileHandle) GetAttr(out *fuse.Attr) fuse.Status {
	glog.V(3).Infof("GetAttr(file) called %s", c.fileData.Name)

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

	fAttr.Owner = c.fileData.Fs.owner

	// Have to copy these one by one as expected by the library
	out.Ino = fAttr.Ino
	out.Size = fAttr.Size
	out.Blocks = fAttr.Blocks
	out.Atime = fAttr.Atime
	out.Mtime = fAttr.Mtime
	out.Ctime = fAttr.Ctime
	out.Atimensec = fAttr.Atimensec
	out.Mtimensec = fAttr.Mtimensec
	out.Ctimensec = fAttr.Ctimensec
	out.Mode = fAttr.Mode
	out.Nlink = fAttr.Nlink
	out.Owner.Uid = fAttr.Owner.Uid
	out.Owner.Gid = fAttr.Owner.Gid
	out.Rdev = fAttr.Rdev
	//out.Blksize = fAttr.Blksize
	//out.Padding = fAttr.Padding

	return fuse.OK
}

func (c *SamFsFileHandle) InnerFile() nodefs.File {
	glog.V(3).Info("InnerFile called")
	return c
}

func (c *SamFsFileHandle) SetInode(i *nodefs.Inode) {
	glog.V(3).Infof("SetInode called %s", c.fileData.Name)
}

func (c *SamFsFileHandle) Truncate(size uint64) fuse.Status {
	glog.V(3).Info("Truncate called")
	return fuse.OK
}

func (c *SamFsFileHandle) Utimens(atime *time.Time,
	mtime *time.Time) fuse.Status {

	glog.V(3).Infof("Utimens(file) called %s", c.fileData.Name)
	return fuse.OK
}

func (c *SamFsFileHandle) write(data []byte, offset int64) (*pb.StatusReply,
	error) {

	glog.V(3).Infof("Write called on %s", c.fileData.Name)
	fh := c.fileData.serverFh

	c.fileData.Lock()
	resp, err := c.fileData.Fs.nfsClient.Write(context.Background(),
		&pb.WriteRequest{
			FileHandle: fh,
			Offset:     offset,
			Size:       int64(len(data)),
			Data:       data,
		})
	c.fileData.Unlock()

	return resp, err
}
