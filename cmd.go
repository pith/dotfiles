// Copyright (c) 2015 by Pierre Thirouin. All rights reserved.

// This file is part of dotfiles, a simple dotfiles manager.

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"os/user"
	"time"
	"bytes"
	"errors"
)

var (
	debugMode = false
	quietMode = false
)

var help =`usage: dotfiles [command]

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

func copyDir() {
	applyCmd("copy", func(fileToCopy string) *exec.Cmd {
		return exec.Command("cp", fileToCopy, RootDir)
	})
}

func linkDir() {
	applyCmd("link", func(fileToLink string) *exec.Cmd {
		return exec.Command("ln", "-sf", fileToLink, RootDir)
	})
}

func sourceDir() {
	applyCmd("source", func(fileToLink string) *exec.Cmd {
		printHeader("Sourcing " + fileToLink)
		return exec.Command("bash", "-c", "source", fileToLink, "echo $FOO")
	})
}

func initDir() []byte {
	out := applyCmd("init", func(initFile string) *exec.Cmd {
		return exec.Command("/bin/bash", initFile)
	})
	if len(out) != 0 {
		fmt.Printf("%s", string(out))
	}
	return out
}

func applyCmd(dir string, cmdFactory func(string) *exec.Cmd) []byte {
	dirPath := filepath.Join(BaseDir, dir)

	files, readErr := ioutil.ReadDir(dirPath)
	if readErr != nil {
		fmt.Errorf("Failed to read %s dir at %s", dir, dirPath)
	}

	output := []byte{}

	for _, file := range files {
		fileToCopy := filepath.Join(dirPath, file.Name())

		cmd := cmdFactory(fileToCopy)
		if debugMode {
			fmt.Printf("=> Executing: %s\n", strings.Join(cmd.Args, " "))
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Errorf("Failed to %s file: %s", dir, fileToCopy)
		}
		if debugMode {
			fmt.Printf("%s\n", string(out))
		}
		output = append(output, out...)

		//if testEnv {
			for _, envv := range cmd.Env {
				fmt.Println(envv)
			}
		//}
	}
	if testEnv {
		fmt.Println(string(output))
	}
	return output
}

// func sourceDir() {
// 	_, err := source()
// 	if err != nil {
// 		fmt.Printf("err: %v", err)
// 	}
// }

func source() ([]byte, error) {
	dir := "source"
	dirPath := filepath.Join(BaseDir, dir)

	files, readErr := ioutil.ReadDir(dirPath)
	if readErr != nil {
		fmt.Errorf("Failed to read %s dir at %s", dir, dirPath)
	}

	output := []byte{}

	for _, file := range files {
		fileToCopy := filepath.Join(dirPath, file.Name())

		cmd := exec.Command("/bin/bash", "-c",  "source", file.Name())
		if debugMode {
			fmt.Printf("=> Executing: %s\n", strings.Join(cmd.Args, " "))
		}

		err := cmd.Start()
		if err != nil {
			fmt.Errorf("Failed to %s file: %s", dir, fileToCopy)
		}

		time.Sleep(100 * time.Millisecond)

		if cmd.Env == nil {
			fmt.Println("no env variable")
		}
		for _, envv := range cmd.Env {
			fmt.Println(envv)
		}

		if cmd.Stdout != nil {
			return nil, errors.New("exec: Stdout already set")
		}
		if cmd.Stderr != nil {
			return nil, errors.New("exec: Stderr already set")
		}
		var b bytes.Buffer
		cmd.Stdout = &b
		cmd.Stderr = &b
		
		cmd.Wait()
		fmt.Println("out: " + string(b.Bytes()))
		
	}
	

	return output, nil
}



var testEnv = false
