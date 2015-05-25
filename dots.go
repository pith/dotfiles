package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
)

var console Console

// Dir corresponds a dotfiles directory
type Dir int

const (
	ln Dir = iota
	cp
	rn
)

func (d Dir) String() string {
	var s string
	switch d {
	case ln:
		s = "link"
	case cp:
		s = "copy"
	case rn:
		s = "init"
	}
	return s
}

// Dotfiles stores all the dot files by directory
type Dotfiles struct {
	Files map[Dir][]string
}

func (dots *Dotfiles) read() {
	dirs := [3]Dir{ln, cp, rn}

	if dots.Files == nil {
		dots.Files = make(map[Dir][]string)
	}

	for _, dir := range dirs {

		dirPath := filepath.Join(BaseDir, dir.String())

		files, err := ioutil.ReadDir(dirPath)
		if err != nil {
			log.Fatalf("Failed to read %s dir: %s", dir, err)
		}

		for _, file := range files {
			dots.Files[dir] = append(dots.Files[dir], filepath.Join(dirPath, file.Name()))
		}
	}

}

// BackupFiles backs up files which will be copyied or linked,
// but which don't appear in the cache
func (dots Dotfiles) backup(dir Dir, action Action) {
	firstBackup := true

	for _, f := range dots.Files[dir] {
		contains, err := cacheContains(action, f)
		if err != nil {
			log.Fatal(err)
		}
		if !contains {
			path, backupPath := backupIfExist(f)

			// print feedback
			if path != "" && backupPath != "" {
				if firstBackup {
					console.printHeader("Backup before " + dir.String() + "ing")
					firstBackup = false
				}
				relPath, err := filepath.Rel(RootDir, backupPath)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf(" %s ➜ %s\n", filepath.Base(path), relPath)
			}
		}
	}

}

func (dots Dotfiles) cp() {
	dots.backup(cp, copy)

	console.printHeader("Copying files into home directory")

	for _, f := range dots.Files[cp] {
		if backgroundCheck(f) {

			console.printArrow(filepath.Base(f))

			cacheAdd(copy, f)

			err := exec.Command("cp", f, RootDir).Run()
			if err != nil {
				fmt.Errorf("Failed to copy %s", f)
			}
		}
	}
}

func (dots Dotfiles) ln() {
	dots.backup(ln, link)

	console.printHeader("Linking files into home directory")

	for _, f := range dots.Files[ln] {
		if backgroundCheck(f) {

			console.printArrow(f)

			cacheAdd(link, f)

			err := exec.Command("ln", "-sf", f, RootDir).Run()
			if err != nil {
				fmt.Errorf("Failed to link %s", f)
			}
		}
	}
}

func (dots Dotfiles) init() []byte {
	scripts := make(map[string]bool)

	// By default run the script not cached
	for _, f := range dots.Files[rn] {
		contains, err := cacheContains(initRun, f)
		if err != nil {
			log.Fatal(err)
		}
		scripts[f] = !contains
	}

	console.printMenu(scripts)

	edited := console.editMenu(scripts)
	if edited != nil {
		console.printMenu(scripts)
	}

	var out []byte

	for _, f := range dots.Files[rn] {
		if scripts[f] {
			console.printHeader("Run " + filepath.Base(f))

			// output, err := exec.Command("/bin/bash", "-c", "source "+f).CombinedOutput()
			// out = append(out, output...)
			// if err != nil {
			// 	cacheAdd(initRun, f)
			// }
		}
	}

	return out
}
