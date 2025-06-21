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
		Sprintf("[%s] Played 'Test Game 1' from 08:00 to 12:00.\n",
			timeNow.Format(time.DateOnly))
	logEntry8HoursPlayed := fmt.
		Sprintf("[%s] Played 'Test Game 2' from 08:00 to 16:00.\n",
			timeNow.Format(time.DateOnly))
	logEntry24HoursPlayed := fmt.
		Sprintf("[%s] Played 'Test Game 3' from 08:00 to %s 08:00.\n",
			timeNow.Format(time.DateOnly),
			time1DayLater.Format(time.DateOnly))
	logEntryRemovedGame := fmt.
		Sprintf("[%s] Played 'Test Game 4' from 17:30 to 19:00.\n",
			timeNow.Format(time.DateOnly))
	logEntryMoreThanTwoWeeksAgo := fmt.
		Sprintf("[%s] Played 'Test Game 5' from 11:29 to 15:01.\n",
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
