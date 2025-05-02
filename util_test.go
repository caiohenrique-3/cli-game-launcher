package main

import (
	"os"
	"runtime"
	"testing"
)

func TestGetCleanPathExpandsTilde(t *testing.T) {
	if runtime.GOOS != "linux" {
		return
	}

	pathInput := "~/"
	s := getCleanPath(pathInput)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
	}

	if s != home {
		t.Error("tilde did not expand user home dir!")
	}
}
