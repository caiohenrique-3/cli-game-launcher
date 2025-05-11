package main

import (
	"os"
	"path/filepath"
	"strings"
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

// ErrorContains checks if the error message in out contains the text in
// want.
//
// This is safe when out is nil. Use an empty string for want if you want to
// test that err is nil.
func ErrorContains(out error, want string) bool {
	if out == nil {
		return want == ""
	}
	if want == "" {
		return false
	}
	return strings.Contains(out.Error(), want)
}
