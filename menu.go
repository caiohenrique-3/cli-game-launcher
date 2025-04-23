package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Prints help information.
func showHelp() {
	fmt.Println("cli game launcher 0.0.1")
	fmt.Println("A minimal game launcher for your terminal")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("cli-game-launcher [OPTIONS] [COMMAND]...")
	fmt.Println("OPTIONS:")
	fmt.Println("	-h, --help 			Print this help information and exit")
	fmt.Println("	-q, --quiet 		Suppress terminal output when launching a game")
	fmt.Println("COMMANDS:")
	fmt.Println("	add 				Add a new game")
	fmt.Println("	help				Print this help information and exit")
}

func showUsageOnInvalidOption(s string) {
	fmt.Println("Usage: cli-game-launcher <command>")
	fmt.Printf("[!] Invalid choice: '%v' (choose from add, help)", s)
}

// Asks the user for the path to a game's executable
// file and its name and saves this info in
// a configuration file.
func addGamePrompt(reader *bufio.Reader) {
	fmt.Print("Enter game name: ")
	gameName, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Enter path to executable: ")
	pathToExecutable, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	gameName = strings.TrimSpace(gameName)
	pathToExecutable = strings.TrimSpace(pathToExecutable)
	pathToExecutable = getCleanPath(pathToExecutable)

	b := fileExists(pathToExecutable)
	if !b {
		log.Fatalf("'%v' does not exist or is a directory.", pathToExecutable)
	}

	fmt.Printf("\n[DEBUG] name: %v; path: %v, exists: %v\n",
		gameName, pathToExecutable, b)

	saveGameToConfig(gameName, pathToExecutable, programHome)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Removes backslashes from path (if not Windows), expands variables,
// expands tilde and makes the path absolute.
func getCleanPath(s string) string {
	if strings.Contains(s, "$") {
		s = os.ExpandEnv(s)
	}
	if strings.HasPrefix(s, "~/") {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		s = filepath.Join(userHomeDir, s[1:])

	}

	if !filepath.IsAbs(s) {
		s, err := filepath.Abs(s)
		if err != nil {
			log.Fatal(err)
		}

		return s
	}
	return s
}
