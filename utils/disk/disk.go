package disk

import "syscall"

type DiskStatus struct {
	Size uint64
	Used uint64
	Free uint64
}

func DiskUsage(path string) (disk *DiskStatus, err error) {
	fs := new(syscall.Statfs_t)
	err = syscall.Statfs(path, fs)
	if err != nil {
		return
	}

	disk = new(DiskStatus)
	disk.Size = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.Size - disk.Free
	return
}
