package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func (c Console) printMenu(scripts []string, shouldBeRun map[string]bool) {
	console.printHeader("Run the following init scripts")

	for i, script := range scripts {
		if shouldBeRun[script] {
			c.printOK(strconv.Itoa(i) + ". " + filepath.Base(script))
		} else {
			c.printKO(strconv.Itoa(i) + ". " + filepath.Base(script))
		}
	}
}

func (c Console) editMenu(scripts []string, shouldBeRun map[string]bool) map[string]bool {
	fmt.Printf("\nEdit scripts to run ? (Y/n): ")
	var input string
	fmt.Scan(&input)

	var toggled []string
	if input == "Y" || input == "y" {
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("\nEnter the script ids that you want to toggle: ")

		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		text = strings.TrimSuffix(text, "\n")

		// check if ',' is used as separator
		if strings.Contains(text, ",") {
			toggled = strings.Split(text, ",")
		} else {
			toggled = strings.Split(text, " ")
		}
	} else {
		return nil
	}

	for _, script := range toggled {
		id, err := strconv.Atoi(script)
		if err != nil {
			fmt.Printf("Expected script ids but found %s\n: %s", script, err)
		}

		for i, f := range scripts {
			if i == id {
				shouldBeRun[f] = !shouldBeRun[f]
			}
		}

	}
	return shouldBeRun
}
