package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Game struct {
	Name             string `json:"name"`
	PathToExecutable string `json:"pathToExecutable"`
}

var ErrInvalidOption = errors.New("invalid option")
var ErrNoGamesFound = errors.New("no games found")

// Creates config.json file in the path specified by dir param.
func createEmptyConfigFileAt(dir string) error {
	configFile := filepath.Join(dir, "config.json")

	arr := [...]Game{}
	jsonData, err := json.MarshalIndent(arr, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configFile, jsonData, os.ModePerm)
	if err != nil {
		return err
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
			return err
		}
	}

	f, err := loadConfigFile(configFileParentDir)
	if err != nil {
		return err
	}

	game := &Game{
		Name:             gameName,
		PathToExecutable: pathToExecutable,
	}

	err = appendNewGameToConfigFile(f, game, configFile)
	if err != nil {
		return err
	}

	return nil
}

// Unmarshals fileData, appends a new game to it and saves the
// new file in pathToConfigFile.
func appendNewGameToConfigFile(fileData []byte, game *Game,
	pathToConfigFile string) error {
	dat := []Game{}
	err := json.Unmarshal(fileData, &dat)
	if err != nil {
		return err
	}

	dat = append(dat, *game)

	b, err := json.MarshalIndent(dat, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(pathToConfigFile, b, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// Removes the selected entry from config.json file.
func removeGameFromConfig(index int, configFileParentDir string) error {
	configFile := filepath.Join(configFileParentDir, "config.json")
	games := []Game{}
	data, err := loadConfigFile(configFileParentDir)
	if err != nil {
		return fmt.Errorf("load config file failed: %w", err)
	}

	err = json.Unmarshal(data, &games)
	if err != nil {
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	if len(games) == 0 {
		return ErrNoGamesFound
	}

	if index > len(games)-1 || index < 0 {
		return ErrInvalidOption
	}

	newGames := []Game{}
	for i, k := range games {
		if i != index {
			newGames = append(newGames, k)
		}
	}

	jsonData, err := json.MarshalIndent(newGames, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal after removal failed: %w", err)
	}

	err = os.WriteFile(configFile, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("save config file after removal failed: %w", err)
	}

	return nil
}

func loadConfigFile(configFileParentDir string) ([]byte, error) {
	configFile := filepath.Join(configFileParentDir, "config.json")
	if !fileExists(configFile) {
		return nil, ErrDoesNotExistOrIsADirectory
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}
	return data, nil
}

// Scans a config file for games and returns them in the form of a slice.
func getGames(configFileParentDir string) ([]Game, error) {
	data, err := loadConfigFile(configFileParentDir)
	if err != nil {
		return nil, fmt.Errorf("load config file failed: %w", err)
	}

	games := []Game{}
	err = json.Unmarshal(data, &games)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return games, nil
}
