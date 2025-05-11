package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCreateEmptyConfigFileAt(t *testing.T) {
	deleteConfigFileFromTempDir(t)

	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Fatal(err)
	}

	d, err := os.MkdirTemp(os.TempDir(), "good-ending")
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(d, "config.json"), nil, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	c := filepath.Join(os.TempDir(), "config.json")
	err = os.WriteFile(c, nil, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	os.Chmod(c, 0400)

	tests := map[string]struct {
		input  string
		errMsg string
	}{
		"path is not a dir": {
			input:  f.Name(),
			errMsg: "not a directory"},
		"path not found": {
			input:  "/test/path/404",
			errMsg: "no such file or directory"},
		"config.json is read-only": {
			input:  os.TempDir(),
			errMsg: "permission denied"},
		"good ending": {
			input:  d,
			errMsg: ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := createEmptyConfigFileAt(test.input)
			want := test.errMsg

			if !ErrorContains(got, want) {
				t.Errorf("got '%v'; want '%s';\n", got, want)
			}

			if name == "good ending" {
				f := loadConfigFile(d)

				dat := []Game{}
				err = json.Unmarshal(f, &dat)
				if err != nil {
					t.Error(err)
				}

				gotLen := len(dat)
				wantLen := 0
				if len(dat) != 0 {
					t.Errorf("got '%v'; want '%v';\n", gotLen, wantLen)
				}
			}
		})
	}
}

func TestAppendNewGameToConfigFileAppendsCorrectly(t *testing.T) {
	testTempFile, testDir := setupTest(t)
	configFile := filepath.Join(testDir, "config.json")

	err := deleteFile(configFile)
	if err != nil {
		t.Error(err)
	}

	err = createEmptyConfigFileAt(testDir)
	if err != nil {
		t.Fatal(err)
	}

	gameName := "Test Game"
	gamePath := testTempFile.Name()
	game := &Game{
		Name:             gameName,
		PathToExecutable: gamePath,
	}

	var f []byte
	for range 3 {
		f = loadConfigFile(testDir)
		appendNewGameToConfigFile(f, game, configFile)
	}

	f = loadConfigFile(testDir)

	data := []Game{}
	err = json.Unmarshal(f, &data)
	if err != nil {
		t.Error(err)
	}
	if len(data) != 3 {
		t.Errorf("unmarshaled data has len %v, want %v",
			len(data), 2)
	}

	for i, v := range data {
		if v.Name != gameName {
			t.Errorf("key %v: want %v; got %v",
				i, gameName, v.Name)
		}

		if v.PathToExecutable != gamePath {
			t.Errorf("key %v: want %v; got %v",
				i, gamePath, v.PathToExecutable)
		}
	}

}

func TestAppendNewGameToConfigFileBadJson(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		testTempFile, testDir := setupTest(t)
		configFile := filepath.
			Join(testDir, "config.json")

		err := deleteFile(configFile)
		if err != nil {
			t.Error(err)
		}

		err = createEmptyConfigFileAt(testDir)
		if err != nil {
			t.Fatal(err)
		}

		testInput := "TESTTESTTEST1234567890"

		err = os.WriteFile(configFile,
			[]byte(testInput), os.ModePerm)
		if err != nil {
			t.Error("writing bad input on file:", err)
			return
		}

		gameName := "Test Game"
		gamePath := testTempFile.Name()
		game := &Game{
			Name:             gameName,
			PathToExecutable: gamePath,
		}

		f := loadConfigFile(testDir)
		appendNewGameToConfigFile(f, game, configFile)
	}

	cmd := exec.Command(os.Args[0],
		"-test.run=TestAppendNewGameToConfigFileBadJson")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v; want exit status 1", err)
}

func TestSaveGameToConfigCreatesConfigFileIfMissing(t *testing.T) {
	f, testDir := setupTest(t)
	configFile := filepath.Join(testDir, "config.json")

	err := deleteFile(configFile)
	if err != nil {
		t.Error(err)
	}

	saveGameToConfig("Test Game", f.Name(), testDir)

	if !fileExists(configFile) {
		t.Error("config.json was not created!")
	}
}

func TestRemoveGameFromConfig(t *testing.T) {
	f, testDir := setupTest(t)

	err := deleteFile(filepath.Join(testDir, "config.json"))
	if err != nil {
		t.Error(err)
	}

	for i := range 10 {
		saveGameToConfig(
			fmt.Sprintf("Test Game %v", i),
			f.Name(),
			testDir)
	}

	removeGameFromConfig(0, testDir)
	removeGameFromConfig(0, testDir)
	removeGameFromConfig(0, testDir)

	games := []Game{}
	err = json.Unmarshal(loadConfigFile(testDir), &games)
	if err != nil {
		t.Error(err)
	}

	if len(games) != 7 {
		t.Errorf("want length=%v; got length=%v", 7, len(games))
	}

	for _, k := range games {
		if k.Name == "Test Game 0" ||
			k.Name == "Test Game 1" ||
			k.Name == "Test Game 2" {
			t.Errorf("%v: was not deleted!", k.Name)
		}
	}
}

func TestRemoveGameFromConfigEmptyGames(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		err := deleteFile(filepath.
			Join(os.TempDir(), "config.json"))
		if err != nil {
			t.Error(err)
		}

		err = createEmptyConfigFileAt(os.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		removeGameFromConfig(0, os.TempDir())
	}

	cmd := exec.Command(os.Args[0],
		"-test.run=TestRemoveGameFromConfigEmptyGames")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v; want exit status 1", err)
}

func TestRemoveGameFromConfigInvalidOption(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		f, testDir := setupTest(t)

		err := deleteFile(filepath.
			Join(os.TempDir(), "config.json"))
		if err != nil {
			t.Error(err)
		}

		saveGameToConfig("Test Game", f.Name(), testDir)
		removeGameFromConfig(2, testDir)
	}

	cmd := exec.Command(os.Args[0],
		"-test.run=TestRemoveGameFromConfigInvalidOption")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with err %v; want exit status 1", err)
}
