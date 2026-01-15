//go:build windows

package aurora

import (
	"syscall"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceExW  = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// getDiskAvailable returns the available disk space in bytes for the given path
func getDiskAvailable(path string) (uint64, error) {
	var freeBytesAvailable uint64

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	ret, _, err := getDiskFreeSpaceExW.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(0),
		uintptr(0),
	)

	if ret == 0 {
		return 0, err
	}

	return freeBytesAvailable, nil
}
