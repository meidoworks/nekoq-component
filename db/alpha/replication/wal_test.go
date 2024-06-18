package replication

import (
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestInitFromExistingWalFile(t *testing.T) {
	memfs := afero.NewMemMapFs()
	_ = memfs.Mkdir("wal_data", 0755)
	w := NewWal(WalOptions{
		FS: afero.NewBasePathFs(memfs, "wal_data"),
	})
	if err := w.Startup(); err != nil {
		t.Fatal(err)
	}

	// write something
	if pos, err := w.Write([]byte("hello world!")); err != nil {
		t.Fatal(err)
	} else {
		if pos[0] != 1 || pos[1] != 1 {
			t.Fatal("position error:", pos)
		}
	}

	if err := w.Shutdown(); err != nil {
		t.Fatal(err)
	}

	// reopen existing wal files
	w = NewWal(WalOptions{
		FS: afero.NewBasePathFs(memfs, "wal_data"),
	})
	if err := w.Startup(); err != nil {
		t.Fatal(err)
	}
	pos, err := w.CurrentPosition()
	if err != nil {
		t.Fatal(err)
	}
	if pos[0] != 1 || pos[1] != 1 {
		t.Fatal("position error:", pos)
	}

	// write something
	if pos, err := w.Write([]byte("hello world!")); err != nil {
		t.Fatal(err)
	} else {
		if pos[0] != 1 || pos[1] != 2 {
			t.Fatal("position error:", pos)
		}
	}

	// check file content
	dataFs := afero.NewBasePathFs(memfs, "wal_data")
	f, err := dataFs.Open("wal_0000000000000001.log")
	if err != nil {
		t.Fatal(err)
	}
	all, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	expect := []byte{
		// record 1
		0, 0, 0, 0, 0, 0, 0, 1, // pos[0]
		0, 0, 0, 0, 0, 0, 0, 1, // pos[1]
		0, 0, 0, 12, // size
		30, 137, 4, 126, // crc
		0, 0, 0, 0, 0, 0, 0, 0, // zero
		104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33, // data
		// record 2
		0, 0, 0, 0, 0, 0, 0, 1, // pos[0]
		0, 0, 0, 0, 0, 0, 0, 2, // pos[1]
		0, 0, 0, 12, // size
		30, 137, 4, 126, // crc
		0, 0, 0, 0, 0, 0, 0, 0, // zero
		104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33, // data
	}
	if !slices.Equal(all, expect) {
		t.Fatal("file data unexpected")
	}
}

func TestInitWrite(t *testing.T) {
	memfs := afero.NewMemMapFs()
	_ = memfs.Mkdir("wal_data", 0755)
	w := NewWal(WalOptions{
		FS: afero.NewBasePathFs(memfs, "wal_data"),
	})
	if err := w.Startup(); err != nil {
		t.Fatal(err)
	}

	// check created file
	files := map[string]struct{}{}
	if err := afero.Walk(memfs, ".", func(path string, info fs.FileInfo, err error) error {
		files[path] = struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	foundFolder := false
	foundFile := false
	for k := range files {
		if k == "wal_data" {
			foundFolder = true
		}
		if strings.HasSuffix(k, "wal_0000000000000001.log") {
			foundFile = true
		}
	}
	if !foundFile || !foundFolder {
		t.Fatal(fmt.Sprint("folder or file not created:", foundFolder, foundFile))
	}

	// write something
	if pos, err := w.Write([]byte("hello world!")); err != nil {
		t.Fatal(err)
	} else {
		if pos[0] != 1 || pos[1] != 1 {
			t.Fatal("position error:", pos)
		}
	}

	// check file content
	dataFs := afero.NewBasePathFs(memfs, "wal_data")
	f, err := dataFs.Open("wal_0000000000000001.log")
	if err != nil {
		t.Fatal(err)
	}
	all, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	expect := []byte{
		0, 0, 0, 0, 0, 0, 0, 1, // pos[0]
		0, 0, 0, 0, 0, 0, 0, 1, // pos[1]
		0, 0, 0, 12, // size
		30, 137, 4, 126, // crc
		0, 0, 0, 0, 0, 0, 0, 0, // zero
		104, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33, // data
	}
	if !slices.Equal(all, expect) {
		t.Fatal("file data unexpected")
	}
}
