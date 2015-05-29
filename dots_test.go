package main

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	initialize()

	feedDir("copy", 1)("some data")
	feedDir("link", 2)("some data")
	feedDir("init", 3)("some data")

	var dots Dotfiles

	dots.read()

	if len(dots.Files) != 3 {
		t.Errorf("Dotfiles should have 3 dirs but found %s", len(dots.Files))
	}

	if len(dots.Files[cp]) != 1 {
		t.Errorf("copy dir should contains 1 files but found %s", len(dots.Files[cp]))
	}

	if len(dots.Files[ln]) != 2 {
		t.Errorf("link dir should contains 2 files but found %s", len(dots.Files[ln]))
	}

	cleanup()
}

func TestBackup(t *testing.T) {
	cleanup()
	initialize()

	feedDir("link", 6)("some data")

	content := "old content"
	err := ioutil.WriteFile(filepath.Join(RootDir, "file4"), []byte(content), 0777)
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile(filepath.Join(RootDir, "file5"), []byte(content), 0777)
	if err != nil {
		t.Error(err)
	}

	var dots Dotfiles

	dots.read()

	dots.backup(ln, link)

	files, err := ioutil.ReadDir(filepath.Join(BaseDir, "backup"))
	if err != nil {
		t.Error(err)
	}
	if len(files) != 2 {
		t.Errorf("2 files should have been backed up, but found %v", len(files))
	}

	cleanup()
}

func TestCp(t *testing.T) {
	initialize()

	feedDir("copy", 5)("data")

	var dots Dotfiles

	dots.read()

	dots.cp()

	// Check if all the copied files are present
	for i := 0; i < 5; i++ {
		isPresent(t, RootDir, mockFileName(i))
	}

	cleanup()
	invalideCache()
}

func TestLn(t *testing.T) {
	initialize()

	feedLink := feedDir("link", 5)
	feedLink("data")

	var dots Dotfiles

	dots.read()

	dots.ln()

	checkRoot := checkDir("../", 5)

	err := checkRoot("data")
	if err != nil {
		t.Errorf("Failed to link the files in %s\n%v", RootDir, err)
	}

	feedLink("new data")

	err = checkRoot("new data")
	if err != nil {
		t.Errorf("Linked files should have been updated\n%v", err)
	}

	cleanup()
	invalideCache()
}

func TestRn(t *testing.T) {
	initialize()

	feedDir("init", 3)("echo foo")

	expected := `foo
foo
foo
`

	var dots Dotfiles

	dots.read()

	out := dots.init()

	if bytes.Compare(out, []byte(expected)) != 0 {
		t.Errorf("%s\n was expected but found\n%s", expected, string(out))
	}
}
