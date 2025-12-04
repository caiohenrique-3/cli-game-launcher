package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_AddGamePrompt(t *testing.T) {
	configFileParentDir := t.TempDir()
	tempFile, err := os.CreateTemp(configFileParentDir, "file")
	if err != nil {
		t.Fatalf("create temp file failed: '%v';\n", err)
	}
	defer tempFile.Close()

	// Changing global variable from real user config directory to
	// OS temp directory
	origProgramHome := programHome
	programHome = configFileParentDir
	defer func() { programHome = origProgramHome }()
	defer quietOutput()()

	userInput := fmt.Sprintf("Test Game\n%s\n", tempFile.Name())

	r := bufio.NewReader(strings.NewReader(userInput))
	err = addGamePrompt(r)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	games, err := getGames(configFileParentDir, false)
	if err != nil {
		t.Fatalf("get games failed: '%v'\n", err)
	}

	if games[0].Id != 1 {
		t.Fatalf("got: '%d'; want: '%d';\n", games[0].Id, 1)
	}
	if games[0].Name != "Test Game" {
		t.Fatalf("got: '%s'; want: '%s';\n", games[0].Name, "Test Game")
	}
	if games[0].Path != tempFile.Name() {
		t.Fatalf("got: '%s'; want: '%s';\n", games[0].Path, tempFile.Name())
	}
	if games[0].Playtime != "0h0m0s" {
		t.Fatalf("got: '%s'; want: '%s';\n", games[0].Playtime, "0h0m0s")
	}
	if games[0].Hidden {
		t.Fatalf("got: '%t'; want: '%t';\n", games[0].Hidden, false)
	}
}

func Test_RemoveGamePrompt(t *testing.T) {
	configFileParentDir := t.TempDir()

	for i := range 2 {
		err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
		if err != nil {
			t.Fatalf("save game no.%d failed: '%v'\n", i, err)
		}
	}

	// Changing global variable from real user config directory to
	// OS temp directory
	origProgramHome := programHome
	programHome = configFileParentDir
	defer func() {
		programHome = origProgramHome
	}()
	defer quietOutput()()

	// Deleting first Game in the slice, with id = 1.
	userInput := bufio.NewReader(strings.NewReader("0\n"))
	err := removeGamePrompt(userInput)
	if err != nil {
		t.Fatalf("got '%v'; want nil;\n", err)
	}

	games, err := getGames(configFileParentDir, false)
	if err != nil {
		t.Fatalf("get games failed: '%v'\n", err)
	}

	if len(games) != 1 {
		t.Fatalf("got length '%d'; want length '%d';\n", len(games), 1)
	}

	if games[0].Id != 2 {
		t.Fatalf("got id '%d'; want id '%d';\n", games[0].Id, 2)
	}
}

func Test_ListGamesNumbered_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()

	for i := range 2 {
		err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
		if err != nil {
			t.Fatalf("save game no.%d failed: '%v'\n", i, err)
		}
	}

	err := toggleHiddenGame(2, configFileParentDir)
	if err != nil {
		t.Fatalf("toggle hidden game failed: '%v'\n", err)
	}

	excludeHiddenGames := false
	games, err := getGames(configFileParentDir, excludeHiddenGames)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	var b bytes.Buffer
	err = listGamesNumbered(games, &b)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	capturedOutput := b.Bytes()
	wantOutput := []byte("[0] Test Game\n[1] Test Game [HIDDEN]\n")
	if !bytes.Equal(capturedOutput, wantOutput) {
		t.Fatalf("got: '%s'; want '%s';\n", capturedOutput, wantOutput)
	}
}

func Test_HideGamePrompt(t *testing.T) {
	configFileParentDir := t.TempDir()

	err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	// Changing global variable from real user config directory to
	// OS temp directory
	origProgramHome := programHome
	programHome = configFileParentDir
	defer func() {
		programHome = origProgramHome
	}()
	defer quietOutput()()

	// Index in the slice, not Game.Id
	userInput := bufio.NewReader(strings.NewReader("0\n"))
	err = hideGamePrompt(userInput)
	if err != nil {
		t.Fatalf("got '%v'; want nil;\n", err)
	}

	games, err := getGames(configFileParentDir, false)
	if err != nil {
		t.Fatalf("get games failed: '%v'\n", err)
	}

	if !games[0].Hidden {
		t.Fatalf("got '%t'; want '%t';\n", games[0].Hidden, true)
	}
}

func Test_GetIntFromUser(t *testing.T) {
	configFileParentDir := t.TempDir()

	for range 6 {
		err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
		if err != nil {
			t.Fatalf("save game failed: '%v'\n", err)
		}
	}

	defer quietOutput()()

	tests := map[string]struct {
		input         string
		wantReturnVal int
		wantErr       error
	}{
		"empty string": {
			input:   "\n",
			wantErr: strconv.ErrSyntax},
		"empty string 2": {
			input:   "\n\n",
			wantErr: strconv.ErrSyntax},
		"happy path": {
			input:         "4\n",
			wantReturnVal: 4,
			wantErr:       nil},
		"input has spaces": {
			input:         "2 \n",
			wantReturnVal: 2,
			wantErr:       nil},
		"input has spaces 2": {
			input:         " 2\n",
			wantReturnVal: 2,
			wantErr:       nil},
		"input is negative": {
			input:         "-2\n",
			wantReturnVal: -2,
			wantErr:       nil},
		"input has emojis": {
			input:   "😀\n",
			wantErr: strconv.ErrSyntax},
		"input has cjk": {
			input:   "접는 사람 テスト 試験 史诗般的\n",
			wantErr: strconv.ErrSyntax},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(test.input))
			userInput, err := getIntFromUser(r)

			if !errors.Is(err, test.wantErr) {
				t.Errorf("got '%v'; want '%v';\n", err, test.wantErr)
			}
			if userInput != test.wantReturnVal {
				t.Errorf("got '%v'; want '%v';\n", userInput, test.wantReturnVal)
			}
		})
	}
}

func Test_ListGamesWithPlaytime_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()

	pathToDb := filepath.Join(configFileParentDir, "data.db")
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		t.Fatalf("open db failed: '%v'\n", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS games (id INTEGER PRIMARY KEY, name TEXT, path TEXT, playtime TEXT, hidden INTEGER)")
	if err != nil {
		t.Fatalf("create table failed: '%v'\n", err)
	}

	_, err = db.Exec("INSERT INTO games (name, path, playtime, hidden) VALUES(?, ?, ?, ?)",
		"Test Game", "Test Path", "2h0m0s", false)
	if err != nil {
		t.Fatalf("query failed: '%v'\n", err)
	}

	var b bytes.Buffer
	err = listGamesWithPlaytime(configFileParentDir, &b)
	if err != nil {
		t.Fatalf("got '%v'; want nil;\n", err)
	}

	gotOutput := b.String()
	if !strings.Contains(gotOutput, "Test Game") {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, "Test Game")
	}
	if !strings.Contains(gotOutput, "2h0m0s") {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, "2h0m0s")
	}
}

func Test_ListGamesWithLastPlayed_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()
	pathToDb := filepath.Join(configFileParentDir, "data.db")

	for range 2 {
		err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
		if err != nil {
			t.Fatalf("save game failed: '%v'\n", err)
		}
	}

	time6DaysAgo := time.Now().AddDate(0, 0, -6).Add(time.Duration(-20) * time.Minute)
	time25MinutesAgo := time.Now().Add(time.Duration(-25) * time.Minute)

	err := saveSession(1, time25MinutesAgo, time25MinutesAgo, pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	err = saveSession(2, time6DaysAgo, time6DaysAgo, pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	var b bytes.Buffer
	err = listGamesWithLastPlayed(configFileParentDir, &b)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	// Won't check for the "(x ago)" part
	gotOutput := b.String()
	wantOutput1 := "Today"
	wantOutput2 := time6DaysAgo.Format(time.DateOnly)

	if !strings.Contains(gotOutput, wantOutput1) {
		t.Errorf("got: \n'%s'; must contain: '%s';\n", gotOutput, wantOutput1)
	}
	if !strings.Contains(gotOutput, wantOutput2) {
		t.Errorf("got: \n'%s'; must contain: '%s';\n", gotOutput, wantOutput2)
	}
}

func Test_ListGamesWithPlaytimeLastTwoWeeks_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()
	pathToDb := filepath.Join(configFileParentDir, "data.db")

	err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	err = saveGameToConfig("Test Game 2", "Test Path", configFileParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	timeNow := time.Now()
	time1 := timeNow.AddDate(0, 0, -3)
	time1End := time1.Add(4 * time.Hour)
	time2 := timeNow.AddDate(0, 0, -6)
	time2End := time2.Add(6*time.Hour + 34*time.Second)

	err = saveSession(1, time1, time1End, pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	err = saveSession(1, time2, time2End, pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	err = saveSession(2, time2, time2.Add(2*time.Hour), pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	var b bytes.Buffer
	err = listGamesWithPlaytimeLastTwoWeeks(configFileParentDir, &b)
	if err != nil {
		t.Fatalf("got '%v'; want nil;\n", err)
	}

	gotOutput := b.String()
	totalHours := "12h0m34s spent playing last two weeks."
	if !strings.Contains(gotOutput, totalHours) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, totalHours)
	}

	playtime1 := "10h0m34s (83.3%)"
	if !strings.Contains(gotOutput, playtime1) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, playtime1)
	}

	playtime2 := "2h0m0s (16.7%)"
	if !strings.Contains(gotOutput, playtime2) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, playtime2)
	}
}
