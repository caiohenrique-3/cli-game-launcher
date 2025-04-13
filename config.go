package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Game struct {
	Name             string        `json:"name"`
	PathToExecutable string        `json:"pathToExecutable"`
	IsInstalled      bool          `json:"isInstalled"`
	Playtime         time.Duration `json:"playtime"`
	LastPlayed       time.Time     `json:"lastPlayed"`
}

// Loads config.json and appends a new Game entry to the file.
// TODO: Tests for saveGameToConfig()
func saveGameToConfig(gameName string, pathToExecutable string) {
	configFile := filepath.Join(programHome, "config.json")

	f, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		Name:             gameName,
		PathToExecutable: pathToExecutable,
		IsInstalled:      true}

	// File is empty
	if len(f) == 0 {
		arr := [1]*Game{game}
		jsonData, err := json.Marshal(arr)
		if err != nil {
			log.Fatal(err)
			return
		}
		err = os.WriteFile(configFile, jsonData, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return
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

		b, err := json.Marshal(newDat)
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(configFile, b, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func loadGameFromConfig() {
	configFile := filepath.Join(programHome, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(string(data))
}
