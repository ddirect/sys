package sys

import (
	"io/fs"

	"golang.org/x/sys/unix"
)

type FileKey struct {
	Device uint64
	Inode  uint64
}

type FileInfo struct {
	FileKey
	Path      string
	Size      int64
	ModTimeNs int64
	Mode      fs.FileMode
}

func Stat(path string) (FileInfo, error) {
	var stat unix.Stat_t
	return statRes(path, &stat, unix.Stat(path, &stat))
}

func Lstat(path string) (FileInfo, error) {
	var stat unix.Stat_t
	return statRes(path, &stat, unix.Lstat(path, &stat))
}

func statRes(path string, stat *unix.Stat_t, err error) (FileInfo, error) {
	if err != nil {
		return FileInfo{}, unixError("stat", path, err)
	}
	return FileInfo{FileKey{stat.Dev, stat.Ino}, path, stat.Size, stat.Mtim.Nano(), fileMode(stat.Mode)}, nil
}

// from fillFileStatFromSys
func fileMode(x uint32) fs.FileMode {
	res := fs.FileMode(x & 0777)
	switch x & unix.S_IFMT {
	case unix.S_IFBLK:
		res |= fs.ModeDevice
	case unix.S_IFCHR:
		res |= fs.ModeDevice | fs.ModeCharDevice
	case unix.S_IFDIR:
		res |= fs.ModeDir
	case unix.S_IFIFO:
		res |= fs.ModeNamedPipe
	case unix.S_IFLNK:
		res |= fs.ModeSymlink
	case unix.S_IFREG:
		// nothing to do
	case unix.S_IFSOCK:
		res |= fs.ModeSocket
	}
	if x&unix.S_ISGID != 0 {
		res |= fs.ModeSetgid
	}
	if x&unix.S_ISUID != 0 {
		res |= fs.ModeSetuid
	}
	if x&unix.S_ISVTX != 0 {
		res |= fs.ModeSticky
	}
	return res
}
