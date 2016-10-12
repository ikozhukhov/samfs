package samfs

import (
	//"bytes"
	"sync"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type SamFsFileHandle struct {
	at       int64
	closed   bool
	fileData *SamFsFileData
}

type SamFsFileData struct {
	sync.Mutex
	Fs    *SamFs
	Refs  int32
	Name  string
	Data  []byte
	Hash  []byte
	Dirty bool
	Attr  *fuse.Attr
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

func NewEmptyFileData(path string) *SamFsFileData {
	return &SamFsFileData{
		Refs:  0,
		Dirty: true,
	}
}

func NewFileData(path string, fs *SamFs, hash []byte, data []byte,
	attr *fuse.Attr) *SamFsFileData {

	return &SamFsFileData{
		Refs:  0,
		Fs:    fs,
		Name:  path,
		Data:  data,
		Hash:  hash,
		Dirty: false,
		Attr:  attr,
	}
}

func (c *SamFsFileHandle) String() string {
	return c.fileData.Name
}

func (c *SamFsFileHandle) Chmod(mode uint32) fuse.Status {
	return fuse.OK
}

func (c *SamFsFileHandle) Chown(uid uint32, gid uint32) fuse.Status {
	return fuse.OK
}

func (c *SamFsFileHandle) Read(buf []byte, off int64) (fuse.ReadResult,
	fuse.Status) {

	return nil, fuse.OK
}

func (c *SamFsFileHandle) Write(data []byte, offset int64) (uint32,
	fuse.Status) {

	return uint32(len(data)), fuse.OK
}

func (c *SamFsFileHandle) Flush() fuse.Status {
	return fuse.OK
}

func (c *SamFsFileHandle) Allocate(off uint64, size uint64,
	mode uint32) fuse.Status {

	return fuse.OK
}

func (c *SamFsFileHandle) Release() {
	c.fileData.Lock()
	c.fileData.Refs--
	c.fileData.Unlock()
	c.closed = true
	return
}

func (c *SamFsFileHandle) Fsync(flags int) fuse.Status {
	return fuse.OK
}

func (c *SamFsFileHandle) GetAttr(out *fuse.Attr) fuse.Status {
	return fuse.OK
}

func (c *SamFsFileHandle) InnerFile() nodefs.File {
	return c
}

func (c *SamFsFileHandle) SetInode(i *nodefs.Inode) {
}

func (c *SamFsFileHandle) Truncate(size uint64) fuse.Status {
	return fuse.ENOENT
}

func (c *SamFsFileHandle) Utimens(atime *time.Time,
	mtime *time.Time) fuse.Status {

	return fuse.OK
}