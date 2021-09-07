package sys

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/unix"
)

type DirEntry struct {
	Name  string
	Inode uint64
	Type  fs.FileMode
}

// dirRelPath is the path to the directory containing the entry, relative to root
type WalkDirFunc func(dirRelPath string, d DirEntry, err error) error

func WalkDir(root string, cb WalkDirFunc) error {
	var core func(string, int, string) error
	core = func(dirRelPath string, dirFd int, dirName string) error {
		sourFd, err := unix.Openat(dirFd, dirName, unix.O_RDONLY, 0)
		if err != nil {
			return unixError("open", filepath.Join(root, dirRelPath, dirName), err)
		}
		defer unix.Close(sourFd)

		// this stays on the stack; a byte slice results in the same code
		// as long as it's smaller than a predefined number of bytes; otherwise a slice would
		// be allocated on the heap, while an array would remain on the stack
		var buf [0x2000]byte

		for {
			wp, err := unix.Getdents(sourFd, buf[:])
			if err != nil {
				return unixError("getdents", filepath.Join(root, dirRelPath, dirName), err)
			}
			if wp == 0 {
				return nil
			}
			for rp := 0; rp < wp; {
				dent := (*unix.Dirent)(unsafe.Pointer(&buf[rp]))
				cut := buf[rp+int(unsafe.Offsetof(dent.Name)):]
				rp += int(dent.Reclen)
				// the names are zero terminated; some garbage may be present between the zero
				// terminated name and the end of the structure
				cut = cut[:bytes.IndexByte(cut, 0)]

				// remove unwanted entries before converting to string; this does not
				// allocate anything and tests inline (borrowed from unix.ParseDirent)
				if string(cut) == "." || string(cut) == ".." {
					continue
				}

				// copy the string here: casting the slice directly (as done in strings.Builder.String())
				// may create problems if the visitor is concurrent; it also decreases performance
				// because it forces the buffer (buf) to be allocated on the heap instead of being
				// kept on the stack
				name := string(cut)
				typ := direntType(dent.Type)
				if err = cb(dirRelPath, DirEntry{name, dent.Ino, typ}, nil); err != nil {
					return err
				}

				if typ.IsDir() {
					if err := core(filepath.Join(dirRelPath, name), sourFd, name); err != nil {
						return err
					}
				}
			}
		}
	}
	return core("", unix.AT_FDCWD, root)
}

// from dirent_linux.go
func direntType(typ uint8) fs.FileMode {
	switch typ {
	case unix.DT_BLK:
		return fs.ModeDevice
	case unix.DT_CHR:
		return fs.ModeDevice | fs.ModeCharDevice
	case unix.DT_DIR:
		return fs.ModeDir
	case unix.DT_FIFO:
		return fs.ModeNamedPipe
	case unix.DT_LNK:
		return fs.ModeSymlink
	case unix.DT_REG:
		return 0
	case unix.DT_SOCK:
		return fs.ModeSocket
	default:
		panic("unknown dirent type")
	}
}
