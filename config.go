package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Game struct {
	Name             string `json:"name"`
	PathToExecutable string `json:"pathToExecutable"`
	TimeSpentPlaying string `json:"timeSpentPlaying"`
	IsHidden         bool   `json:"isHidden"`
}

var ErrInvalidOption = errors.New("invalid option")
var ErrNoGamesFound = errors.New("no games found")

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

// Loads config.json and appends a new Game entry to the file.
// Creates a config.json file if it doesn't already exist.
func saveGameToConfig(gameName string, pathToExecutable string,
	configFileParentDir string) error {
	configFile := filepath.Join(configFileParentDir, "config.json")

	if !fileExists(configFile) {
		err := createEmptyConfigFileAt(configFileParentDir)
		if err != nil {
			return fmt.Errorf("create empty config file failed: %w", err)
		}
	}

	configFileData, err := loadConfigFile(configFileParentDir)
	if err != nil {
		return fmt.Errorf("load config file failed: %w", err)
	}

	game := &Game{
		Name:             gameName,
		PathToExecutable: pathToExecutable,
		TimeSpentPlaying: "0h0m0s",
	}

	err = appendNewGameToConfigFile(configFileData, game, configFile)
	if err != nil {
		return fmt.Errorf("append new game failed: %w", err)
	}

	return nil
}

// Unmarshals fileData, appends a new game to it and saves the
// new file in pathToConfigFile.
func appendNewGameToConfigFile(configFileData []byte, gameToAdd *Game,
	pathToConfigFile string) error {
	games := []Game{}
	err := json.Unmarshal(configFileData, &games)
	if err != nil {
		return fmt.Errorf("old config file unmarshal failed: %w", err)
	}

	games = append(games, *gameToAdd)

	newConfigFileData, err := json.MarshalIndent(games, "", "    ")
	if err != nil {
		return fmt.Errorf("new config file marshal failed: %w", err)
	}

	err = os.WriteFile(pathToConfigFile, newConfigFileData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write file after append failed: %w", err)
	}

	return nil
}

// Removes the selected entry from config.json file.
func removeGameFromConfig(indexToRemove int, configFileParentDir string) error {
	configFileData, err := loadConfigFile(configFileParentDir)
	if err != nil {
		return fmt.Errorf("load config file failed: %w", err)
	}

	games := []Game{}
	err = json.Unmarshal(configFileData, &games)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	if len(games) == 0 {
		return ErrNoGamesFound
	}

	if indexToRemove > len(games)-1 || indexToRemove < 0 {
		return ErrInvalidOption
	}

	// Create a new slice of games without the game in index picked by user.
	newGames := []Game{}
	for i, game := range games {
		if i != indexToRemove {
			newGames = append(newGames, game)
		}
	}

	newConfigFileData, err := json.MarshalIndent(newGames, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal after removal failed: %w", err)
	}

	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	err = os.WriteFile(pathToConfigFile, newConfigFileData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("save config file after removal failed: %w", err)
	}

	fmt.Printf("'%s' removed.\n", games[indexToRemove].Name)

	return nil
}

// Reads file "config.json" inside the parent dir.
func loadConfigFile(configFileParentDir string) ([]byte, error) {
	configFile := filepath.Join(configFileParentDir, "config.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}
	return data, nil
}

// Scans a config file for games and returns them in the form of a slice.
func getGames(configFileParentDir string, excludeHiddenGames bool) ([]Game, error) {
	data, err := loadConfigFile(configFileParentDir)
	if err != nil {
		return nil, fmt.Errorf("load config file failed: %w", err)
	}

	games := []Game{}
	err = json.Unmarshal(data, &games)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if excludeHiddenGames {
		newGames := make([]Game, len(games))

		for _, game := range games {
			if game.IsHidden {
				continue
			}
			newGames = append(newGames, game)
		}
		return newGames, nil
	}

	return games, nil
}

// Called after the game is finished running,
// updates playtime of the game in the config file.
func saveTimeSpentPlayingToConfig(games []Game, gameIndexToUpdate int,
	timeSpentPlaying time.Duration, pathToConfigFile string) error {
	// Getting new time spent playing
	currentTimeSpentPlayingString := games[gameIndexToUpdate].TimeSpentPlaying
	currentTimeSpentPlayingDuration, err := time.
		ParseDuration(currentTimeSpentPlayingString)
	if err != nil {
		return fmt.Errorf("parse duration failed: %w", err)
	}

	newTimeSpentPlayingDuration := currentTimeSpentPlayingDuration + timeSpentPlaying
	newTimeSpentPlayingDuration = newTimeSpentPlayingDuration.Round(time.Second)

	// Update it in the slice
	games[gameIndexToUpdate].TimeSpentPlaying = newTimeSpentPlayingDuration.String()

	// Save updated data to config file
	newConfigFileData, err := json.MarshalIndent(games, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal new config file data failed: %w", err)
	}

	err = os.WriteFile(pathToConfigFile, newConfigFileData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write to config file failed: %w", err)
	}

	return nil
}

func toggleHiddenGame(indexOfChosenGame int, pathToConfigFile string) error {
	configFileParentDir := filepath.Dir(pathToConfigFile)
	excludeHiddenGames := false
	games, err := getGames(configFileParentDir, excludeHiddenGames)
	if err != nil {
		return fmt.Errorf("get games failed: %w", err)
	}

	if games[indexOfChosenGame].IsHidden {
		games[indexOfChosenGame].IsHidden = false
	} else {
		games[indexOfChosenGame].IsHidden = true
	}

	newConfigFileData, err := json.MarshalIndent(games, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	err = os.WriteFile(pathToConfigFile, newConfigFileData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}
