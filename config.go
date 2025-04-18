package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type Game struct {
	Name             string `json:"name"`
	PathToExecutable string `json:"pathToExecutable"`
}

// Creates config.json file in the path specified by dir param.
func createEmptyConfigFileAt(dir string) {
	configFile := filepath.Join(dir, "config.json")

	arr := [...]Game{}
	jsonData, err := json.MarshalIndent(arr, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(configFile, jsonData, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return
	}
}

// Loads config.json and appends a new Game entry to the file.
// Accepts the game display name, the absolute path to the
// game executable file and the path where config.json will be saved.
func saveGameToConfig(gameName string, pathToExecutable string,
	configFileParentDir string) {
	configFile := filepath.Join(configFileParentDir, "config.json")

	f, err := os.ReadFile(configFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			createEmptyConfigFileAt(configFileParentDir)
		} else {
			log.Fatal(err)
		}
	}

	game := &Game{
		Name:             gameName,
		PathToExecutable: pathToExecutable,
	}

	// File is empty
	// TODO: Extract this
	if len(f) == 0 {
		arr := [1]*Game{game}
		jsonData, err := json.MarshalIndent(arr, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile(configFile, jsonData, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		dat := []Game{}
		err := json.Unmarshal(f, &dat)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("[DEBUG] dat before:", dat)

		newDat := append(dat, *game)

		fmt.Println("[DEBUG] dat after:", newDat)

		b, err := json.MarshalIndent(newDat, "", "    ")
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(configFile, b, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
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
