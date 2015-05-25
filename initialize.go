package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Initialize creates a new .dotfiles repo
func initialize() {
	console.printHeader("Scaffold " + BaseDir)

	for _, dir := range dirs {
		console.printArrow(dir)
		os.MkdirAll(filepath.Join(BaseDir, dir), 0777)
	}
	if !quietMode {
		fmt.Println("")
	}
	console.printHeader("All done !")
}

func cpToDot(cmd, path string) {
	if err := exec.Command("cp", path, filepath.Join(BaseDir, cmd)).Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Added %s\n", filepath.Join(BaseDir, cmd, filepath.Base(path)))
}
