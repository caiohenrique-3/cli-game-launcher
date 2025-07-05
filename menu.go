package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrDoesNotExistOrIsADirectory = errors.New("file does not exist or is a directory.")

// Prints help information.
func showHelp() {
	sb := strings.Builder{}
	sb.WriteString("cli game launcher 0.0.1\n")
	sb.WriteString("A minimal game launcher for your terminal\n\n")
	sb.WriteString("USAGE:\n")
	sb.WriteString("cli-game-launcher [COMMAND]\n\n")
	sb.WriteString("COMMANDS:\n")
	sb.WriteString("\tadd\t\tAdd a new game\n")
	sb.WriteString("\tremove\t\tRemove a game\n")
	sb.WriteString("\trun\t\tRun a game\n")
	sb.WriteString("\tlist\t\tPrint all known games\n")
	sb.WriteString("\thelp\t\tPrint this help information and exit\n")
	fmt.Print(sb.String())
}

func showUsageOnInvalidOption(s string) {
	fmt.Println("Usage: cli-game-launcher <command>")
	fmt.Printf("[!] Invalid choice: '%v' (choose from add, remove, list, run, help)\n", s)
}

func runGamePrompt(reader *bufio.Reader) error {
	err := listGamesNumbered(programHome, os.Stdout)
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

	err = runNative(games[userInput].PathToExecutable, reader, os.Stdout)
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
		return fmt.Errorf("get game name failed: %w", err)
	}

	fmt.Print("Enter path to executable: ")
	pathToExecutable, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("get executable path failed: %w", err)
	}

	gameName = strings.TrimSpace(gameName)
	pathToExecutable = strings.TrimSpace(pathToExecutable)

	if !filepath.IsAbs(pathToExecutable) {
		newPathToExecutable, err := filepath.Abs(pathToExecutable)
		if err != nil {
			return fmt.Errorf("get absolute path failed: %w", err)
		}
		pathToExecutable = newPathToExecutable
	}

	if !fileExists(pathToExecutable) {
		return ErrDoesNotExistOrIsADirectory
	}

	err = saveGameToConfig(gameName, pathToExecutable, programHome)
	if err != nil {
		return fmt.Errorf("save game to config failed: %w", err)
	}

	fmt.Printf("'%s' added.\n", gameName)

	return nil
}

// Shows the list of game entries in the config file and removes the user chosen option.
func removeGamePrompt(reader *bufio.Reader) error {
	err := listGamesNumbered(programHome, os.Stdout)
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
func listGamesNumbered(parentDir string, writer io.Writer) error {
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

	fmt.Fprint(writer, s)
	return nil
}

func getIntFromUser(reader *bufio.Reader) (int, error) {
	fmt.Print("Enter a number: ")
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("read string failed: %w", err)
	}

	userInput = strings.TrimSpace(userInput)
	intValue, err := strconv.Atoi(userInput)
	if err != nil {
		return 0, fmt.Errorf("string to int failed: %w", err)
	}

	return intValue, nil
}
