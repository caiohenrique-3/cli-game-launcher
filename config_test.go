package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveGameToConfigNotEmpty(t *testing.T) {
	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Error()
	}
	defer f.Close()

	testDir := os.TempDir()

	saveGameToConfig("Test Game", f.Name(), testDir)

	savedConfigData := loadConfigFile(testDir)
	dat := []Game{}
	err = json.Unmarshal(savedConfigData, &dat)
	if err != nil {
		t.Error(err)
	}
	if len(dat) == 0 {
		t.Error("unmarshaled []Game is empty!")
	}
}

func TestSaveGameToConfigAppendsCorrectly(t *testing.T) {
	testDir := os.TempDir()
	err := deleteFile(filepath.Join(testDir, "config.json"))
	if err != nil {
		t.Error(err)
	}

	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	saveGameToConfig("Test Game", f.Name(), testDir)
	saveGameToConfig("Test Game 2", f.Name(), testDir)

	foo := loadConfigFile(testDir)
	dat := []Game{}
	err = json.Unmarshal(foo, &dat)
	if err != nil {
		t.Error(err)
	}
	if len(dat) != 2 {
		t.Errorf("unmarshaled data has len %v, want %v", len(dat), 2)
	}
}

func TestSaveGameToConfigCreatesConfigFile(t *testing.T) {
	configFile := filepath.Join(os.TempDir(), "config.json")

	err := deleteFile(configFile)
	if err != nil {
		t.Error(err)
	}

	testDir := os.TempDir()

	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Error()
	}
	defer f.Close()

	saveGameToConfig("Test Game", f.Name(), testDir)

	if fileExists(configFile) != true {
		t.Error("config.json was not created!")
	}
}

func TestCreateEmptyConfigFile(t *testing.T) {
	configFile := filepath.Join(os.TempDir(), "config.json")
	err := deleteFile(configFile)
	if err != nil {
		t.Error(err)
	}

	createEmptyConfigFileAt(os.TempDir())

	f := loadConfigFile(os.TempDir())

	dat := []Game{}
	err = json.Unmarshal(f, &dat)
	if err != nil {
		t.Error(err)
	}

	if len(dat) != 0 {
		t.Error("config.json data not empty!")
	}
}

// If file exists, deletes it.
func deleteFile(path string) error {
	if fileExists(path) {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func createTempFileOnOsTempDir() (*os.File, error) {
	// empty string means use os.TempDir() return value
	f, err := os.CreateTemp("", "test_file")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f, nil // calling .name() on this is safe after close
}
