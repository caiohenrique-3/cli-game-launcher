package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	deleteConfigFileFromTempDir(t)

	game := []Game{{Name: "Test Game", PathToExecutable: "/test/path/here"}}
	goodData, err := json.Marshal(game)
	if err != nil {
		t.Fatalf("error during marshal: %v", err)
	}

	emptyConfigFileDir := setupLoadConfigFileTest(t, nil)
	badConfigFileDir := setupLoadConfigFileTest(t, []byte("test"))
	goodEndingDir := setupLoadConfigFileTest(t, goodData)

	tests := map[string]struct {
		input    string
		wantData []byte
		wantErr  error
	}{
		"path doesnt exist": {
			input:    "/some/404/path",
			wantData: nil,
			wantErr:  ErrDoesNotExistOrIsADirectory},
		"empty config file": {
			input:    emptyConfigFileDir,
			wantData: []byte("[]"),
			wantErr:  nil},
		"bad config file": {
			input:    badConfigFileDir,
			wantData: []byte("test"),
			wantErr:  nil},
		"path is a dir": {
			input:    os.TempDir(),
			wantData: nil,
			wantErr:  ErrDoesNotExistOrIsADirectory},
		"good ending": {
			input:    goodEndingDir,
			wantData: goodData,
			wantErr:  nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotData, gotErr := loadConfigFile(test.input)
			if !slices.Equal(gotData, test.wantData) {
				t.Errorf("got '%s'; want '%s'\n", gotData, test.wantData)
			}
			if !errors.Is(gotErr, test.wantErr) {
				t.Errorf("got '%v'; want '%v'\n", gotErr, test.wantErr)
			}
		})
	}
}

func TestAppendNewGameToConfigFileErrors(t *testing.T) {
	// test if it appended correctly
	deleteConfigFileFromTempDir(t)
	err := createEmptyConfigFileAt(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}

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
	err := createEmptyConfigFileAt(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}

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

	err = saveGameToConfig("Test Game", f.Name(), testDir)
	if err != nil {
		t.Error(err)
	}

	if !fileExists(configFile) {
		t.Error("config.json was not created!")
	}
}

func TestRemoveGameFromConfig(t *testing.T) {
	f, testDir := setupTest(t)

	deleteConfigFileFromTempDir(t) //testDir is tempdir

	for i := range 10 {
		err := saveGameToConfig(
			fmt.Sprintf("Test Game %v", i),
			f.Name(),
			testDir)
		if err != nil {
			t.Error(err)
		}
	}

	err := removeGameFromConfig(0, testDir)
	if err != nil {
		t.Error(err)
	}
	err = removeGameFromConfig(0, testDir)
	if err != nil {
		t.Error(err)
	}
	err = removeGameFromConfig(0, testDir)
	if err != nil {
		t.Error(err)
	}

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

func TestRemoveGameFromConfigNoGamesFound(t *testing.T) {
	deleteConfigFileFromTempDir(t)
	err := createEmptyConfigFileAt(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	err = removeGameFromConfig(0, os.TempDir())
	if !errors.Is(err, ErrNoGamesFound) {
		t.Errorf("want %v; got %v;\n", ErrNoGamesFound, err)
	}
}

func TestRemoveGameFromConfigInvalidOption(t *testing.T) {
	f, testDir := setupTest(t)
	deleteConfigFileFromTempDir(t)

	err := saveGameToConfig("Test Game", f.Name(), testDir)
	if err != nil {
		t.Errorf("save game to config failed: %v", err)
	}

	err = removeGameFromConfig(2, testDir)
	if !errors.Is(err, ErrInvalidOption) {
		t.Errorf("want %v; got %v;\n", ErrInvalidOption, err)
	}
}

// Creates a directory in os.TempDir with a config.json inside it.
// Optionally writes data to the config.json file.
func setupLoadConfigFileTest(t *testing.T, data []byte) string {
	dir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatalf("error creating dir: %v", err)
	}
	err = createEmptyConfigFileAt(dir)
	if err != nil {
		t.Fatalf("error creating config file: %v", err)
	}
	if data != nil {
		f := filepath.Join(dir, "config.json")
		err := os.WriteFile(f, data, os.ModePerm)
		if err != nil {
			t.Fatalf("error writing data to file: %v", err)
		}
	}

	return dir
}
