package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Game struct {
	Name             string `json:"name"`
	PathToExecutable string `json:"pathToExecutable"`
}

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
// Accepts the game display name, the absolute path to the
// game executable file and the path where config.json will be saved.
func saveGameToConfig(gameName string, pathToExecutable string,
	configFileParentDir string) {
	configFile := filepath.Join(configFileParentDir, "config.json")

	if !fileExists(configFile) {
		err := createEmptyConfigFileAt(configFileParentDir)
		if err != nil {
			// return err
		}
	}

	f := loadConfigFile(configFileParentDir)

	game := &Game{
		Name:             gameName,
		PathToExecutable: pathToExecutable,
	}

	appendNewGameToConfigFile(f, game, configFile)
}

// Unmarshals fileData, appends a new game to it and saves the
// new file in pathToConfigFile.
func appendNewGameToConfigFile(fileData []byte, game *Game,
	pathToConfigFile string) {
	dat := []Game{}
	err := json.Unmarshal(fileData, &dat)
	if err != nil {
		log.Fatal(err)
	}

	dat = append(dat, *game)

	b, err := json.MarshalIndent(dat, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(pathToConfigFile, b, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

// Removes the selected entry from config.json file.
func removeGameFromConfig(index int, configFileParentDir string) {
	configFile := filepath.Join(configFileParentDir, "config.json")
	games := []Game{}
	json.Unmarshal(loadConfigFile(configFileParentDir), &games)

	if len(games) == 0 {
		log.Fatal("No games found!")
	}

	if index > len(games)-1 || index < 0 {
		log.Fatal("Invalid option!")
	}

	newGames := []Game{}
	for i, k := range games {
		if i != index {
			newGames = append(newGames, k)
		}
	}

	jsonData, err := json.MarshalIndent(newGames, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(configFile, jsonData, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func loadConfigFile(configFileParentDir string) []byte {
	configFile := filepath.Join(configFileParentDir, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	return data
}
