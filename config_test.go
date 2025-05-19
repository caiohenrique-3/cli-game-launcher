package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAppendNewGameToConfigFileErrors(t *testing.T) {
	// test if it appended correctly
	deleteConfigFileFromTempDir(t)
	createEmptyConfigFileAt(os.TempDir())

	badFileData := []byte("ABC\tTEST\n99")

	testDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}

	testFile, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Fatal(err)
	}

	testGame := &Game{
		Name:             "Test Game",
		PathToExecutable: testFile.Name()}

	testConfigFilePath := filepath.Join(os.TempDir(),
		"config.json")

	testFileData, err := os.ReadFile(testConfigFilePath)
	if err != nil {
		t.Fatal(err)
	}

	readOnlyPath, err := os.CreateTemp(testDir, "read-only")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chmod(readOnlyPath.Name(), 0400)
	if err != nil {
		t.Fatal(err)
	}
	readOnlyPath.Close()

	tests := map[string]struct {
		fileData         []byte
		game             *Game
		pathToConfigFile string
		errMsg           string
	}{
		"fileData is bad": {
			fileData:         badFileData,
			game:             testGame,
			pathToConfigFile: testConfigFilePath,
			errMsg:           "invalid character"},
		"empty Game": {
			fileData:         testFileData,
			game:             &Game{},
			pathToConfigFile: testConfigFilePath,
			errMsg:           ""},
		"path not found": {
			fileData:         testFileData,
			game:             testGame,
			pathToConfigFile: "/test/404/path",
			errMsg:           "no such file or directory"},
		"path is read-only": {
			fileData:         testFileData,
			game:             testGame,
			pathToConfigFile: readOnlyPath.Name(),
			errMsg:           "permission denied"},
		"path is a dir": {
			fileData:         testFileData,
			game:             testGame,
			pathToConfigFile: testDir,
			errMsg:           "is a directory"},
		"good ending": {
			fileData:         testFileData,
			game:             testGame,
			pathToConfigFile: testConfigFilePath,
			errMsg:           ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := appendNewGameToConfigFile(
				test.fileData, test.game, test.pathToConfigFile)
			want := test.errMsg
			if !ErrorContains(got, want) {
				t.Errorf("got '%v'; want '%v';\n", got, want)
			}
		})
	}

}

func TestAppendNewGameToConfigFile(t *testing.T) {
	deleteConfigFileFromTempDir(t)
	createEmptyConfigFileAt(os.TempDir())
	testConfigFilePath := filepath.Join(os.TempDir(), "config.json")

	testTempFile, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Fatal(err)
	}

	want := 10
	for i := range want {
		testGame := &Game{
			Name:             fmt.Sprintf("Test Game %v", i),
			PathToExecutable: testTempFile.Name()}

		configFileData, err := os.ReadFile(testConfigFilePath)
		if err != nil {
			t.Fatalf("error reading file: %v\n", err)
		}

		appendNewGameToConfigFile(configFileData, testGame, testConfigFilePath)
	}

	configFileData, err := os.ReadFile(testConfigFilePath)
	if err != nil {
		t.Fatalf("error reading file: %v\n", err)
	}

	games := []Game{}
	err = json.Unmarshal(configFileData, &games)
	if err != nil {
		t.Fatalf("error on unmarshal: %v\n", err)

	}

	got := len(games)
	if got != want {
		t.Errorf("got length='%v'; want length='%v';\n", got, want)
	}
}

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
	err = os.Chmod(c, 0400)
	if err != nil {
		t.Fatal(err)
	}

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
				f, err := loadConfigFile(d)
				if err != nil {
					t.Error(err)
				}

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
	data, err := loadConfigFile(testDir)
	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(data, &games)
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
