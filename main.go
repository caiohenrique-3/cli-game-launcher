package main

import (
	"log"
	"os"
	"path/filepath"
)

// Directory "cli-game-launcher" inside the OS user config dir.
var programHome string

func init() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
		return
	}

	programHome = filepath.Join(userConfigDir, "cli-game-launcher")
	err = os.MkdirAll(programHome, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
		return
	}
}

func main() {
	err := handleCommandLine()
	if err != nil {
		log.Fatalln(err)
	}
}
