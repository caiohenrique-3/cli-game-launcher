package main

import (
	"bufio"
	"fmt"
	"os"
)

func handleCommandLine() error {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) == 0 {
		showHelp()
	}

	for n := range len(argsWithoutProg) {
		switch argsWithoutProg[n] {
		case "--help", "-h", "help":
			showHelp()
			return nil
		case "add":
			err := addGamePrompt(bufio.NewReader(os.Stdin))
			if err != nil {
				return err
			}
		case "remove":
			err := removeGamePrompt(bufio.NewReader(os.Stdin))
			if err != nil {
				return err
			}
		case "run":
			err := runGamePrompt(bufio.NewReader(os.Stdin))
			if err != nil {
				return err
			}
		default:
			showUsageOnInvalidOption(argsWithoutProg[n])
			var ErrInvalidOption = fmt.Errorf(
				"invalid command-line argument: '%v'", argsWithoutProg[n])
			return ErrInvalidOption
		}
	}

	return nil
}
