// +build darwin

package samfs

import (
	"github.com/hanwen/go-fuse/fuse"
	pb "github.com/smihir/samfs/src/proto"
	"syscall"
)

func GetInodeAndGenerationNumbers(filePath string) (uint64, uint32, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(filePath, &stat); err != nil {
		return 0, 0, err
	}
	// for non-root uses generation number will be 0 on osx
	return stat.Ino, stat.Gen, nil
}

// TODO(mihir): not a good place to keep these functions.
// they need a new file.

func StatToProtoAttr(stat *syscall.Stat_t) *pb.GetAttrReply {
	return &pb.GetAttrReply{
		Ino:     uint64(stat.Ino),
		Size:    uint64(stat.Size),
		Blocks:  uint64(stat.Blocks),
		Mode:    uint32(stat.Mode),
		Nlink:   uint32(stat.Nlink),
		Rdev:    uint32(stat.Rdev),
		Blksize: uint32(stat.Blksize),
	}
}

func ProtoToFuseAttr(protoAttr *pb.GetAttrReply) *fuse.Attr {
	return &fuse.Attr{
		Ino:       protoAttr.Ino,
		Size:      protoAttr.Size,
		Blocks:    protoAttr.Blocks,
		Atime:     protoAttr.Atime,
		Mtime:     protoAttr.Mtime,
		Ctime:     protoAttr.Ctime,
		Atimensec: protoAttr.Atimensec,
		Mtimensec: protoAttr.Mtimensec,
		Ctimensec: protoAttr.Ctimensec,
		Mode:      protoAttr.Mode,
		Nlink:     protoAttr.Nlink,
		Rdev:      protoAttr.Rdev,
	}
}
