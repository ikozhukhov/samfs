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
import (
	"C"
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
