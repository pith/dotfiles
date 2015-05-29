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
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
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
	cachePath = filepath.Join(BaseDir, "cache", "cache.json")
}

func main() {
	flag.Parse()

	usr, err := user.Current()
	if err != nil {
		fmt.Errorf("%v", err)
	}
	changeRootDir(usr.HomeDir)

	console.printHeader("    .: Dotfiles :.")
	run()
}

func run() {
	// Not initialize yet
	_, err := os.Stat(BaseDir)
	if err != nil && os.IsNotExist(err) {
		console.printHeader("Your .dotfiles repository is not setup yet.")
		fmt.Printf("\nDo you want to (C)lone a dot repo, Create a (N)ew one, See the (H)elp or (Q)uit ? ")

		var answer string
		fmt.Scan(&answer)
		switch answer {
		case "c", "C":
			fmt.Printf("\nEnter a git URL: ")
			var answer string
			fmt.Scan(&answer)
			cloneRepo(answer)
		case "n", "N":
			initialize()
			os.Exit(1)
		case "h", "H", "?":
			fmt.Println(help)
			os.Exit(1)
		case "q", "Q":
			os.Exit(1)
		default:
			fmt.Println(help)
			os.Exit(1)
		}

	}

	// run the config

	loadCache()

	var dots Dotfiles
	dots.read()
	dots.cp()
	dots.ln()
	dots.init()

	console.printHeader("All done !")
}

// CloneRepo clones the given git repository
func cloneRepo(gitrepo string) {
	console.printHeader("Clone " + gitrepo)
	git, err := exec.LookPath("git")
	if err != nil {
		log.Fatal("git is required to clone the dotfiles repo")
	}

	err = exec.Command(git, "clone", "--recursive", gitrepo, BaseDir).Run()
	if err != nil {
		log.Fatalf("Failed to clone %s: %s", gitrepo, err)
	}

	console.printHeader(BaseDir + " is ready !")
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

// BackupIfExist move a file in the backup dir if it exists
func backupIfExist(file string) (string, string) {
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

// SourceDir source all the files in the source dir
// Solution from here: http://stackoverflow.com/a/29995987/1292605

// func sourceDir() {
// 	applyCmd("source", func(file string) error {
// 		printHeader("Sourcing " + file)

// 		cmd := exec.Command("/bin/bash", "-c", "source "+file+" ; echo '<<<ENVIRONMENT>>>' ; env")
// 		out, err := cmd.CombinedOutput()
// 		if err != nil {
// 			return err
// 		}

// 		s := bufio.NewScanner(bytes.NewReader(out))
// 		start := false
// 		for s.Scan() {
// 			if s.Text() == "<<<ENVIRONMENT>>>" {
// 				start = true
// 			} else if start {
// 				kv := strings.SplitN(s.Text(), "=", 2)
// 				if len(kv) == 2 {
// 					os.Setenv(kv[0], kv[1])
// 				}
// 			}
// 		}

// 		return nil
// 	})
// }
