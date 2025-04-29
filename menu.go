package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	fmt.Println("	-h, --help 		Print this help information and exit")
	fmt.Println("	-q, --quiet 		Suppress terminal output when launching a game")
	fmt.Println("COMMANDS:")
	fmt.Println("	add 			Add a new game")
	fmt.Println("	remove			Remove a game")
	fmt.Println("	help			Print this help information and exit")
}

func showUsageOnInvalidOption(s string) {
	fmt.Println("Usage: cli-game-launcher <command>")
	fmt.Printf("[!] Invalid choice: '%v' (choose from add, remove, help)", s)
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

// Shows the list of game entries in the config file and removes the user chosen option.
func removeGamePrompt(reader *bufio.Reader) {
	listGamesNumbered(programHome)
	fmt.Print("Enter number you want to delete: ")
	userInput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	userInput = strings.TrimSpace(userInput)
	intVal, err := strconv.Atoi(userInput)
	if err != nil {
		log.Fatal(err)
	}

	removeGameFromConfig(intVal, programHome)
}

// Looks for a file named config.json in the parentDir path and prints
// the index and the game name of the entries in the file.
func listGamesNumbered(parentDir string) {
	configFile := filepath.Join(parentDir, "config.json")
	if !fileExists(configFile) {
		log.Fatalf("'%v' does not exist or is a directory.", configFile)
	}

	// TODO: Extract this & move to config.go
	f := loadConfigFile(parentDir)
	games := []Game{}
	err := json.Unmarshal(f, &games)
	if err != nil {
		log.Fatal(err)
	}

	var s string
	for i, game := range games {
		s = s + fmt.Sprintf("[%v] %v\n", i, game.Name)
	}

	fmt.Print(s)
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
