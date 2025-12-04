package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Test_SaveGameToConfig_HappyPath(t *testing.T) {
	dbParentDir := t.TempDir()
	pathToDb := filepath.Join(dbParentDir, "data.db")

	err := saveGameToConfig("Test Game", "Test Path", dbParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		t.Fatalf("open db failed: '%v'\n", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM games")
	if err != nil {
		t.Fatalf("query failed: '%v'\n", err)
	}
	defer rows.Close()

	var (
		gotId                         int
		gotName, gotPath, gotPlaytime string
		isHidden                      bool
	)

	for rows.Next() {
		err := rows.Scan(&gotId, &gotName, &gotPath, &gotPlaytime, &isHidden)
		if err != nil {
			t.Fatalf("scan failed: '%v'\n", err)
		}
	}

	if gotId != 1 {
		t.Fatalf("got: '%d'; want: '%d';\n", gotId, 1)
	}

	if gotName != "Test Game" {
		t.Fatalf("got: '%s'; want: '%s';\n", gotName, "Test Game")
	}

	if gotPath != "Test Path" {
		t.Fatalf("got: '%s'; want: '%s';\n", gotPath, "Test Path")
	}

	if gotPlaytime != "0h0m0s" {
		t.Fatalf("got: '%s'; want: '%s';\n", gotPlaytime, "0h0m0s")
	}

	if isHidden != false {
		t.Fatalf("got: '%t'; want: '%t';\n", isHidden, false)
	}
}

func Test_RemoveGameFromConfig_HappyPath(t *testing.T) {
	dbParentDir := t.TempDir()
	pathToDb := filepath.Join(dbParentDir, "data.db")

	err := saveGameToConfig("Test Game", "Test Path", dbParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	err = removeGameFromConfig(1, dbParentDir)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		t.Fatalf("open db failed: '%v'\n", err)
	}
	defer db.Close()

	gameCount, err := db.Query("SELECT COUNT(*) FROM games")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer gameCount.Close()

	var numGames int
	for gameCount.Next() {
		err := gameCount.Scan(&numGames)
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}
	}

	if numGames != 0 {
		t.Fatalf("got: '%d'; want: '%d';\n", numGames, 0)
	}
}

func Test_RemoveGameFromConfig_NoGamesFound(t *testing.T) {
	dbParentDir := t.TempDir()
	pathToDb := filepath.Join(dbParentDir, "data.db")

	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		t.Fatalf("open db failed: '%v'\n", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS games (id INTEGER PRIMARY KEY, name TEXT, path TEXT, playtime TEXT, hidden INTEGER)")
	if err != nil {
		t.Fatalf("create table failed: '%v'\n", err)
	}

	err = removeGameFromConfig(1, dbParentDir)
	if !errors.Is(err, ErrNoGamesFound) {
		t.Fatalf("got '%v'; want '%v';\n", err, ErrNoGamesFound)
	}
}

func Test_RemoveGameFromConfig_InvalidIndex(t *testing.T) {
	err := removeGameFromConfig(0, "Test Path")
	if !errors.Is(err, ErrInvalidOption) {
		t.Errorf("got '%v'; want '%v';\n", err, ErrInvalidOption)
	}
	err = removeGameFromConfig(-1, "Test Path")
	if !errors.Is(err, ErrInvalidOption) {
		t.Errorf("got '%v'; want '%v';\n", err, ErrInvalidOption)
	}
}

func Test_GetGames_HappyPath(t *testing.T) {
	dbParentDir := t.TempDir()
	pathToDb := filepath.Join(dbParentDir, "data.db")

	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		t.Fatalf("open db failed: '%v'\n", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS games (id INTEGER PRIMARY KEY, name TEXT, path TEXT, playtime TEXT, hidden INTEGER)")
	if err != nil {
		t.Fatalf("create table failed: '%v'\n", err)
	}

	gameName := "Test Game"
	gamePath := "Test Path"
	playtime := "0h0m0s"
	hidden := false

	games := []Game{
		{Id: 1, Name: gameName, Path: gamePath, Playtime: playtime, Hidden: hidden},
		{Id: 2, Name: gameName, Path: gamePath, Playtime: playtime, Hidden: hidden},
		{Id: 3, Name: gameName, Path: gamePath, Playtime: playtime, Hidden: hidden},
		{Id: 4, Name: gameName, Path: gamePath, Playtime: playtime, Hidden: true},
		{Id: 5, Name: gameName, Path: gamePath, Playtime: playtime, Hidden: true},
	}

	transaction, err := db.Begin()
	if err != nil {
		t.Fatalf("begin transaction failed: '%v'\n", err)
	}

	for i, game := range games {
		transaction.Exec("INSERT INTO games (name, path, playtime, hidden) VALUES(?, ?, ?, ?)")
		_, err = transaction.Exec("INSERT INTO games (name, path, playtime, hidden) VALUES(?, ?, ?, ?)",
			game.Name, game.Path, game.Playtime, game.Hidden)
		if err != nil {
			t.Fatalf("insert no.%d failed: '%v'\n", i, err)
		}
	}

	err = transaction.Commit()
	if err != nil {
		t.Fatalf("commit transaction failed: '%v'\n", err)
	}

	excludeHiddenGames := false
	gotGames, err := getGames(dbParentDir, excludeHiddenGames)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	if len(gotGames) != len(games) {
		t.Fatalf("got: '%d'; want: '%d';\n", len(gotGames), len(games))
	}

	if !reflect.DeepEqual(games, gotGames) {
		t.Fatalf("got: '%v'; want: '%v';\n", gotGames, games)
	}

	excludeHiddenGames = true
	nonHiddenGames, err := getGames(dbParentDir, excludeHiddenGames)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	if len(nonHiddenGames) != 3 {
		t.Fatalf("got: '%d'; want: '%d';\n", len(nonHiddenGames), 3)
	}

	for _, game := range nonHiddenGames {
		if game.Hidden {
			t.Fatalf("got: '%t'; want '%t';\n", game.Hidden, false)
		}
	}
}

func Test_SaveTimeSpentPlaying_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()
	err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	gameId := 1
	timeSpent := time.Duration(2) * time.Hour
	pathToDb := filepath.Join(configFileParentDir, "data.db")

	err = saveTimeSpentPlaying(gameId, timeSpent, pathToDb)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	games, err := getGames(configFileParentDir, false)
	if err != nil {
		t.Fatalf("get games failed: '%v'\n", err)
	}

	wantPlaytime := "2h0m0s"
	if games[0].Playtime != wantPlaytime {
		t.Fatalf("got: '%s'; want: '%s';\n", games[0].Playtime, wantPlaytime)
	}
}

func Test_SaveTimeSpentPlaying_InvalidGameId(t *testing.T) {
	err := saveTimeSpentPlaying(0, 0, "")
	if !errors.Is(err, ErrInvalidGameId) {
		t.Fatalf("got: '%v'; want '%v';\n", err, ErrInvalidGameId)
	}
}

func Test_ToggleHiddenGame_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()
	err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	idOfChosenGame := 1 // Sqlite begins with 1
	err = toggleHiddenGame(idOfChosenGame, configFileParentDir)
	if err != nil {
		t.Fatalf("got: '%v'; want nil\n", err)
	}

	excludeHiddenGames := false
	gotGames, err := getGames(configFileParentDir, excludeHiddenGames)
	if err != nil {
		t.Fatalf("get games failed: '%v'\n", err)
	}

	if !gotGames[0].Hidden {
		t.Fatalf("got: '%t'; want: '%t';\n", gotGames[0].Hidden, true)
	}

	err = toggleHiddenGame(idOfChosenGame, configFileParentDir)
	if err != nil {
		t.Fatalf("got: '%v'; want nil\n", err)
	}

	gotGames, err = getGames(configFileParentDir, excludeHiddenGames)
	if err != nil {
		t.Fatalf("get games failed: '%v'\n", err)
	}

	if gotGames[0].Hidden {
		t.Fatalf("got: '%t'; want: '%t';\n", gotGames[0].Hidden, false)
	}
}

func Test_SaveSession_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()
	pathToDb := filepath.Join(configFileParentDir, "data.db")

	err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	timeStart, err := time.Parse(time.RFC1123, "Mon, 02 Jan 2020 15:00:00 UTC")
	if err != nil {
		t.Fatalf("parse failed: '%v'\n", err)
	}

	timeEnd, err := time.Parse(time.RFC1123, "Mon, 02 Jan 2020 15:32:05 UTC")
	if err != nil {
		t.Fatalf("parse failed: '%v'\n", err)
	}

	gameId := 1
	err = saveSession(gameId, timeStart, timeEnd, pathToDb)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		t.Fatalf("open db failed: '%v'\n", err)
	}
	defer db.Close()

	row, err := db.Query("SELECT * FROM sessions WHERE gameId = ?", gameId)
	if err != nil {
		t.Fatalf("query failed: '%v'\n", err)
	}
	defer row.Close()

	var gotId int
	var gotStartStr, gotEndStr string

	for row.Next() {
		err := row.Scan(&gotId, &gotStartStr, &gotEndStr)
		if err != nil {
			t.Fatalf("scan failed: '%v'\n", err)
		}
	}

	gotStart, err := time.Parse(time.RFC1123, gotStartStr)
	if err != nil {
		t.Fatalf("time parse failed: '%v'\n", err)
	}

	gotEnd, err := time.Parse(time.RFC1123, gotEndStr)
	if err != nil {
		t.Fatalf("time parse failed: '%v'\n", err)
	}

	if gotId != gameId {
		t.Fatalf("got: '%d'; want: '%d';\n", gotId, gameId)
	}

	if !gotStart.Equal(timeStart) {
		t.Fatalf("got: '%s'; want: '%s';\n", gotStartStr, timeStart)
	}

	if !gotEnd.Equal(timeEnd) {
		t.Fatalf("got: '%s'; want: '%s';\n", gotEndStr, timeEnd)
	}
}

func Test_GetGamesWithLastPlayedTime_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()
	pathToDb := filepath.Join(configFileParentDir, "data.db")

	for range 2 {
		err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
		if err != nil {
			t.Fatalf("save game failed: '%v'\n", err)
		}
	}

	gameId := 1
	timeStart := time.Now()
	t1 := timeStart.Add(time.Duration(30) * time.Minute)

	err := saveSession(gameId, timeStart, t1, pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	// Only latest session will be retrieved
	t2 := timeStart.Add(time.Duration(2) * time.Hour)
	err = saveSession(gameId, timeStart, t2, pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	t3 := timeStart.Add(time.Duration(4) * time.Hour)
	gameId = 2
	err = saveSession(gameId, timeStart, t3, pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	games, err := getGamesWithLastPlayedTime(configFileParentDir)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	if len(games) != 2 {
		t.Fatalf("got: length '%d'; want length '%d';\n", len(games), 2)
	}

	gotTime1 := games[0].LastPlayed.Format(time.RFC1123)
	wantTime1 := t2.Format(time.RFC1123)
	if gotTime1 != wantTime1 {
		t.Fatalf("got: '%s'; want '%s';\n", gotTime1, wantTime1)
	}

	gotTime2 := games[1].LastPlayed.Format(time.RFC1123)
	wantTime2 := t3.Format(time.RFC1123)
	if gotTime2 != wantTime2 {
		t.Fatalf("got: '%s'; want '%s';\n", gotTime2, wantTime2)
	}
}

func Test_GetGamesWithPlaytimeLastTwoWeeks_HappyPath(t *testing.T) {
	configFileParentDir := t.TempDir()
	pathToDb := filepath.Join(configFileParentDir, "data.db")

	err := saveGameToConfig("Test Game", "Test Path", configFileParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	timeNow := time.Now()
	time20DaysAgo := timeNow.AddDate(0, 0, -20).Add(1 * time.Hour)
	time20DaysAgoEnd := time20DaysAgo.Add(3 * time.Hour)
	timeYesterday := timeNow.AddDate(0, 0, -1)
	timeYesterdayEnd := timeYesterday.Add(30 * time.Minute)

	err = saveSession(1, time20DaysAgo, time20DaysAgoEnd, pathToDb)
	if err != nil {
		t.Fatalf("save session failed: '%v'\n", err)
	}

	for range 2 {
		err = saveSession(1, timeYesterday, timeYesterdayEnd, pathToDb)
		if err != nil {
			t.Fatalf("save session failed: '%v'\n", err)
		}
	}

	games, totalPlaytimeL2W, err := getGamesWithPlaytimeLastTwoWeeks(configFileParentDir)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	if len(games) != 1 {
		t.Fatalf("got length '%d'; want length '%d';\n", len(games), 1)
	}

	wantPlaytime := "1h0m0s"
	if totalPlaytimeL2W.String() != wantPlaytime {
		t.Fatalf("got: '%s'; want: '%s';\n", totalPlaytimeL2W, wantPlaytime)
	}

	gotPlaytime := games[1].Playtime.String()
	if gotPlaytime != wantPlaytime {
		t.Fatalf("got: '%s'; want: '%s';\n", gotPlaytime, wantPlaytime)
	}
}
