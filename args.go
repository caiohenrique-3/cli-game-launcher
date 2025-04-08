package main

import "os"

// Handles the command-line arguments and commands.
func readArgs() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) == 0 {
		showHelp()
	}

	for n := range len(argsWithoutProg) {
		switch argsWithoutProg[n] {
		case "--help", "-h", "help":
			showHelp()
			return
		case "add":
			addGamePrompt()
			return
		default:
			showUsageOnInvalidOption(argsWithoutProg[n])
			return
		}
	}
}
