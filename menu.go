package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
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
	fmt.Println("COMMANDS:")
	fmt.Println("	add 			Add a new game")
	fmt.Println("	remove			Remove a game")
	fmt.Println("	run			Run a game")
	fmt.Println("	list			Display all games")
	fmt.Println("	playtime		Show playtime for each game")
	fmt.Println("	last-played		View last played dates")
	fmt.Println("	last-two-weeks		Show playtime for last two weeks")
	fmt.Println("	help			Print this help information and exit")
}

func showUsageOnInvalidOption(s string) {
	fmt.Println("Usage: cli-game-launcher <command>")
	fmt.Printf("[!] Invalid choice: '%v' (choose from add, remove, run, list, playtime, last-played, last-two-weeks, help)\n", s)
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

	// Running game
	timeStart := time.Now() // This is needed for saving to logs later.
	timeSpentPlaying, err :=
		runNative(games[userInput].PathToExecutable, reader, os.Stdout)
	if err != nil {
		return fmt.Errorf("run %v failed: %w", userInput, err)
	}

	// Saving playtime
	pathToConfigFile := filepath.Join(programHome, "config.json")
	err = saveTimeSpentPlayingToConfig(games, userInput,
		timeSpentPlaying, pathToConfigFile)
	if err != nil {
		return fmt.Errorf("save time spent playing failed: %w", err)
	}

	// Saving gaming session to logs
	pathToLogFile := filepath.Join(programHome, "logs.json")
	timeEnd := timeStart.Add(timeSpentPlaying)
	err = saveSessionToLogs(games[userInput].Name, timeStart, timeEnd, pathToLogFile)
	if err != nil {
		return fmt.Errorf("log session failed: %w", err)
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

// Prints a table with the game name and time spent playing that game.
func listGamesWithPlaytime(configFileParentDir string, writer io.Writer) error {
	games, err := getGames(configFileParentDir)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	padding := 3
	tw := tabwriter.
		NewWriter(writer, 0, 0, padding, ' ', tabwriter.AlignRight)

	// Table header
	fmt.Fprintln(tw, "Game Name\tTime Spent Playing\t")

	// Table game rows
	var gameRows strings.Builder
	for _, game := range games {
		if game.TimeSpentPlaying != "" {
			gameRows.WriteString(
				fmt.Sprintf("%s\t%s\t\n",
					game.Name,
					game.TimeSpentPlaying))
		} else {
			gameRows.WriteString(
				fmt.Sprintf("%s\t0h0m0s\t\n",
					game.Name))
		}
	}

	fmt.Fprintln(tw, gameRows.String())
	tw.Flush()
	return nil
}

/*
Reads 'logs.json' file inside config file parent dir
and, if the game from the log entry is also in the config
file (that means it's not removed), it prints the game name
and it's last played date, alongside how many days ago it was.
*/
func listGamesWithLastPlayed(configFileParentDir string, writer io.Writer) error {
	gamesInConfigFile, err := getGames(configFileParentDir)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")
	logFile, err := os.Open(pathToLogFile)
	if err != nil {
		return fmt.Errorf("open log file failed: %w", err)
	}
	defer logFile.Close()

	// Size of the file is needed for the backwards Scanner
	logFileInfo, err := os.Stat(pathToLogFile)
	if err != nil {
		return fmt.Errorf("read log file info failed: %w", err)
	}

	gamesWithLastPlayedDate := make(map[string]time.Time, len(gamesInConfigFile))
	reDate := regexp.MustCompile(`\[(.*?)\]`)
	reGameName := regexp.MustCompile(`\'(.*?)\'`)
	backScanner := NewScanner(logFile, int(logFileInfo.Size()))

	/* Reading log file line by line, starting by the end to the start,
	 extracting date and game name.
	Example: [2025-06-13] Played 'My Game' from 14:26 to 14:27. */
	for {
		line, _, err := backScanner.Line()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("scan line failed: %w", err)
			} else {
				break
			}
		}

		if line != "" {
			// Using regex to find date and game name
			dateMatch := reDate.FindStringSubmatch(line)
			gameNameMatch := reGameName.FindStringSubmatch(line)
			if gameNameMatch[1] == "" || dateMatch[1] == "" {
				continue
			}

			for _, game := range gamesInConfigFile {
				/* Check if the game extracted from the log file
				is also on the config file, to prevent deleted
				games from showing up in the output*/
				if game.Name != gameNameMatch[1] {
					continue
				}

				/* If key exists, we don't need to update it,
				since we are reading from the end of the file
				to the start, the first entry we came across is
				already the most recent one.*/
				if _, ok := gamesWithLastPlayedDate[game.Name]; ok {
					continue
				}

				lastTimePlayed, err := time.
					ParseInLocation(
						time.DateOnly, dateMatch[1], time.UTC)
				if err != nil {
					return fmt.Errorf("parse date failed: %w", err)
				}
				gamesWithLastPlayedDate[game.Name] = lastTimePlayed
			}
		}
	}

	if len(gamesWithLastPlayedDate) == 0 {
		fmt.Fprintln(writer, "No games found in log file.")
		return nil
	}

	padding := 4
	tw := tabwriter.
		NewWriter(writer, 0, 0, padding, ' ', tabwriter.AlignRight)

	// Table header
	fmt.Fprintln(tw, "Game Name\tLast Time Played\t")

	// Table game rows
	var gameRows strings.Builder
	timeNow := time.Now().UTC()

	for gameName, lastPlayedTime := range gamesWithLastPlayedDate {
		if _, err := gameRows.WriteString(fmt.
			Sprintf("%s\t", gameName)); err != nil {
			return fmt.Errorf("string builder write failed: %w", err)
		}

		// Comparing time now to last played time (both are UTC)
		daysAgo := int(timeNow.Sub(lastPlayedTime) / (24 * time.Hour))

		// Prints "Today"
		if daysAgo == 0 {
			if _, err := gameRows.WriteString(fmt.
				Sprintf("%s", "Today\t\n")); err != nil {
				return fmt.Errorf("string builder write failed: %w", err)
			}
			continue
		}

		// Prints "<date> (1 day ago)"
		if daysAgo == 1 {
			if _, err := gameRows.WriteString(fmt.
				Sprintf("%s (%d day ago)\t\n",
					lastPlayedTime.Format(time.DateOnly),
					daysAgo)); err != nil {
				return fmt.Errorf("string builder write failed: %w", err)
			}
			continue
		}

		// Prints "<date> (x days ago)"
		if _, err := gameRows.WriteString(fmt.
			Sprintf("%s (%d days ago)\t\n",
				lastPlayedTime.Format(time.DateOnly),
				daysAgo)); err != nil {
			return fmt.Errorf("string builder write failed: %w", err)
		}
	}

	fmt.Fprintln(tw, gameRows.String())
	tw.Flush()
	return nil
}

// Prints table with game names and total time played in the last two weeks.
func listGamesWithPlaytimeLastTwoWeeks(configFileParentDir string, writer io.Writer) error {
	games, totalPlaytime, err :=
		getGamesWithPlaytimeLastTwoWeeks(configFileParentDir)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	if len(games) == 0 {
		fmt.Fprintln(writer, "No games found in log file.")
		return nil
	}

	padding := 4
	tw := tabwriter.
		NewWriter(writer, 0, 0, padding, ' ', tabwriter.AlignRight)

	// Table header
	fmt.Fprintln(tw, "Game Name\tTime Spent Playing Last Two Weeks\t")

	// Table game rows
	var gameRows strings.Builder

	for gameName, playtime := range games {
		if _, err := gameRows.WriteString(fmt.
			Sprintf("%s\t", gameName)); err != nil {
			return fmt.Errorf("string builder write failed: %w", err)
		}

		percentOfTotal := (playtime.Hours() / totalPlaytime.Hours()) * 100
		if _, err := gameRows.WriteString(fmt.
			Sprintf("%s (%.1f%%)\t\n",
				playtime,
				percentOfTotal)); err != nil {
			return fmt.Errorf("string builder write failed: %w", err)
		}
	}

	fmt.Fprint(tw, gameRows.String())
	tw.Flush()
	fmt.Fprintf(writer, "\n%s spent playing last two weeks.\n", totalPlaytime)
	return nil
}
