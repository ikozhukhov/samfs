#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <unistd.h>
#include <string.h>
#include <errno.h>
#include <signal.h>
#include <fcntl.h>
#include <ctype.h>
#include <termios.h>
#include <sys/types.h>
#include <sys/mman.h>

#define PRINT_ERROR \
        do { \
                fprintf(stderr, "Error at line %d, file %s (%d) [%s]\n", \
                __LINE__, __FILE__, errno, strerror(errno)); exit(1); \
        } while(0)

int main(int argc, char **argv){
    void *map_base;
    int fd, newfd;
    int *device_array;
    char buff[10] = "HelloWorld";

    
    if ((fd = open(argv[1], O_CREAT | O_RDWR | O_SYNC, S_IRWXU)) == -1) PRINT_ERROR;
    printf("%s file opened.\n", argv[1]);
    //fflush(stdout);
    
    //write
    if ((write(fd, buff, sizeof(buff))) == -1) PRINT_ERROR;
    if ((read(fd, buff, sizeof(buff))) == -1) PRINT_ERROR;
    //dup
    newfd = dup(fd);

    //clode fd1
    //if (close(fd)) PRINT_ERROR;
    
    //write fd2
    if ((write(newfd, buff, sizeof(buff))) == -1) PRINT_ERROR;
    
    //close fd2
    if (close(newfd)) PRINT_ERROR;
    printf("%s file closed.\n", argv[1]);
    printf("success!\n");
    return 0;
}
