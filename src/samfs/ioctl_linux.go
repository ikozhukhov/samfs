// +build linux

package samfs

//#include <stdint.h>
//#include <sys/ioctl.h>
//#include <sys/fcntl.h>
//#include <linux/fs.h>
//#include <stdio.h>
//#include <errno.h>
//
//uint32_t getGenerationNumber (char *f) {
//    int fileno = open(f, O_RDONLY);
//    uint32_t generation = -1;
//    if (ioctl(fileno, FS_IOC_GETVERSION, &generation)) {
//        return generation;
//    }
//    return generation;
//}
import "C"
import (
	"github.com/hanwen/go-fuse/fuse"
	pb "github.com/smihir/samfs/src/proto"
	"syscall"
)

func getGenerationNumber(filePath string) (uint32, error) {
	cg, err := C.getGenerationNumber(C.CString(filePath))
	return uint32(cg), err
}

func GetInodeAndGenerationNumbers(filePath string) (uint64, uint32, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(filePath, &stat); err != nil {
		return 0, 0, err
	}

	genNumber, err := getGenerationNumber(filePath)
	if err != nil {
		return 0, 0, err
	}

	return stat.Ino, genNumber, nil
}

// TODO(mihir): not a good place to keep these functions.
// they need a new file.

func StatToProtoAttr(stat *syscall.Stat_t) *pb.GetAttrReply {
	return &pb.GetAttrReply{
		Ino:       uint64(stat.Ino),
		Size:      uint64(stat.Size),
		Blocks:    uint64(stat.Blocks),
		Atime:     uint64(stat.Atim.Sec),
		Mtime:     uint64(stat.Mtim.Sec),
		Ctime:     uint64(stat.Ctim.Sec),
		Atimensec: uint32(stat.Atim.Nsec),
		Mtimensec: uint32(stat.Mtim.Nsec),
		Ctimensec: uint32(stat.Ctim.Nsec),
		Mode:      uint32(stat.Mode),
		Nlink:     uint32(stat.Nlink),
		Rdev:      uint32(stat.Rdev),
		Blksize:   uint32(stat.Blksize),
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
		Blksize:   protoAttr.Blksize,
	}
}
