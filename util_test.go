package main

import "testing"

func Test_FileExists_ReturnsTrue(t *testing.T) {
	setupTestsDir(t)
	pathToTempFile := createTempFileForTests(t)
	bFileExists := fileExists(pathToTempFile)
	if !bFileExists {
		t.Errorf("got '%t'; want '%t';\n", bFileExists, true)
	}
}

func Test_FileExists_ReturnsFalse(t *testing.T) {
	bFileExists := fileExists("")
	if bFileExists {
		t.Errorf("got: '%t'; want '%t';\n", bFileExists, false)
	}
}

func Test_FileExists_ReturnsFalseIfPathIsADirectory(t *testing.T) {
	setupTestsDir(t)
	bFileExists := fileExists(testsDir)
	if bFileExists {
		t.Errorf("got '%t'; want '%t';\n", bFileExists, false)
	}
}
