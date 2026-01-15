//go:build !windows

package aurora

import "syscall"

// getDiskAvailable returns the available disk space in bytes for the given path
func getDiskAvailable(path string) (uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	// Bavail = blocks available to unprivileged users
	return stat.Bavail * uint64(stat.Bsize), nil
}
