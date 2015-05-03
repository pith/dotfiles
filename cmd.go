// Copyright (c) 2015 by Pierre Thirouin. All rights reserved.

// This file is part of dotfiles, a simple dotfiles manager.

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var (
	debugMode = false
	quietMode = false
)

var help = `usage: dotfiles [command]

The commands are:
  create    Create a new .dotfiles directory
  init      Initialize your config difine in the .dotfiles dir
`

var (
	// RootDir is the directory where files will be link or copy.
	RootDir = "."

	// DotFilesDir is the name of the directory where the dotfiles are stored.
	DotFilesDir = ".dotfiles"

	// BaseDir is the path to the dotfiles directory
	BaseDir = filepath.Join(RootDir, DotFilesDir)

	dirs = [8]string{"bin", "conf", "copy", "init", "link", "source", "test", "vendor"}
)

func changeRootDir(path string) {
	RootDir = path
	BaseDir = filepath.Join(RootDir, DotFilesDir)
}

func main() {
	flag.Parse()

	command := flag.Arg(0)

	usr, err := user.Current()
	if err != nil {
		fmt.Errorf("%v", err)
	}

	changeRootDir(usr.HomeDir)

	if command == "create" {
		initialize()
	} else if command == "init" {
		printHeader("Copying files into home directory")
		copyDir()

		printHeader("Linking files into home directory")
		linkDir()

		printHeader("Run the following init scripts")
		initDir()

		sourceDir()

		printHeader("All done !")
	} else {
		fmt.Print(help)
	}
}

func printHeader(s string) {
	if !quietMode {
		fmt.Printf("\033[1m%s\033[0m\n", s)
	}
}

func printArrow(s string) {
	if !quietMode {
		fmt.Printf(" \033[1;34m➜\033[0m  %s\n", s)
	}
}

// Initialize creates a new .dotfiles repo
func initialize() {
	printHeader("Scaffold " + BaseDir)

	for _, dir := range dirs {
		printArrow(dir)
		os.MkdirAll(filepath.Join(BaseDir, dir), 0777)
	}
	if !quietMode {
		fmt.Println("")
	}
	printHeader("All done !")
}

// CloneRepo clones the given git repository
func cloneRepo(gitrepo string) {
	printHeader("Clone " + gitrepo)
	git, err := exec.LookPath("git")
	if err != nil {
		fmt.Errorf("git is required to clone the dotfiles repo")
	}

	err = exec.Command(git, "clone", gitrepo, BaseDir).Run()
	if err != nil {
		fmt.Errorf("Failed to clone %s", gitrepo)
	}

	printHeader(BaseDir + " is ready !")
}

func backup(file string) {
	// If there is no backup dir yet create it
	if _, err := os.Stat(filepath.Join(BaseDir, "backup")); os.IsNotExist(err) {
		os.Mkdir(filepath.Join(BaseDir, "backup"), 0777)
	}

	if _, err := os.Stat(filepath.Join(RootDir, file)); !os.IsNotExist(err) {
		// The file already exists so backup it
		printHeader("Backup %s")

		err = exec.Command("mv", filepath.Join(RootDir, file), filepath.Join(BaseDir, "backup", file)).Run()
		if err != nil {
			fmt.Errorf("Failed to backup %s\n%v", file, err)
		}
	}
}

// CopyDir copies all the files in the copy dir at ~/
func copyDir() {
	applyCmd("copy", func(fileToCopy string) error {
		return exec.Command("cp", fileToCopy, RootDir).Run()
	})
}

// LinkDir links all the files in the link dir at ~/
func linkDir() {
	applyCmd("link", func(file string) error {
		return exec.Command("ln", "-sf", file, RootDir).Run()
	})
}

// InitDir executes all the scripts in the init dir
func initDir() []byte {
	var out []byte

	applyCmd("init", func(initFile string) error {
		output, err := exec.Command("/bin/bash", initFile).CombinedOutput()
		out = append(out, output...)
		return err
	})

	if !quietMode && len(out) != 0 {
		fmt.Printf("%s", string(out))
	}
	return out
}

// SourceDir source all the files in the source dir
// Solution from here: http://stackoverflow.com/a/29995987/1292605
func sourceDir() {
	applyCmd("source", func(file string) error {
		printHeader("Sourcing " + file)

		cmd := exec.Command("/bin/bash", "-c", "source "+file+" ; echo '<<<ENVIRONMENT>>>' ; env")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}

		s := bufio.NewScanner(bytes.NewReader(out))
		start := false
		for s.Scan() {
			if s.Text() == "<<<ENVIRONMENT>>>" {
				start = true
			} else if start {
				kv := strings.SplitN(s.Text(), "=", 2)
				if len(kv) == 2 {
					os.Setenv(kv[0], kv[1])
				}
			}
		}

		return nil
	})
}

func applyCmd(dir string, execCmd func(string) error) {
	dirPath := filepath.Join(BaseDir, dir)

	files, readErr := ioutil.ReadDir(dirPath)
	if readErr != nil {
		fmt.Errorf("Failed to read %s dir at %s", dir, dirPath)
	}

	for _, file := range files {
		err := execCmd(filepath.Join(dirPath, file.Name()))
		if err != nil {
			fmt.Errorf("Failed to %s file: %s", dir, file.Name())
		}
	}
}
