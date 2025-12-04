package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

// Prints help information.
func showHelp() {
	var sb strings.Builder
	sb.WriteString("cli game launcher\n")
	sb.WriteString("A minimal game launcher for your terminal\n\n")
	sb.WriteString("USAGE:\n")
	sb.WriteString("cli-game-launcher [COMMAND] [OPTIONS]\n\n")
	sb.WriteString("list OPTIONS:\n")
	sb.WriteString("\t-show-hidden, --show-hidden=<BOOL>\tShow hidden games in output (default: false)\n")
	sb.WriteString("\t-only-hidden, --only-hidden=<BOOL>\tOnly show hidden games in output  (default: false)\n\n")
	sb.WriteString("run OPTIONS:\n")
	sb.WriteString("\t-show-hidden, --show-hidden=<BOOL>\tShow hidden games in output (default: false)\n\n")
	sb.WriteString("COMMANDS:\n")
	sb.WriteString("\tadd\t\tAdd a new game\n")
	sb.WriteString("\tremove\t\tRemove a game\n")
	sb.WriteString("\trun\t\tRun a game\n")
	sb.WriteString("\tlist\t\tPrint all known games\n")
	sb.WriteString("\tplaytime\tShow playtime for each game\n")
	sb.WriteString("\tlast-played\tView last played dates\n")
	sb.WriteString("\tlast-two-weeks\tShow playtime for last two weeks\n")
	sb.WriteString("\thide\t\tToggle the visibility of a game\n")
	sb.WriteString("\thelp\t\tPrint this help information and exit\n")

	fmt.Print(sb.String())
}

// Asks for a game in the database and runs it.
func runGamePrompt(reader *bufio.Reader, cmdOptions RunCmdOptions) error {
	var excludeHiddenGames bool
	if cmdOptions.showHidden {
		excludeHiddenGames = false
	} else {
		excludeHiddenGames = true
	}

	games, err := getGames(programHome, excludeHiddenGames)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	err = listGamesNumbered(games, os.Stdout)
	if err != nil {
		return fmt.Errorf("list games failed: %w", err)
	}

	userInput, err := getIntFromUser(reader)
	if err != nil {
		return fmt.Errorf("get user input failed: %w", err)
	}

	// Running game
	timeStart := time.Now()
	timeSpentPlaying, err := runNative(games[userInput].Path,
		reader, os.Stdout)
	if err != nil {
		fmt.Printf("run %v failed: %v", userInput, err)
	}

	// Saving playtime
	pathToDb := filepath.Join(programHome, "data.db")
	err = saveTimeSpentPlaying(games[userInput].Id, timeSpentPlaying, pathToDb)
	if err != nil {
		return fmt.Errorf("save time spent playing failed: %w", err)
	}

	// Saving gaming session to logs
	timeEnd := timeStart.Add(timeSpentPlaying)
	err = saveSession(games[userInput].Id, timeStart, timeEnd, pathToDb)
	if err != nil {
		return fmt.Errorf("log session failed: %w", err)
	}

	return nil
}

// Asks a game name and path and adds it to the config file.
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

	err = saveGameToConfig(gameName, pathToExecutable, programHome)
	if err != nil {
		return fmt.Errorf("save game to config failed: %w", err)
	}

	fmt.Printf("'%s' added.\n", gameName)
	return nil
}

// Shows the list of game entries in the database and removes the chosen option.
func removeGamePrompt(reader *bufio.Reader) error {
	games, err := getGames(programHome, false)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	err = listGamesNumbered(games, os.Stdout)
	if err != nil {
		return fmt.Errorf("list games failed: %w", err)
	}

	userInput, err := getIntFromUser(reader)
	if err != nil {
		return fmt.Errorf("get user input failed: %w", err)
	}

	err = removeGameFromConfig(games[userInput].Id, programHome)
	if err != nil {
		return fmt.Errorf("remove game failed: %w", err)
	}

	return nil
}

// Shows the list of games in the database and hides or unhides the chosen option.
func hideGamePrompt(reader *bufio.Reader) error {
	games, err := getGames(programHome, false)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	err = listGamesNumbered(games, os.Stdout)
	if err != nil {
		return fmt.Errorf("list games failed: %w", err)
	}

	userInput, err := getIntFromUser(reader)
	if err != nil {
		return fmt.Errorf("get user input failed: %w", err)
	}

	err = toggleHiddenGame(games[userInput].Id, programHome)
	if err != nil {
		return fmt.Errorf("hide game failed: %w", err)
	}

	return nil
}

// Prints list of games with names and their indexes on the slice.
func listGamesNumbered(games []Game, writer io.Writer) error {
	var sb strings.Builder
	for i, game := range games {
		if game.Hidden {
			fmt.Fprintf(&sb, "[%d] %s [HIDDEN]\n", i, game.Name)
		} else {
			fmt.Fprintf(&sb, "[%d] %s\n", i, game.Name)
		}
	}

	fmt.Fprint(writer, sb.String())
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

// Prints a table with games and the time spent playing them.
func listGamesWithPlaytime(configFileParentDir string, writer io.Writer) error {
	pathToDb := filepath.Join(configFileParentDir, "data.db")
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT name, playtime FROM games")
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	padding := 3
	tw := tabwriter.
		NewWriter(writer, 0, 0, padding, ' ', tabwriter.AlignRight)

	// Table header
	fmt.Fprintln(tw, "Game Name\tTime Spent Playing\t")

	// Table game rows
	var gameRows strings.Builder

	for rows.Next() {
		var gameName, playtime string
		err := rows.Scan(&gameName, &playtime)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		gameRows.WriteString(fmt.Sprintf("%s\t%s\t\n",
			gameName, playtime))
	}

	fmt.Fprintln(tw, gameRows.String())
	tw.Flush()
	return nil
}

// Prints a table with games and their last played date.
func listGamesWithLastPlayed(configFileParentDir string, writer io.Writer) error {
	games, err := getGamesWithLastPlayedTime(configFileParentDir)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	if len(games) == 0 {
		return ErrNoGamesFound
	}

	padding := 4
	tw := tabwriter.
		NewWriter(writer, 0, 0, padding, ' ', tabwriter.AlignRight)

	// Table header
	fmt.Fprintln(tw, "ID\tGame Name\tLast Time Played\t")

	// Table game rows
	var gameRows strings.Builder
	timeNow := time.Now()

	for _, game := range games {
		gameRows.WriteString(fmt.Sprintf("%d\t", game.Id))
		gameRows.WriteString(fmt.Sprintf("%s\t", game.Name))

		timeSince := time.Duration.Round(timeNow.Sub(game.LastPlayed), time.Second)
		sameDay := timeNow.Format(time.DateOnly) == game.LastPlayed.Format(time.DateOnly)

		if sameDay {
			// Prints "Today (x ago)"
			gameRows.WriteString(fmt.Sprintf("%s (%s ago)\t\n",
				"Today",
				timeSince))
		} else {
			// Prints "<date> (x ago)"
			gameRows.WriteString(fmt.Sprintf("%s (%s ago)\t\n",
				game.LastPlayed.Format(time.DateOnly),
				timeSince))
		}
	}

	fmt.Fprintln(tw, gameRows.String())
	tw.Flush()
	return nil
}

// Prints a table with games and the time spent playing them in the last two weeks.
func listGamesWithPlaytimeLastTwoWeeks(configFileParentDir string, writer io.Writer) error {
	games, totalPlaytime, err := getGamesWithPlaytimeLastTwoWeeks(configFileParentDir)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	if len(games) == 0 {
		return ErrNoGamesFound
	}

	padding := 4
	tw := tabwriter.
		NewWriter(writer, 0, 0, padding, ' ', tabwriter.AlignRight)

	// Table header
	fmt.Fprintln(tw, "ID\tGame Name\tTime Spent Playing Last Two Weeks\t")

	// Table game rows
	var gameRows strings.Builder

	for _, game := range games {
		gameRows.WriteString(fmt.Sprintf("%d\t", game.Id))
		gameRows.WriteString(fmt.Sprintf("%s\t", game.Name))

		percentOfTotal := (game.Playtime.Hours() / totalPlaytime.Hours()) * 100
		gameRows.WriteString(fmt.
			Sprintf("%s (%.1f%%)\t\n",
				game.Playtime,
				percentOfTotal))
	}

	fmt.Fprint(tw, gameRows.String())
	tw.Flush()
	fmt.Fprintf(writer, "\n%s spent playing last two weeks.\n", totalPlaytime)
	return nil
}
