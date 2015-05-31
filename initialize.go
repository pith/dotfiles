// Copyright (c) 2015 by Pierre Thirouin. All rights reserved.

// This file is part of dotfiles, a simple dotfiles manager.

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
