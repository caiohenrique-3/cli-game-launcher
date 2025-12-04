package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Game struct {
	Id       int
	Name     string
	Path     string
	Playtime string
	Hidden   bool
}

// Game with last time played.
type GameLP struct {
	Id         int
	Name       string
	LastPlayed time.Time
}

// Game with playtime in the last two weeks.
type GamePlaytimeL2W struct {
	Id       int
	Name     string
	Playtime time.Duration
}

var ErrInvalidOption = errors.New("invalid option")
var ErrNoGamesFound = errors.New("no games found")
var ErrInvalidGameId = errors.New("invalid game id")

// Creates config.json file in the path specified by dir param.
func createEmptyConfigFileAt(parentDir string) error {
	pathToConfigFile := filepath.Join(parentDir, "config.json")

	games := []Game{}
	configFileData, err := json.MarshalIndent(games, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	err = os.WriteFile(pathToConfigFile, configFileData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

// Saves a new game in the database.
func saveGameToConfig(gameName string, pathToExecutable string,
	configFileParentDir string) error {
	pathToDb := filepath.Join(configFileParentDir, "data.db")
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS games (id INTEGER PRIMARY KEY, name TEXT, path TEXT, playtime TEXT, hidden INTEGER)")
	if err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	_, err = db.Exec("INSERT INTO games (name, path, playtime, hidden) VALUES(?, ?, ?, ?)",
		gameName, pathToExecutable, "0h0m0s", false)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}

// Removes a game from the database.
func removeGameFromConfig(gameId int, configFileParentDir string) error {
	if gameId <= 0 {
		return ErrInvalidOption
	}

	pathToDb := filepath.Join(configFileParentDir, "data.db")
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	// 0 if the table is empty, 1 if it contains any rows.
	isTableEmptyRow, err := db.Query("SELECT EXISTS (SELECT 1 FROM games)")
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer isTableEmptyRow.Close()

	var tableHasRows bool
	for isTableEmptyRow.Next() {
		err := isTableEmptyRow.Scan(&tableHasRows)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
	}

	if !tableHasRows {
		return ErrNoGamesFound
	}

	_, err = db.Exec("DELETE FROM games WHERE id = ?", gameId)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

// Returns all games in the database.
func getGames(configFileParentDir string, excludeHiddenGames bool) ([]Game, error) {
	pathToDb := filepath.Join(configFileParentDir, "data.db")
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	var sb strings.Builder
	sb.WriteString("SELECT * FROM games")
	if excludeHiddenGames {
		sb.WriteString(" WHERE hidden != 1")
	}

	rows, err := db.Query(sb.String())
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	games := []Game{}
	for rows.Next() {
		newGame := Game{}
		err := rows.Scan(&newGame.Id, &newGame.Name, &newGame.Path,
			&newGame.Playtime, &newGame.Hidden)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		games = append(games, newGame)
	}

	return games, nil
}

// Updates playtime of the game in the database.
func saveTimeSpentPlaying(gameIdToUpdate int, timeSpent time.Duration,
	pathToDb string) error {
	if gameIdToUpdate <= 0 {
		return ErrInvalidGameId
	}

	if timeSpent == 0 {
		return nil
	}

	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	row, err := db.Query("SELECT playtime FROM games WHERE id = ?", gameIdToUpdate)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer row.Close()

	var currentPlaytimeStr string
	for row.Next() {
		err := row.Scan(&currentPlaytimeStr)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
	}

	// Calculating new playtime for Game
	currentPlaytime, err := time.ParseDuration(currentPlaytimeStr)
	if err != nil {
		return fmt.Errorf("parse duration failed: %w", err)
	}

	newPlaytime := currentPlaytime + timeSpent
	newPlaytime = newPlaytime.Round(time.Second)

	_, err = db.Exec("UPDATE games SET playtime = ? WHERE id = ?",
		newPlaytime.String(), gameIdToUpdate)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	return nil
}

// Takes the ID of game in the database and flips the value of hidden variable.
func toggleHiddenGame(idOfChosenGame int, configFileParentDir string) error {
	pathToDb := filepath.Join(configFileParentDir, "data.db")
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	row, err := db.Query("SELECT hidden FROM games WHERE id = ?", idOfChosenGame)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer row.Close()

	var isHidden bool
	for row.Next() {
		err := row.Scan(&isHidden)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
	}

	isHidden = !isHidden
	_, err = db.Exec("UPDATE games SET hidden = ? WHERE id = ?",
		isHidden, idOfChosenGame)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	return nil
}

// Saves ID of the game played, the start time and the end time to the database file.
func saveSession(gameId int, timeStart time.Time, timeEnd time.Time,
	pathToDb string) error {
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS sessions (gameId INTEGER, timeStart TEXT, timeEnd TEXT, FOREIGN KEY(gameId) REFERENCES games(id) ON UPDATE CASCADE ON DELETE CASCADE)")
	if err != nil {
		return fmt.Errorf("create table failed: %w", err)
	}

	_, err = db.Exec("INSERT INTO sessions (gameId, timeStart, timeEnd) VALUES(?, ?, ?)",
		gameId, timeStart.Format(time.RFC1123), timeEnd.Format(time.RFC1123))
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	return nil
}

// Returns a list of games with their last played times.
func getGamesWithLastPlayedTime(configFileParentDir string) ([]GameLP, error) {
	pathToDb := filepath.Join(configFileParentDir, "data.db")
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT l.id, l.name, r.timeEnd FROM games l INNER JOIN (SELECT gameId, timeEnd, ROW_NUMBER() OVER (PARTITION BY gameId ORDER BY timeEnd DESC) AS rn FROM sessions) r ON r.gameId = l.id WHERE r.rn = 1")
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var games []GameLP
	for rows.Next() {
		var id int
		var name, timeEndStr string
		err := rows.Scan(&id, &name, &timeEndStr)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		timeEnd, err := time.Parse(time.RFC1123, timeEndStr)
		if err != nil {
			return nil, fmt.Errorf("time parse failed: %w", err)
		}

		games = append(games, GameLP{Id: id, Name: name, LastPlayed: timeEnd})
	}

	return games, nil
}

// Returns a map of games with their playtimes in the last two weeks, alongside
// total playtime across all games in the last two weeks. The map key is the
// game id in the database.
func getGamesWithPlaytimeLastTwoWeeks(configFileParentDir string) (map[int]*GamePlaytimeL2W,
	time.Duration, error) {
	pathToDb := filepath.Join(configFileParentDir, "data.db")
	db, err := sql.Open("sqlite3", pathToDb)
	if err != nil {
		return nil, 0, fmt.Errorf("open db failed: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, timeStart, timeEnd FROM sessions INNER JOIN games ON games.id = sessions.gameId ORDER BY timeEnd ASC")
	if err != nil {
		return nil, 0, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	timeNow := time.Now()
	twoWeeksAgo := timeNow.AddDate(0, 0, -14)
	games := make(map[int]*GamePlaytimeL2W)
	var totalPlaytimeL2W time.Duration

	for rows.Next() {
		var id int
		var name, timeStartStr, timeEndStr string
		err := rows.Scan(&id, &name, &timeStartStr, &timeEndStr)
		if err != nil {
			return nil, 0, fmt.Errorf("scan failed: %w", err)
		}

		timeStart, err := time.Parse(time.RFC1123, timeStartStr)
		if err != nil {
			return nil, 0, fmt.Errorf("time format failed: %w", err)
		}

		// Exit loop early when session is not in the last two weeks
		if twoWeeksAgo.After(timeStart) {
			break
		}

		timeEnd, err := time.Parse(time.RFC1123, timeEndStr)
		if err != nil {
			return nil, 0, fmt.Errorf("time format failed: %w", err)
		}

		playtime := time.Duration.Round(timeEnd.Sub(timeStart), time.Second)

		_, ok := games[id]
		if ok {
			// If game already exists on the map, just update the playtime
			games[id].Playtime += playtime
		} else {
			games[id] = &GamePlaytimeL2W{Id: id, Name: name, Playtime: playtime}
		}

		totalPlaytimeL2W += playtime
	}

	return games, totalPlaytimeL2W, nil
}
