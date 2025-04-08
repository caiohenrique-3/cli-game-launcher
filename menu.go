package main

import "fmt"

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
// file and its name and saves this information in
// a configuration file under XDG_CONFIG_HOME.
func addGamePrompt() {
	fmt.Println("(≖_≖ ) soon.")
}
