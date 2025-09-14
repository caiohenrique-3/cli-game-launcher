package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// Saves name of the game played, time start and end to 'logs.json' file.
func saveSessionToLogs(gameName string, timeStart time.Time,
	timeEnd time.Time, pathToLogFile string) error {
	timeStartDateString := timeStart.Format(time.DateOnly)
	timeEndDateString := timeEnd.Format(time.DateOnly)
	var stringToAppend string

	if timeStartDateString != timeEndDateString {
		// That means the date has changed from when it started.
		// [2006-01-02] Played 'game' from 15:04:05 to 2015-07-03 15:32:05.
		stringToAppend = fmt.Sprintf("[%s] Played '%s' from %s to %s.\n",
			timeStartDateString,
			gameName,
			timeStart.Format(time.TimeOnly),
			timeEnd.Format(time.DateTime))
	} else {
		// [2006-01-02] Played 'game' from 15:04:05 to 15:32:05.
		stringToAppend = fmt.Sprintf("[%s] Played '%s' from %s to %s.\n",
			timeStartDateString,
			gameName,
			timeStart.Format(time.TimeOnly),
			timeEnd.Format(time.TimeOnly))
	}

	logFile, err := os.OpenFile(pathToLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("open log file failed: %w", err)
	}
	defer logFile.Close()

	_, err = logFile.WriteString(stringToAppend)
	if err != nil {
		return fmt.Errorf("write to log file failed: %w", err)
	}

	return nil
}

// Returns a map of games with their last played times.
func getGamesWithLastPlayedTime(configFileParentDir string) (map[string]time.Time, error) {
	gamesInConfigFile, err := getGames(configFileParentDir, false)
	if err != nil {
		return nil, fmt.Errorf("get games failed: %w", err)
	}

	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")
	logFile, err := os.Open(pathToLogFile)
	if err != nil {
		return nil, fmt.Errorf("open log file failed: %w", err)
	}
	defer logFile.Close()

	// Size of the file is needed for the backwards Scanner
	logFileInfo, err := os.Stat(pathToLogFile)
	if err != nil {
		return nil, fmt.Errorf("read log file info failed: %w", err)
	}

	gamesWithLastPlayedDate := make(map[string]time.Time, len(gamesInConfigFile))
	regexDate := regexp.MustCompile(`\[(.*?)\]`)
	regexGameName := regexp.MustCompile(`\'(.*?)\'`)
	regexTime := regexp.MustCompile(`\d{2}:\d{2}:\d{2}`)
	backScanner := NewScanner(logFile, int(logFileInfo.Size()))

	/* Reading log file line by line, starting by the end to the start,
	extracting date and game name.
	Example: [2025-06-13] Played 'My Game' from 14:26:30 to 14:27:00. */
	for {
		line, _, err := backScanner.Line()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("scan line failed: %w", err)
			} else {
				return gamesWithLastPlayedDate, nil
			}
		}

		if line == "" {
			continue
		}

		dateMatch := regexDate.FindStringSubmatch(line)
		gameNameMatch := regexGameName.FindStringSubmatch(line)
		timesMatch := regexTime.FindAllString(line, 2)
		if gameNameMatch[1] == "" || dateMatch[1] == "" ||
			len(timesMatch) != 2 {
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

			startDateFromLogs := dateMatch[1]
			endHourFromLogs := timesMatch[1]

			lastTimePlayed, err := getTimeFromStrings(startDateFromLogs,
				endHourFromLogs)
			if err != nil {
				return nil, fmt.Errorf("get time from strings failed: %w",
					err)
			}

			gamesWithLastPlayedDate[game.Name] = lastTimePlayed
		}
	}
}

// Returns a map of games with their playtimes and the
// total playtime across all games in the last two weeks.
func getGamesWithPlaytimeLastTwoWeeks(
	configFileParentDir string) (map[string]time.Duration, time.Duration, error) {
	gamesInConfigFile, err := getGames(configFileParentDir, false)
	if err != nil {
		return nil, 0, fmt.Errorf("get games failed: %w", err)
	}

	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")
	logFile, err := os.Open(pathToLogFile)
	if err != nil {
		return nil, 0, fmt.Errorf("open log file failed: %w", err)
	}
	defer logFile.Close()

	// Size of the file is needed for the backwards Scanner
	logFileInfo, err := os.Stat(pathToLogFile)
	if err != nil {
		return nil, 0, fmt.Errorf("read log file info failed: %w", err)
	}

	/* Reading log file line by line, starting by the end to the start,
	extracting game name, time spent playing and start and end date
	of the gaming session.

	Example:
	1. "[2025-06-13] Played 'My Game' from 14:26 to 14:27."
	2. "[2025-06-13] Played 'My Game' from 22:30 to 2025-06-14 00:27."

	2 only happens when the day changes while playing a game. */

	games := make(map[string]time.Duration, len(gamesInConfigFile))
	regexGameName := regexp.MustCompile(`\'(.*?)\'`)
	regexDate := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	regexTime := regexp.MustCompile(`\d{2}:\d{2}:\d{2}`)
	backScanner := NewScanner(logFile, int(logFileInfo.Size()))
	timeTwoWeeksAgo := time.Now().UTC().AddDate(0, 0, -14)
	var totalPlaytime time.Duration

	// Turns to false when entry older than two weeks is found.
	continueScanning := true
	for continueScanning {
		line, _, err := backScanner.Line()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, 0, fmt.Errorf("scan line failed: %w", err)
			} else {
				return games, totalPlaytime.Round(time.Second), nil
			}
		}

		if line == "" {
			continue
		}

		datesMatch := regexDate.FindAllString(line, 2)
		gameNameMatch := regexGameName.FindStringSubmatch(line)
		timesMatch := regexTime.FindAllString(line, 2)
		if gameNameMatch[1] == "" || datesMatch[0] == "" ||
			len(timesMatch) != 2 {
			continue
		}

		// If there is a start date and an end date matched by regex.
		dayChanged := len(datesMatch) == 2

		startDateFromLogs := datesMatch[0]
		var endDateFromLogs string
		if dayChanged {
			endDateFromLogs = datesMatch[1]
		}
		startHourFromLogs := timesMatch[0]
		endHourFromLogs := timesMatch[1]

		for _, game := range gamesInConfigFile {
			/* Check if the game extracted from the log file
			is also on the config file, to prevent deleted
			games from showing up in the output*/
			if game.Name != gameNameMatch[1] {
				continue
			}

			startTime, err := getTimeFromStrings(
				startDateFromLogs, startHourFromLogs)
			if err != nil {
				return nil, 0, fmt.Errorf(
					"get time from strings failed: %w",
					err)
			}

			/* Check if start date of entry is in the last two weeks
			If not, stop scanning the logs. Since reading backwards
			shows the most recent log entry, we can stop reading
			the file if we reach an entry that is older than two weeks
			ago.*/
			if timeTwoWeeksAgo.After(startTime) {
				continueScanning = false
				break
			}

			var endTime time.Time
			if dayChanged {
				endTime, err = getTimeFromStrings(endDateFromLogs,
					endHourFromLogs)
				if err != nil {
					return nil, 0, fmt.Errorf(
						"get time from strings failed: %w",
						err)
				}
			} else {
				endTime, err = getTimeFromStrings(startDateFromLogs,
					endHourFromLogs)
				if err != nil {
					return nil, 0, fmt.Errorf(
						"get time from strings failed: %w",
						err)
				}

			}

			playtime := endTime.Sub(startTime)
			games[game.Name] = games[game.Name] + playtime
			totalPlaytime = totalPlaytime + playtime
			break
		}
	}

	return games, totalPlaytime.Round(time.Second), nil
}
