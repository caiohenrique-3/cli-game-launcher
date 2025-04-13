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
	// Get executable path from somewhere
	// Then pass it to runNative()
	err := handleCommandLine()
	if err != nil {
		os.Exit(1)
	}

	// Pass a full comand to RunNative here,
	// eg. "sbx run instance"
	// then split it and pass
	/* 	err := run.RunNative()
	   	if err != nil {
	   		fmt.Println(err)
	   	} */
}
