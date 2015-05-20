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
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	debugMode = false
	quietMode = false
)

var help = `usage: dotfiles [command]

Command list

Setup your machine:
  get [url] Clone a dotfiles project at the given Git URL     
  add      Add a file to one of the dotfiles directories
  run      Initialize your config based on ~/.dotfiles

Additional commands:
  init     Run the init scripts
  copy     Copy the files in ~/.dotfiles/copy to ~/
  link     Link the files in ~/.dotfiles/link to ~/

`

// Default paths
var (
	// RootDir is the directory where files will be link or copy.
	RootDir = "."

	// DotFilesDir is the name of the directory where the dotfiles are stored.
	DotFilesDir = ".dotfiles"

	// BaseDir is the path to the dotfiles directory
	BaseDir = filepath.Join(RootDir, DotFilesDir)

	dirs = [8]string{"bin", "conf", "copy", "init", "link", "source", "test", "vendor"}
)

// flags
var (
	noCache = flag.Bool("nocache", false, "The script will be run like the first time.")
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

	loadCache()

	changeRootDir(usr.HomeDir)

	printHeader("    .: Dotfiles :.")

	switch command {
	case "create":
		initialize()
	case "get":
		cloneRepo(flag.Arg(1))
	case "run":
		copyDir()

		linkDir()

		printHeader("Run the following init scripts")
		initDir()

		printHeader("All done !")
	case "add":
		addCmd(flag.Arg(1), flag.Arg(2))
	case "init":
		initDir()
	case "copy":
		copyDir()
	case "link":
		linkDir()
	default:
		fmt.Print(help)
	}
}

func addCmd(cmd, file string) {
	switch cmd {
	case "bin":
		cpToDot("bin", file)
	case "conf":
		cpToDot("conf", file)
	case "copy":
		cpToDot("copy", file)
	case "init":
		cpToDot("init", file)
	case "link":
		cpToDot("link", file)
	case "source":
		cpToDot("source", file)
	case "test":
		cpToDot("test", file)
	case "vendor":
		cpToDot("vendor", file)
	default:
		fmt.Printf("Unknown dotfiles directory: %s\n", cmd)
	}
}

func printHeader(s string) {
	if !quietMode {
		fmt.Printf("\n\033[1m%s\033[0m\n", s)
	}
}

func printArrow(s string) {
	if !quietMode {
		fmt.Printf(" \033[1;34m➜\033[0m  %s\n", s)
	}
}

// CloneRepo clones the given git repository
func cloneRepo(gitrepo string) {
	printHeader("Clone " + gitrepo)
	git, err := exec.LookPath("git")
	if err != nil {
		fmt.Errorf("git is required to clone the dotfiles repo")
	}

	err = exec.Command(git, "clone", "--recursive", gitrepo, BaseDir).Run()
	if err != nil {
		fmt.Errorf("Failed to clone %s", gitrepo)
	}

	printHeader(BaseDir + " is ready !")
}

// BackgroundCheck verifies if there are some actions to do on the given file
// Returns true if the destination file doesn't exist or if it is different
// from the source file
func backgroundCheck(file string) bool {
	source, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		// Can't background check a file which doesn't exists
		fmt.Errorf("%s: no such file or directory", file)
	}

	_, err = os.Stat(filepath.Join(RootDir, filepath.Base(file)))
	if err != nil && os.IsNotExist(err) {
		// The destination file doesn't exist so go ahead
		return true
	}

	if !source.Mode().IsRegular() {
		// Don't do a deep check on non-regular files (eg. directories, link, etc.),
		// so if the destination file exists don't do anything
		return false
	}

	// Deep comparison between the two files
	sf, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	df, err := os.Open(filepath.Join(RootDir, filepath.Base(file)))
	if err != nil {
		log.Fatal(err)
	}

	sscan := bufio.NewScanner(sf)
	dscan := bufio.NewScanner(df)

	for sscan.Scan() {
		dscan.Scan()
		if !bytes.Equal(sscan.Bytes(), dscan.Bytes()) {
			return true
		}
	}

	return false

}

// BackupFiles backs up files which will be copyied or linked,
// but which don't appear in the cache
func backupFiles(dirName string, action Action) {
	firstBackup := true

	applyCmd(dirName, func(fileToCopy string) error {
		contains, err := cacheContains(action, fileToCopy)
		if err != nil {
			log.Fatal(err)
		}
		if !contains {
			path, backupPath := backupIfExist(fileToCopy)
			if path != "" && backupPath != "" {
				if firstBackup {
					printHeader("Backup before " + dirName + "ing")
					firstBackup = false
				}
				fmt.Printf(" %s ➜ %s\n", path, backupPath)
			}
		}

		return nil
	})

}

// BackupIfExist move a file in the backup dir if it exists
func backupIfExist(file string) (string, string){
	file = filepath.Base(file)

	// If there is no backup dir yet create it
	if _, err := os.Stat(filepath.Join(BaseDir, "backup")); os.IsNotExist(err) {
		err = os.Mkdir(filepath.Join(BaseDir, "backup"), 0777)
		if err != nil {
			log.Fatal("Failed to create backup dir: ", err)
		}
	}

	path := filepath.Join(RootDir, file)

	backupPath := filepath.Join(BaseDir, "backup", file)
	if _, err := os.Stat(path); err == nil {
		// The file already exists so backup it
		err = exec.Command("mv", path, backupPath).Run()
		if err != nil {
			fmt.Errorf("Failed to backup %s\n%v", file, err)
		}
		return path, backupPath
	}
	return "", ""
}

// CopyDir copies all the files in the copy dir at ~/
func copyDir() {
	backupFiles("copy", copy)
	
	printHeader("Copying files into home directory")

	applyCmd("copy", func(fileToCopy string) error {
		if backgroundCheck(fileToCopy) {
			printArrow(fileToCopy)
			cacheAdd(copy, fileToCopy)
			return exec.Command("cp", fileToCopy, RootDir).Run()
		}
		return nil
	})
}

// LinkDir links all the files in the link dir at ~/
func linkDir() {
	backupFiles("link", link)
	
	printHeader("Linking files into home directory")

	applyCmd("link", func(file string) error {
		if backgroundCheck(file) {
			printArrow(file)
			cacheAdd(link, file)
			return exec.Command("ln", "-sf", file, RootDir).Run()
		}
		return nil
	})
}

func firstInit() bool {
	if *noCache {
		return true
	}

	if _, err := os.Stat(filepath.Join(BaseDir, "cache")); os.IsNotExist(err) {
		return true
	}
	return false
}

// InitDir executes all the scripts in the init dir
func initDir() []byte {

	i := 1
	applyCmd("init", func(initFile string) error {
		printArrow(strconv.Itoa(i) + ". " + initFile)
		return nil
	})

	// TODO add the possibility to run them again based on user input
	var out []byte
	if firstInit() {
		out = doInitDir()

	} else {
		fmt.Printf("\nRerun init scripts (Y/n): ")
		var input string
		fmt.Scan(&input)
		if input == "Y" || input == "y" {
			doInitDir()
		}
	}
	return out
}

func doInitDir() []byte {
	var out []byte

	applyCmd("init", func(initFile string) error {
		printHeader("Run " + initFile)

		output, err := exec.Command("/bin/bash", "-c", "source "+initFile).CombinedOutput()
		out = append(out, output...)
		if err != nil {
			cacheAdd(initRun, initFile)
		}
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
