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
	cachePath = filepath.Join(BaseDir, "cache", "cache.json")

	if debugMode {
		fmt.Printf("BaseDir: %s\n", BaseDir)
	}
	res := m.Run()

	cleanup()
	fmt.Println(BaseDir)
	os.Exit(res)
}

func cleanup() {
	err := os.RemoveAll(BaseDir)
	if err != nil {
		log.Fatal(err)
	}
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

	cleanup()
}

// Ignore test: Doesn't build on travis
func _TestCloneDir(t *testing.T) {

	cloneRepo("_dotfilesSample")

	if _, err := os.Stat(BaseDir); os.IsNotExist(err) {
		t.Errorf("no such file or directory: %s", BaseDir)
		return
	}

	cleanup()
}

func TestBackgroundCheck(t *testing.T) {
	initialize()

	feedDir("copy", 1)("data")

	if !backgroundCheck(filepath.Join(BaseDir, "copy", "file0")) {
		t.Errorf("Background check should be ok")
	}

	copyDir()

	if backgroundCheck(filepath.Join(BaseDir, "copy", "file0")) {
		t.Errorf("Background check should be ko")
	}

	feedDir("copy", 1)("data\n newdata")

	if !backgroundCheck(filepath.Join(BaseDir, "copy", "file0")) {
		t.Errorf("Background check should be ok")
	}

	cleanup()
}

func TestBackupFile(t *testing.T) {
	initialize()

	file := "fileToBackup"

	err := ioutil.WriteFile(filepath.Join(RootDir, file), []byte("some old config"), 0777)
	if err != nil {
		fmt.Print("coucou")
		log.Fatal(err)
	}

	backupIfExist(file)

	if _, err := os.Stat(filepath.Join(BaseDir, "backup", file)); os.IsNotExist(err) {
		t.Errorf("Failed to backup %s", filepath.Join(RootDir, file))
	}

	cleanup()
}

func TestCopy(t *testing.T) {
	initialize()

	feedDir("copy", 5)("data")

	copyDir()

	// Check if all the copied files are present
	for i := 0; i < 5; i++ {
		isPresent(t, RootDir, mockFileName(i))
	}

	cleanup()
}

func TestLink(t *testing.T) {
	initialize()

	feedLink := feedDir("link", 5)
	feedLink("data")

	linkDir()

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
}

func TestFirstInit(t *testing.T) {
	initialize()

	if !firstInit() {
		t.Errorf("It should be the first init")
	}

	loadCache()

	if firstInit() {
		t.Errorf("It should not be the first init")
	}

	cleanup()
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
		t.Errorf("%s\n was expected but found\n%s", expected, string(out))
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
				return errors.New("expected \"" + content + "\" but found \"" + string(bytes) + "\" in " + path)
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
