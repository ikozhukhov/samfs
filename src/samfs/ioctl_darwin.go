// +build darwin

package samfs

import (
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
