package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

// "cgl-tests" inside OS temp directory.
var testsDir string = filepath.Join(os.TempDir(), "cgl-tests")

// Creates "cgl-tests" dir inside OS temp directory if it doesn't exist.
func setupTestsDir(t *testing.T) {
	_, err := os.Stat(testsDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(testsDir, os.ModePerm)
		if err != nil {
			t.Fatalf("error creating tests dir: '%v'\n", err)
		}
	} else {
		if err != nil {
			t.Fatalf("error setting up tests dir: '%v'\n", err)
		}
	}
}

// Creates an empty file in directory "cgl-tests" inside OS temp directory and
// returns the path to it.
func createTempFileForTests(t *testing.T) string {
	testTempFile, err := os.CreateTemp(testsDir, "testTempFile")
	if err != nil {
		t.Errorf("error creating temp file: '%v'\n", err)
	}
	defer testTempFile.Close()

	return testTempFile.Name()
}

// Reads testsDir global variable and creates a directory on that path
// and a 'config.json' file inside the new directory.
// Returns path to the new created directory.
func createTempDirWithConfigFile(t *testing.T, configFileData []byte) string {
	tempDir, err := os.MkdirTemp(testsDir, "testdir")
	if err != nil {
		t.Errorf("error creating temp dir: %v\n", err)
	}
	err = createEmptyConfigFileAt(tempDir)
	if err != nil {
		t.Errorf("error creating config file at temp dir: %v\n", err)
	}

	if configFileData != nil {
		f := filepath.Join(tempDir, "config.json")
		err := os.WriteFile(f, configFileData, os.ModePerm)
		if err != nil {
			t.Errorf("error writing data to file: %v\n", err)
		}
	}

	return tempDir
}

// Supresses output in the terminal.
func quietOutput() func() {
	null, _ := os.Open(os.DevNull)
	sout := os.Stdout
	serr := os.Stderr
	os.Stdout = null
	os.Stderr = null
	log.SetOutput(null)
	return func() {
		defer null.Close()
		os.Stdout = sout
		os.Stderr = serr
		log.SetOutput(os.Stderr)
	}
}
