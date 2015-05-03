// Copyright (c) 2015 by Pierre Thirouin. All rights reserved.

// This file is part of dotfiles, a simple dotfiles manager.

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// Setup a temporary dir where the tests are run and cleanup at the end
func TestMain(m *testing.M) {
	quietMode = true

	tmpDir, err := ioutil.TempDir("", "go-test")
	if err != nil {
		fmt.Errorf("Failed to create tmp dir: %v", err)
	}

	RootDir = filepath.Join(tmpDir, "root")
	os.Mkdir(RootDir, 0777)
	changeRootDir(RootDir)

	if debugMode {
		fmt.Printf("BaseDir: %s\n", BaseDir)
	}
	res := m.Run()

	err = os.RemoveAll(RootDir)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(res)
}

func TestInitDotfilesDir(t *testing.T) {
	initialize()

	expectedDirs := []string{"bin", "conf", "copy", "init", "link", "source", "test", "vendor"}

	// Check if all the expected directories are present
	for _, dirName := range expectedDirs {
		isPresent(t, BaseDir, dirName)
	}

	// Cleanup for TestCloneDir
	err := os.RemoveAll(RootDir)
	if err != nil {
		log.Fatal(err)
	}
}

func TestCloneDir(t *testing.T) {

	cloneRepo("_dotfilesSample")

	if _, err := os.Stat(BaseDir); os.IsNotExist(err) {
		t.Errorf("no such file or directory: %s", BaseDir)
		return
	}
}

func TestBackupFile(t *testing.T) {
	file := "fileToBackup"

	err := ioutil.WriteFile(filepath.Join(RootDir, file), []byte("some old config"), 0777)
	if err != nil {
		log.Fatal(err)
	}

	backup(file)

	if _, err := os.Stat(filepath.Join(BaseDir, "backup", file)); os.IsNotExist(err) {
		t.Errorf("Failed to backup %s", filepath.Join(RootDir, file))
	}
}

func TestCopy(t *testing.T) {

	initialize()

	feedDir("copy", 5)("data")

	copyDir()

	// Check if all the copied files are present
	for i := 0; i < 5; i++ {
		isPresent(t, RootDir, mockFileName(i))
	}
}

func TestLink(t *testing.T) {

	initialize()

	feedLink := feedDir("link", 5)
	feedLink("data")

	err := checkDir("link", 5)("data")
	if err != nil {
		t.Errorf("Failed to init the copy dir, %v", err)
	}

	linkDir()

	checkRoot := checkDir("../", 5)

	err = checkRoot("data")
	if err != nil {
		t.Errorf("Failed to link the files in %s\n%v", RootDir, err)
	}

	feedLink("new data")

	err = checkRoot("new data")
	if err != nil {
		t.Errorf("Linked files should have been updated\n%v", err)
	}

}

func TestInit(t *testing.T) {
	initialize()

	feedDir("init", 3)("echo foo")

	expected := `foo
foo
foo
`

	out := initDir()
	if bytes.Compare(out, []byte(expected)) != 0 {
		t.Errorf("%s\n was expected but found\n%s", string(out), expected)
	}
}

func TestSource(t *testing.T) {
	initialize()

	feedDir("source", 1)("echo plop; export FOO=\"youpi !!\"; env")

	sourceDir()

	if os.Getenv("FOO") != "youpi !!" {
		t.Errorf("Failed to source files. Expected \"FOO=youpi !!\", but found \"%s\"", os.Getenv("FOO"))
	}
}

// ====== Utils ======

func mockFileName(i int) string {
	return "file" + strconv.Itoa(i)
}

func feedDir(base string, count int) func(string) {
	return func(content string) {
		for i := 0; i < count; i++ {
			path := filepath.Join(BaseDir, base, mockFileName(i))
			err := ioutil.WriteFile(path, []byte(content), 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func checkDir(base string, count int) func(string) error {
	return func(content string) error {
		for i := 0; i < count; i++ {
			path := filepath.Join(BaseDir, base, mockFileName(i))
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			if string(bytes) != content {
				return errors.New("expected \"" + string(bytes) + "\" but found \"" + content + "\" in " + path)
			}
		}
		return nil
	}
}

func isPresent(t *testing.T, base, name string) {
	filename := filepath.Join(base, name)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("no such file or directory: %s", filename)
		return
	}
}
