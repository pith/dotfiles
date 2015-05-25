package main

import (
	"fmt"
	"strconv"
)

// Console prints string to the console
type Console struct{}

func (c Console) print(s string) {
	if !quietMode {
		fmt.Printf("%s", s)
	}
}

func (c Console) printHeader(s string) {
	c.print(fmt.Sprintf("\n\033[1m%s\033[0m\n", s))
}

func (c Console) printArrow(s string) {
	c.print(fmt.Sprintf(" \033[1;34m➜\033[0m  %s\n", s))
}

func (c Console) printOK(s string) {
	c.print(fmt.Sprintf(" \033[1;32m✔\033[0m  %s\n", s))
}

func (c Console) printKO(s string) {
	c.print(fmt.Sprintf(" \033[1;32m✖\033[0m  %s\n", s))
}

func (c Console) printMenu(scripts map[string]bool) {
	i := 1
	for name, ok := range scripts {
		if ok {
			c.printOK(strconv.Itoa(i) + ". " + name)
		} else {
			c.printKO(strconv.Itoa(i) + ". " + name)
		}
		i++
	}
}

func (c Console) editMenu(scripts map[string]bool) map[string]bool {
	fmt.Printf("\nEdit scripts to run ? (Y/n): ")
	var input string
	fmt.Scan(&input)

	var toggled []int
	if input == "Y" || input == "y" {
		fmt.Printf("\nEnter the script ids that you want to toggle: ")
		fmt.Scan(&toggled)
	} else {
		return nil
	}

	for _, script := range toggled {
		i := 1
		for name, ok := range scripts {
			if i == script {
				scripts[name] = !ok
			}
			i++
		}
	}
	return scripts
}
