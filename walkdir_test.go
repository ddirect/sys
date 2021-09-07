package sys

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/ddirect/check"
	"github.com/ddirect/filetest"
	ft "github.com/ddirect/filetest"
)

func (e DirEntry) String() string {
	return fmt.Sprintf("%v %x %d %s", e.Type, uint(e.Type), e.Inode, e.Name)
}

func TestWalkDir(t *testing.T) {
	root, _, _ := filetest.CommitNewRandomTree(t, "walkdirtest", ft.TreeOptions{
		ft.ValidChars,
		ft.MinMax{240, 250},
		ft.MinMax{50, 100},
		ft.MinMax{2, 4},
		4,
	}, ft.Mixes{0, 50, 50, 80})

	t1 := make(map[string]DirEntry)
	check.E(filepath.WalkDir(root, func(path string, d fs.DirEntry, err1 error) error {
		check.E(err1)
		fi, err := d.Info()
		check.E(err)
		t1[path] = DirEntry{d.Name(), fi.Sys().(*syscall.Stat_t).Ino, d.Type()}
		return nil
	}))

	t2 := make(map[string]DirEntry)
	// simulate what filepath.WalkDir does:
	rootstat, err := Stat(root)
	check.E(err)
	e := DirEntry{filepath.Base(root), rootstat.Inode, rootstat.Mode.Type()}
	t2[root] = e
	check.E(WalkDir(root, func(path string, d DirEntry, err1 error) error {
		check.E(err1)
		t2[filepath.Join(root, path, d.Name)] = DirEntry{d.Name, d.Inode, d.Type}
		return nil
	}))

	if !reflect.DeepEqual(t1, t2) {
		t.Fail()
	}
}
