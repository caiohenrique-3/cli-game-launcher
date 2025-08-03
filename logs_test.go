package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_SaveSessionToLogs_HappyPath(t *testing.T) {
	setupTestsDir(t)
	pathToLogFile := createTempFileForTests(t)

	timeStart, err := time.Parse(time.DateTime, "2020-01-03 15:00:00")
	if err != nil {
		t.Errorf("error parsing start time: %v\n", err)
	}
	timeEnd, err := time.Parse(time.DateTime, "2020-01-03 15:32:05")
	if err != nil {
		t.Errorf("error parsing start end: %v\n", err)
	}
	gameName := "test"

	err = saveSessionToLogs(gameName, timeStart, timeEnd, pathToLogFile)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	logFileData, err := os.ReadFile(pathToLogFile)
	if err != nil {
		t.Errorf("error reading file: %v\n", err)
	}

	wantString := fmt.Sprintf("[%s] Played '%s' from %s to %s.\n",
		timeStart.Format(time.DateOnly),
		gameName, timeStart.Format(time.TimeOnly),
		timeEnd.Format(time.TimeOnly))

	if string(logFileData) != wantString {
		t.Errorf("got: '%s'; want: '%s';\n", string(logFileData), wantString)
	}
}

func Test_SaveSessionToLogs_AppendsCorrectly(t *testing.T) {
	setupTestsDir(t)
	pathToLogFile := createTempFileForTests(t)

	timeStart, err := time.Parse(time.DateTime, "2015-02-03 15:00:00")
	if err != nil {
		t.Errorf("error parsing start time: %v\n", err)
	}
	timeEnd, err := time.Parse(time.DateTime, "2015-02-03 15:32:04")
	if err != nil {
		t.Errorf("error parsing start end: %v\n", err)
	}
	gameName := "test"

	for i := range 2 {
		err = saveSessionToLogs(gameName, timeStart, timeEnd, pathToLogFile)
		if err != nil {
			t.Errorf("[%d] got: '%v'; want nil;\n", i, err)
		}
	}

	logFileData, err := os.ReadFile(pathToLogFile)
	if err != nil {
		t.Errorf("error reading file: %v\n", err)
	}

	timeString := fmt.Sprintf("[%s] Played '%s' from %s to %s.\n",
		timeStart.Format(time.DateOnly),
		gameName,
		timeStart.Format(time.TimeOnly),
		timeEnd.Format(time.TimeOnly))

	var sb strings.Builder
	// 2x because we call saveSessionToLogs two times with same input.
	sb.WriteString(timeString)
	sb.WriteString(timeString)
	wantString := sb.String()

	if string(logFileData) != wantString {
		t.Errorf("got: '%s'; want: '%s';\n", string(logFileData), wantString)
	}
}

func Test_SaveSessionToLogs_DisplaysDateIfEndTimeDayIsDifferent(t *testing.T) {
	setupTestsDir(t)
	pathToLogFile := createTempFileForTests(t)

	timeStart, err := time.Parse(time.DateTime, "2006-01-02 15:00:00")
	if err != nil {
		t.Errorf("error parsing start time: %v\n", err)
	}
	timeEnd, err := time.Parse(time.DateTime, "2012-03-05 15:32:05")
	if err != nil {
		t.Errorf("error parsing start end: %v\n", err)
	}
	gameName := "test"

	err = saveSessionToLogs(gameName, timeStart, timeEnd, pathToLogFile)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	logFileData, err := os.ReadFile(pathToLogFile)
	if err != nil {
		t.Errorf("error reading file: %v\n", err)
	}

	wantString := fmt.Sprintf("[%s] Played '%s' from %s to %s.\n",
		timeStart.Format(time.DateOnly),
		gameName,
		timeStart.Format(time.TimeOnly),
		timeEnd.Format(time.DateTime))

	if string(logFileData) != wantString {
		t.Errorf("got: '%s'; want: '%s';\n", string(logFileData), wantString)
	}
}

func Test_GetGamesWithLastPlayedTime_HappyPath(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToTempFile := createTempFileForTests(t)
	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")

	// Saving some games before calling the function because
	// it skips games that are not in the config file.
	for i := 1; i < 3; i++ {
		err := saveGameToConfig("Test Game "+strconv.Itoa(i),
			pathToTempFile, configFileParentDir)
		if err != nil {
			t.Errorf("[%d] error saving game to config: %v\n", i, err)
		}
	}

	time6DaysAgo := time.Now().UTC().AddDate(0, 0, -6)
	timeToday := time.Now().UTC()

	time6DaysAgoString := fmt.
		Sprintf("[%s] Played '%s' from 06:00:00 to 06:30:00.\n",
			time6DaysAgo.Format(time.DateOnly),
			"Test Game 1")
	timeTodayString := fmt.
		Sprintf("[%s] Played '%s' from 12:30:00 to 14:00:00.\n",
			timeToday.Format(time.DateOnly),
			"Test Game 2")
	removedGameString := fmt.
		Sprintf("[%s] Played '%s' from 12:00:00 to 12:30:00.\n",
			timeToday.Format(time.DateOnly),
			"Test Game 5")

	var sb strings.Builder
	// Not checking for sb errors here because the test will fail
	// if these are not right anyway.
	sb.WriteString(time6DaysAgoString)
	sb.WriteString(timeTodayString)
	sb.WriteString(removedGameString)

	logFileDataString := sb.String()

	err := os.WriteFile(pathToLogFile, []byte(logFileDataString), os.ModePerm)
	if err != nil {
		t.Errorf("error writing to log file: %v\n", err)
	}

	games, err := getGamesWithLastPlayedTime(configFileParentDir)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}
	if len(games) != 2 {
		t.Errorf("got: '%d'; want '%d';\n", len(games), 2)
	}

	for game := range games {
		if game == "Test Game 5" {
			t.Errorf("got: '%s'; must NOT contain: '%s'\n",
				game, game)
		}
	}
}

func Test_GetGamesWithLastPlayed_SkipsEmptyNamesAndDates(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")

	emptyLogEntryString := "[] Played '' from 12:00 to 12:30.\n"

	err := os.WriteFile(pathToLogFile, []byte(emptyLogEntryString), os.ModePerm)
	if err != nil {
		t.Errorf("error writing to log file: %v\n", err)
	}

	games, err := getGamesWithLastPlayedTime(configFileParentDir)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}
	if len(games) != 0 {
		t.Errorf("got: '%d'; want: '%d';\n", len(games), 0)
	}
}

func Test_GetGamesWithPlaytimeLastTwoWeeks_HappyPath(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")
	pathToTempFile := createTempFileForTests(t)

	// Games that are not in the config file are skipped,
	// so we need to save some games before calling the function.
	for i := 1; i < 6; i++ {
		if i == 4 { // Skipping "Test Game 4"
			continue
		}
		err := saveGameToConfig(
			"Test Game "+strconv.Itoa(i),
			pathToTempFile,
			configFileParentDir)
		if err != nil {
			t.Errorf("[%d] error saving game to config: %v\n", i, err)
		}
	}

	// These dates have to be in the last two weeks.
	timeNow := time.Now().UTC()
	time1DayLater := timeNow.AddDate(0, 0, 1)
	time20DaysAgo := timeNow.AddDate(0, 0, -20)

	logEntry4HoursPlayed := fmt.
		Sprintf("[%s] Played 'Test Game 1' from 08:00:00 to 12:00:00.\n",
			timeNow.Format(time.DateOnly))
	logEntry8HoursPlayed := fmt.
		Sprintf("[%s] Played 'Test Game 2' from 08:00:00 to 16:00:00.\n",
			timeNow.Format(time.DateOnly))
	logEntry24HoursPlayed := fmt.
		Sprintf("[%s] Played 'Test Game 3' from 08:00:00 to %s 08:00:00.\n",
			timeNow.Format(time.DateOnly),
			time1DayLater.Format(time.DateOnly))
	logEntryRemovedGame := fmt.
		Sprintf("[%s] Played 'Test Game 4' from 17:30:00 to 19:00:00.\n",
			timeNow.Format(time.DateOnly))
	logEntryMoreThanTwoWeeksAgo := fmt.
		Sprintf("[%s] Played 'Test Game 5' from 11:29:00 to 15:01:00.\n",
			time20DaysAgo.Format(time.DateOnly))

	// Not checking for write errors, if this fails the test will fail too.
	var sb strings.Builder
	sb.WriteString(logEntryMoreThanTwoWeeksAgo)
	sb.WriteString(logEntry4HoursPlayed)
	sb.WriteString(logEntry8HoursPlayed)
	sb.WriteString(logEntry24HoursPlayed)
	sb.WriteString(logEntryRemovedGame)

	err := os.WriteFile(pathToLogFile, []byte(sb.String()), os.ModePerm)
	if err != nil {
		t.Errorf("error writing to log file: %v\n", err)
	}

	games, totalPlaytime, err := getGamesWithPlaytimeLastTwoWeeks(configFileParentDir)
	if err != nil {
		t.Errorf("got :'%v'; want nil;\n", err)
	}
	if totalPlaytime.String() != "36h0m0s" {
		t.Errorf("got: '%s'; want: '%s';\n",
			totalPlaytime, "36h0m0s")
	}

	wantDurations := map[string]string{
		"Test Game 1": "4h0m0s",
		"Test Game 2": "8h0m0s",
		"Test Game 3": "24h0m0s",
	}

	for game, duration := range games {
		if game == "Test Game 4" || game == "Test Game 5" {
			t.Errorf("got: '%s'; must NOT contain: '%s';\n",
				game, "Test Game 4 || Test Game 5")
		}
		if wantDurations[game] != duration.String() {
			t.Errorf("got: '%s'; want: '%s';\n",
				duration, wantDurations[game])
		}
	}
}

func Test_GetGamesWithPlaytimeLastTwoWeeks_SkipsEmptyNamesAndDates(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")

	emptyLogEntryString := "[] Played '' from 12:00 to 12:30.\n"

	err := os.WriteFile(pathToLogFile, []byte(emptyLogEntryString), os.ModePerm)
	if err != nil {
		t.Errorf("error writing to log file: %v\n", err)
	}

	games, _, err := getGamesWithPlaytimeLastTwoWeeks(configFileParentDir)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}
	if len(games) != 0 {
		t.Errorf("got: '%d'; want: '%d';\n", len(games), 0)
	}
}
