package main

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrDoesNotExistOrIsADirectory = errors.New("file does not exist or is a directory.")

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
	fmt.Println("	run			Run a game")
	fmt.Println("	help			Print this help information and exit")
}

func showUsageOnInvalidOption(s string) {
	fmt.Println("Usage: cli-game-launcher <command>")
	fmt.Printf("[!] Invalid choice: '%v' (choose from add, remove, run, help)", s)
}

func runGamePrompt(reader *bufio.Reader) error {
	err := listGamesNumbered(programHome)
	if err != nil {
		return fmt.Errorf("list games failed: %w", err)
	}

	userInput, err := getIntFromUser(reader)
	if err != nil {
		return fmt.Errorf("get user input failed: %w", err)
	}

	games, err := getGames(programHome)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	err = runNative(games[userInput].PathToExecutable)
	if err != nil {
		return fmt.Errorf("run %v failed: %w", userInput, err)
	}

	return nil
}

// Asks the user for the path to a game's executable
// file and its name and saves this info in
// a configuration file.
func addGamePrompt(reader *bufio.Reader) error {
	fmt.Print("Enter game name: ")
	gameName, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Print("Enter path to executable: ")
	pathToExecutable, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	gameName = strings.TrimSpace(gameName)
	pathToExecutable = strings.TrimSpace(pathToExecutable)
	pathToExecutable, err = getAbsolutePath(pathToExecutable)
	if err != nil {
		return err
	}

	b := fileExists(pathToExecutable)
	if !b {
		return ErrDoesNotExistOrIsADirectory
	}

	fmt.Printf("\n[DEBUG] name: %v; path: %v, exists: %v\n",
		gameName, pathToExecutable, b)

	err = saveGameToConfig(gameName, pathToExecutable, programHome)
	if err != nil {
		return err
	}

	return nil
}

// Shows the list of game entries in the config file and removes the user chosen option.
func removeGamePrompt(reader *bufio.Reader) error {
	err := listGamesNumbered(programHome)
	if err != nil {
		return fmt.Errorf("list games failed: %w", err)
	}

	userInput, err := getIntFromUser(reader)
	if err != nil {
		return fmt.Errorf("get user input failed: %w", err)
	}

	err = removeGameFromConfig(userInput, programHome)
	if err != nil {
		return fmt.Errorf("remove game failed: %w", err)
	}

	return nil
}

// Looks for a file named config.json in the parentDir path and prints
// the index and the game name of the entries in the file.
func listGamesNumbered(parentDir string) error {
	configFile := filepath.Join(parentDir, "config.json")
	if !fileExists(configFile) {
		return ErrDoesNotExistOrIsADirectory
	}

	games, err := getGames(parentDir)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	if len(games) == 0 {
		return ErrNoGamesFound
	}

	var s string
	for i, game := range games {
		s = s + fmt.Sprintf("[%v] %v\n", i, game.Name)
	}

	fmt.Print(s)
	return nil
}

func getIntFromUser(reader *bufio.Reader) (int, error) {
	fmt.Print("Enter a number: ")
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("read string failed: %w", err)
	}

	userInput = strings.TrimSpace(userInput)
	intVal, err := strconv.Atoi(userInput)
	if err != nil {
		return 0, fmt.Errorf("string to int failed: %w", err)
	}

	return intVal, nil
}
