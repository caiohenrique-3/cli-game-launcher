package main

import (
	"testing"
	"time"
)

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

func Test_GetTimeFromStrings_HappyPath(t *testing.T) {
	dateStr := "2025-09-12"
	hourStr := "16:49:30"
	got, err := getTimeFromStrings(dateStr, hourStr)
	gotDate := got.Format(time.DateOnly)
	gotTime := got.Format(time.TimeOnly)

	if err != nil {
		t.Errorf("got '%v'; want nil;\n", err)
	}
	if gotDate != dateStr {
		t.Errorf("got '%s'; want '%s';\n", gotDate, dateStr)
	}
	if gotTime != hourStr {
		t.Errorf("got '%s'; want '%s';\n", gotTime, hourStr)
	}
}
