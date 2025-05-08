package main

import (
	"os"
	"path/filepath"
	"testing"
)

func createTempFileOnOsTempDir() (*os.File, error) {
	// empty string means use os.TempDir() return value
	f, err := os.CreateTemp("", "test_file")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f, nil // calling .name() on this is safe after close
}

func deleteConfigFileFromTempDir(t *testing.T) {
	f := filepath.Join(os.TempDir(), "config.json")
	if fileExists(f) {
		err := os.Remove(f)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// Returns a temporary file and its parent dir.
func setupTest(t *testing.T) (*os.File, string) {
	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Error("error creating temp file:", err)
	}
	defer f.Close()

	testDir := os.TempDir()

	return f, testDir
}
