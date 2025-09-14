package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"
)

func Test_CreateEmptyConfigFileAt_PathIsNotADirectory(t *testing.T) {
	setupTestsDir(t)
	tempFile, err := os.CreateTemp(testsDir, "test")
	if err != nil {
		t.Errorf("error creating temp file: %v\n", err)
	}
	defer tempFile.Close()

	var wantErr *fs.PathError
	err = createEmptyConfigFileAt(tempFile.Name())
	if !errors.As(err, &wantErr) {
		t.Errorf("got: '%v'; want: '*fs.PathError';\n", err)
	}
}

func Test_CreateEmptyConfigFileAt_HappyPath(t *testing.T) {
	setupTestsDir(t)
	pathToTempDir, err := os.MkdirTemp(testsDir, "tempDir")
	pathToConfigFile := filepath.Join(pathToTempDir, "config.json")
	if err != nil {
		t.Errorf("error creating temp dir: %v\n", err)
	}

	err = createEmptyConfigFileAt(pathToTempDir)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	gotConfigFileData, err := os.ReadFile(pathToConfigFile)
	if err != nil {
		t.Errorf("error reading file: %v\n", err)
	}

	wantConfigFileData := []byte("[]")
	if !bytes.Equal(gotConfigFileData, wantConfigFileData) {
		t.Errorf("got: '%s'; want '%s';\n", gotConfigFileData, wantConfigFileData)
	}
}

func Test_SaveGameToConfig_CreatesConfigFileIfMissing(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir, err := os.MkdirTemp(testsDir, "tempDir")
	if err != nil {
		t.Errorf("error creating temp dir: %v\n", err)
	}
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")
	pathToTempFile := createTempFileForTests(t)

	err = saveGameToConfig("Test Game", pathToTempFile, configFileParentDir)
	if err != nil {
		t.Error(err)
	}
	if !fileExists(pathToConfigFile) {
		t.Errorf("'%s' does not exist!\n", pathToConfigFile)
	}

	configFileData, err := loadConfigFile(configFileParentDir)
	if err != nil {
		t.Errorf("error loading config file: %v\n", err)
	}

	games := []Game{}
	err = json.Unmarshal(configFileData, &games)
	if err != nil {
		t.Errorf("error unmarshaling: %v\n", err)
	}

	if len(games) != 1 {
		t.Errorf("got length %v; want length %v;\n", len(games), 1)
	}
}

func Test_AppendNewGameToConfigFile_ConfigFileHasInvalidData(t *testing.T) {
	err := appendNewGameToConfigFile([]byte("invalid test"), nil, "")
	var wantErr *json.SyntaxError
	if !errors.As(err, &wantErr) {
		t.Errorf("got '%v'; want '*json.SyntaxError';\n", err)
	}
}

// TODO: I think passing an empty Game to it should throw some error.
func Test_AppendNewGameToConfigFile_GameIsEmpty(t *testing.T) {
	setupTestsDir(t)
	configFileData := []byte("[]")
	configFileParentDir := createTempDirWithConfigFile(t, configFileData)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")
	var game Game
	err := appendNewGameToConfigFile(configFileData, &game, pathToConfigFile)
	if err != nil {
		t.Errorf("got '%v'; want nil;\n", err)
	}
}

func Test_AppendNewGameToConfigFile_PathIsADirectory(t *testing.T) {
	setupTestsDir(t)
	configFileData := []byte("[]")
	configFileParentDir := createTempDirWithConfigFile(t, configFileData)

	var game Game
	err := appendNewGameToConfigFile(configFileData, &game, configFileParentDir)

	var wantErr *fs.PathError
	if !errors.As(err, &wantErr) {
		t.Errorf("got '%v'; want '*fs.PathError';\n", err)
	}
}

func Test_AppendNewGameToConfigFile_HappyPath(t *testing.T) {
	setupTestsDir(t)
	configFileData := []byte("[]")
	configFileParentDir := createTempDirWithConfigFile(t, configFileData)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	wantGamesLength := 10
	for i := range wantGamesLength {
		game := &Game{
			Name:             fmt.Sprintf("Test Game %v", i),
			PathToExecutable: ""}

		configFileData, err := os.ReadFile(pathToConfigFile)
		if err != nil {
			t.Errorf("error reading config file: %v\n", err)
		}

		err = appendNewGameToConfigFile(configFileData, game, pathToConfigFile)
		if err != nil {
			t.Errorf("error appending new game to config file: %v\n", err)
		}
	}

	configFileData, err := os.ReadFile(pathToConfigFile)
	if err != nil {
		t.Errorf("error reading config file: %v\n", err)
	}

	games := []Game{}
	err = json.Unmarshal(configFileData, &games)
	if err != nil {
		t.Errorf("error unmarshaling config file: %v\n", err)

	}

	gotGamesLength := len(games)
	if gotGamesLength != wantGamesLength {
		t.Errorf("got length='%v'; want length='%v';\n",
			gotGamesLength, wantGamesLength)
	}

	for i, game := range games {
		wantGameName := fmt.Sprintf("Test Game %v", i)
		if game.Name != wantGameName {
			t.Errorf("[%v]: got '%s'; want '%s';\n",
				i, game.Name, wantGameName)
		}
	}
}

func Test_RemoveGameFromConfig_HappyPath(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, []byte("[]"))
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	// pathToExecutable accepts any filepath as long as it exists.
	for i := range 10 {
		err := saveGameToConfig(
			fmt.Sprintf("Test Game %v", i),
			pathToConfigFile,
			configFileParentDir)
		if err != nil {
			t.Errorf("[%v] error saving game to config: %v\n", i, err)
		}
	}

	for i := range 3 {
		err := removeGameFromConfig(0, configFileParentDir)
		if err != nil {
			t.Errorf("[%v] error removing game from config: %v\n", i, err)
		}
	}

	games := []Game{}
	configFileData, err := loadConfigFile(configFileParentDir)
	if err != nil {
		t.Errorf("error loading config file: %v\n", err)
	}

	err = json.Unmarshal(configFileData, &games)
	if err != nil {
		t.Errorf("error unmarshaling: %v\n", err)
	}

	gotLength := len(games)
	wantLength := 7
	if gotLength != wantLength {
		t.Errorf("got length %v; want length %v\n", gotLength, wantLength)
	}

	for _, game := range games {
		if game.Name == "Test Game 0" ||
			game.Name == "Test Game 1" ||
			game.Name == "Test Game 2" {
			t.Errorf("'%s' was not deleted!\n", game.Name)
		}
	}
}

func Test_RemoveGameFromConfig_NoGamesFoundInConfigFile(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)

	err := removeGameFromConfig(0, configFileParentDir)
	if !errors.Is(err, ErrNoGamesFound) {
		t.Errorf("got '%v'; want '%v';\n", err, ErrNoGamesFound)
	}
}

func Test_RemoveGameFromConfig_InvalidOption(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)

	// pathToExecutable takes any path as long as it exists
	err := saveGameToConfig("Test Game", configFileParentDir, configFileParentDir)
	if err != nil {
		t.Errorf("save game to config failed: %v\n", err)
	}

	err = removeGameFromConfig(2, configFileParentDir)
	if !errors.Is(err, ErrInvalidOption) {
		t.Errorf("got '%v'; want '%v';\n", err, ErrInvalidOption)
	}
	err = removeGameFromConfig(-1, configFileParentDir)
	if !errors.Is(err, ErrInvalidOption) {
		t.Errorf("got '%v'; want '%v';\n", err, ErrInvalidOption)
	}
}

func Test_LoadConfigFile(t *testing.T) {
	setupTestsDir(t)
	games := []Game{{Name: "Test Game", PathToExecutable: ""}}
	dataForHappyPathTest, err := json.Marshal(games)
	if err != nil {
		t.Fatalf("error during marshal: %v\n", err)
	}

	emptyConfigFileParentDir := createTempDirWithConfigFile(t, nil)
	badConfigFileParentDir := createTempDirWithConfigFile(t, []byte("test"))
	happyPathParentDir := createTempDirWithConfigFile(t, dataForHappyPathTest)

	tests := map[string]struct {
		input    string
		wantData []byte
		wantErr  error
	}{
		"path does not exist": {
			input:    "a/b/c/404",
			wantData: nil,
			wantErr:  fs.ErrNotExist},
		"empty config file": {
			input:    emptyConfigFileParentDir,
			wantData: []byte("[]"),
			wantErr:  nil},
		"bad config file": {
			input:    badConfigFileParentDir,
			wantData: []byte("test"),
			wantErr:  nil},
		"happy path": {
			input:    happyPathParentDir,
			wantData: dataForHappyPathTest,
			wantErr:  nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotData, gotErr := loadConfigFile(test.input)
			if !slices.Equal(gotData, test.wantData) {
				t.Errorf("got '%s'; want '%s';\n", gotData, test.wantData)
			}
			if !errors.Is(gotErr, test.wantErr) {
				t.Errorf("got '%v'; want '%v';\n", gotErr, test.wantErr)
			}
		})
	}
}

func Test_GetGames_HappyPath(t *testing.T) {
	goodEndingGames := []Game{
		{Name: "Test Game 1", PathToExecutable: "/test/path/bar"},
		{Name: "Test Game 2", PathToExecutable: "/test/path/foo"}}

	goodEndingData, err := json.Marshal(goodEndingGames)
	if err != nil {
		t.Errorf("error during marshal: %v\n", err)
	}

	goodConfigFileDir := createTempDirWithConfigFile(t, goodEndingData)

	games, err := getGames(goodConfigFileDir, false)
	if !reflect.DeepEqual(games, goodEndingGames) {
		t.Errorf("got '%v'; want '%v';\n", games, goodEndingGames)
	}
}

func Test_GetGames_InvalidConfigFile(t *testing.T) {
	configFileParentDir := createTempDirWithConfigFile(t, []byte("some data"))
	games, err := getGames(configFileParentDir, false)

	var wantErr *json.SyntaxError
	if !errors.As(err, &wantErr) {
		t.Errorf("got '%v'; want '*json.SyntaxError';\n", err)
	}
	if games != nil {
		t.Errorf("got '%v'; want '%v'\n", games, nil)
	}
}

func Test_SaveTimeSpentPlayingToConfig_HappyPath(t *testing.T) {
	setupTestsDir(t)
	games := []Game{{
		Name:             "TestGame",
		PathToExecutable: "",
		TimeSpentPlaying: "0h0m0s"}}
	configFileData, err := json.Marshal(games)
	if err != nil {
		t.Errorf("error marshaling games: %v\n", err)
	}

	configFileParentDir := createTempDirWithConfigFile(t, configFileData)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")
	timeSpentPlaying := time.Duration(2) * time.Hour
	userInput := 0

	err = saveTimeSpentPlayingToConfig(games, userInput, timeSpentPlaying, pathToConfigFile)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	newConfigFileData, err := loadConfigFile(configFileParentDir)
	var newGames []Game
	err = json.Unmarshal(newConfigFileData, &newGames)
	if err != nil {
		t.Errorf("error unmarshaling new config file: %v\n", err)
	}

	gotTimeSpentPlaying := newGames[0].TimeSpentPlaying
	if gotTimeSpentPlaying != timeSpentPlaying.String() {
		t.Errorf("got: '%s'; want '%s';\n", gotTimeSpentPlaying, timeSpentPlaying)
	}
}

func Test_GetGames_ExcludesHiddenGames(t *testing.T) {
	setupTestsDir(t)

	games := []Game{
		{Name: "Test Game 1"},
		{Name: "Test Game 2", IsHidden: true}}

	configFileData, err := json.Marshal(games)
	if err != nil {
		t.Fatalf("marshal failed: '%v'\n", err)
	}

	configFileParentDir := createTempDirWithConfigFile(t, configFileData)

	excludeHiddenGames := true
	gotGames, err := getGames(configFileParentDir, excludeHiddenGames)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	for _, g := range gotGames {
		if g.IsHidden {
			t.Fatalf("got: '%v'; must NOT contain hidden games.\n",
				g)
		}

		if g.Name == "Test Game 2" {
			t.Fatalf("must NOT contain 'Test Game 2'.\n")
		}
	}
}

func Test_ToggleHiddenGame_HappyPath(t *testing.T) {
	setupTestsDir(t)

	games := []Game{{Name: "Test Game 1", IsHidden: false}}
	configFileData, err := json.Marshal(games)
	if err != nil {
		t.Fatalf("marshal failed: '%v'\n", err)
	}

	configFileParentDir := createTempDirWithConfigFile(t, configFileData)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")
	userInput := 0

	err = toggleHiddenGame(userInput, pathToConfigFile)
	if err != nil {
		t.Fatalf("got: '%v'; want nil\n", err)
	}

	gotGames, err := getGames(configFileParentDir, false)
	if err != nil {
		t.Fatalf("get games failed: '%v'\n", err)
	}

	for _, game := range gotGames {
		if !game.IsHidden {
			t.Errorf("got: '%t'; want: '%t';\n",
				game.IsHidden, true)
		}
	}

	err = toggleHiddenGame(userInput, pathToConfigFile)
	if err != nil {
		t.Fatalf("got: '%v'; want nil\n", err)
	}

	// Config file changed with toggleHiddenGame call, getting games again
	gotGames, err = getGames(configFileParentDir, false)
	if err != nil {
		t.Fatalf("get games failed: '%v'\n", err)
	}

	for _, game := range gotGames {
		if game.IsHidden {
			t.Errorf("got: '%t'; want: '%t';\n",
				game.IsHidden, false)
		}
	}
}
