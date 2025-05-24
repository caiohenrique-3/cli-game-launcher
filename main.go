package main

import (
	"log"
	"os"
	"path/filepath"
)

var programHome string

func init() {
	dirname, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
		return
	}

	programHome = filepath.Join(dirname, "cli-game-launcher")
	err = os.MkdirAll(programHome, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {
	err := handleCommandLine()
	if err != nil {
		log.Fatalln(err)
	}
}
