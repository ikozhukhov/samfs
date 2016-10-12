package samfs

//#include <sys/ioctl.h>
//#include <sys/fcntl.h>
//#include <linux/fs.h>
//#include <stdio.h>
//#include <errno.h>
//
//int getGenerationNumber (char *f) {
//    int fileno = open(f, O_RDONLY);
//    int generation = -1;
//    if (ioctl(fileno, FS_IOC_GETVERSION, &generation)) {
//        return generation;
//    }
//    return generation;
//}
import (
	"C"
)

func getGenerationNumber(filePath string) (int, error) {
	cg, err := C.getGenerationNumber(C.CString(filePath))
	return int(cg), err
}
