package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func Test_AddGamePrompt(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	// Using this for happy path, any path will do as long as it exists.
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	fileWithSpaces := setupAddGamePromptTest(t, "With Some Spaces")
	fileWithEmojis := setupAddGamePromptTest(t, "😀😃😄😁🤣🥲🥹☺️")
	fileWithCJK := setupAddGamePromptTest(t, "史诗 テスト 파일")

	// From OS user config directory to OS temp directory
	origProgramHome := programHome
	programHome = configFileParentDir
	defer func() { programHome = origProgramHome }()
	defer quietOutput()()

	// input is: "<game name>\n<path to executable>\n"
	tests := map[string]struct {
		input  string
		result error
	}{
		"empty string": {
			input:  "\n\n",
			result: ErrDoesNotExistOrIsADirectory},
		"file not found": {
			input:  "test\nSomeNonExistentFileHere\n",
			result: ErrDoesNotExistOrIsADirectory},
		"happy path": {
			input:  fmt.Sprintf("test\n%s\n", pathToConfigFile),
			result: nil},
		"path with spaces": {
			input:  fmt.Sprintf("test\n%s\n", fileWithSpaces.Name()),
			result: nil},
		"path is a directory": {
			input:  fmt.Sprintf("test\n%s\n", testsDir),
			result: ErrDoesNotExistOrIsADirectory},
		"path has emojis": {
			input:  fmt.Sprintf("test\n%s\n", fileWithEmojis.Name()),
			result: nil},
		"cjk path": {
			input:  fmt.Sprintf("test\n%s\n", fileWithCJK.Name()),
			result: nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(test.input))

			got := addGamePrompt(r)
			want := test.result

			if !errors.Is(got, want) {
				t.Fatalf("got '%v'; want '%v';\n", got, want)
			}
		})
	}
}

func Test_RemoveGamePrompt(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	for i := range 10 {
		// pathToExecutable takes any path as long as it exists.
		err := saveGameToConfig("Test Game", pathToConfigFile, configFileParentDir)
		if err != nil {
			t.Errorf("[%v] error saving game to config: %v\n", i, err)
		}
	}

	origProgramHome := programHome
	programHome = configFileParentDir
	defer func() {
		programHome = origProgramHome
	}()
	defer quietOutput()()

	tests := map[string]struct {
		input  string
		result error
	}{
		"empty string": {
			input:  "\n",
			result: strconv.ErrSyntax},
		"happy path": {
			input:  "0\n",
			result: nil},
		"input with spaces": {
			input:  "1 2 3 4 5 6 7\n",
			result: strconv.ErrSyntax},
		"input has no numbers": {
			input:  "no numbers here\n",
			result: strconv.ErrSyntax},
		"emoji input": {
			input:  "☺️\n",
			result: strconv.ErrSyntax},
		"cjk input": {
			input:  "史诗 テスト 파일\n",
			result: strconv.ErrSyntax},
		"less than zero": {
			input:  "-1\n",
			result: ErrInvalidOption},
		"notation": {
			input:  "1e9\n",
			result: strconv.ErrSyntax},
		"negative notation": {
			input:  "-1e9\n",
			result: strconv.ErrSyntax},
		"tab character": {
			input:  "2\t\n",
			result: nil},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(test.input))

			got := removeGamePrompt(r)
			want := test.result

			if !errors.Is(got, want) {
				t.Errorf("got '%v'; want '%v';\n", got, want)
			}
		})
	}
}

func Test_ListGamesNumbered_HappyPath(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")
	// pathToExecutable can be any file as long as it exists.
	err := saveGameToConfig("Test Game", pathToConfigFile, configFileParentDir)
	if err != nil {
		t.Errorf("error saving game to config: %v\n", err)
	}

	var b bytes.Buffer
	err = listGamesNumbered(configFileParentDir, &b, ListCmdOptions{})
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	capturedOutput := b.Bytes()
	wantOutput := []byte("[0] Test Game\n")
	if !bytes.Equal(capturedOutput, wantOutput) {
		t.Errorf("got: '%s'; want '%s';\n", capturedOutput, wantOutput)
	}
}

func Test_ListGamesNumbered_ConfigFileNotFound(t *testing.T) {
	setupTestsDir(t)
	pathToEmptyDir, err := os.MkdirTemp(testsDir, "emptyDir")
	if err != nil {
		t.Errorf("error creating temp dir: %v\n", err)
	}

	err = listGamesNumbered(pathToEmptyDir, nil, ListCmdOptions{})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("got: '%v'; want '%v';\n", err, fs.ErrNotExist)
	}
}

func Test_ListGamesNumbered_NoGamesFound(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)

	err := listGamesNumbered(configFileParentDir, nil, ListCmdOptions{})
	if !errors.Is(err, ErrNoGamesFound) {
		t.Errorf("got: '%v'; want '%v';\n", err, ErrNoGamesFound)
	}
}

func Test_ListGamesNumbered_DoesNotShowAHiddenGame(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	games := []Game{
		{IsHidden: true, Name: "Test Game"},
	}

	configFileData, err := json.Marshal(games)
	if err != nil {
		t.Fatalf("marshal failed: '%v'\n", err)
	}

	err = os.WriteFile(pathToConfigFile, configFileData, os.ModePerm)
	if err != nil {
		t.Fatalf("write file failed: '%v'\n", err)
	}

	var b bytes.Buffer
	cmdOptions := &ListCmdOptions{showHidden: false}
	err = listGamesNumbered(configFileParentDir, &b, *cmdOptions)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	capturedOutput := b.Bytes()
	wantOutput := []byte("")
	if !bytes.Equal(capturedOutput, wantOutput) {
		t.Fatalf("got: '%s'; want '%s';\n", capturedOutput, wantOutput)
	}
}

func Test_ListGamesNumbered_OnlyShowsHiddenGames(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	games := []Game{
		{IsHidden: true, Name: "Test Game 1"},
		{IsHidden: true, Name: "Test Game 2"},
		{IsHidden: false, Name: "Test Game 3"},
		{IsHidden: false, Name: "Test Game 4"},
	}

	configFileData, err := json.Marshal(games)
	if err != nil {
		t.Fatalf("marshal failed: '%v'\n", err)
	}

	err = os.WriteFile(pathToConfigFile, configFileData, os.ModePerm)
	if err != nil {
		t.Fatalf("write file failed: '%v'\n", err)
	}

	var b bytes.Buffer
	cmdOptions := &ListCmdOptions{onlyShowHidden: true}
	err = listGamesNumbered(configFileParentDir, &b, *cmdOptions)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	capturedOutput := b.Bytes()
	wantOutput := []byte("[0] Test Game 1\n[1] Test Game 2\n")
	if !bytes.Equal(capturedOutput, wantOutput) {
		t.Fatalf("got: '%s'; want '%s';\n", capturedOutput, wantOutput)
	}
}

func Test_ListGamesNumbered_ShowsHiddenGameIndicators(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	games := []Game{
		{IsHidden: true, Name: "Test Game 1"},
		{IsHidden: true, Name: "Test Game 2"},
		{IsHidden: false, Name: "Test Game 3"},
		{IsHidden: false, Name: "Test Game 4"},
	}

	configFileData, err := json.Marshal(games)
	if err != nil {
		t.Fatalf("marshal failed: '%v'\n", err)
	}

	err = os.WriteFile(pathToConfigFile, configFileData, os.ModePerm)
	if err != nil {
		t.Fatalf("write file failed: '%v'\n", err)
	}

	var b bytes.Buffer
	cmdOptions := &ListCmdOptions{showHidden: true, hiddenGameIndicator: true}
	err = listGamesNumbered(configFileParentDir, &b, *cmdOptions)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	capturedOutput := b.Bytes()
	wantOutput := []byte("[0] Test Game 1 [HIDDEN]\n[1] Test Game 2 [HIDDEN]\n[2] Test Game 3\n[3] Test Game 4\n")
	if !bytes.Equal(capturedOutput, wantOutput) {
		t.Fatalf("got: '%s'; want '%s';\n", capturedOutput, wantOutput)
	}
}

func Test_GetIntFromUser(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToConfigFile := filepath.Join(configFileParentDir, "config.json")

	for range 6 {
		// pathToExecutable accepts any path as long as it exists.
		err := saveGameToConfig("Test Game", pathToConfigFile, configFileParentDir)
		if err != nil {
			t.Errorf("error saving game to config: %v\n", err)
		}
	}

	defer quietOutput()()

	tests := map[string]struct {
		input         string
		wantReturnVal int
		wantErr       error
	}{
		"empty string": {
			input:   "\n",
			wantErr: strconv.ErrSyntax},
		"empty string 2": {
			input:   "\n\n",
			wantErr: strconv.ErrSyntax},
		"happy path": {
			input:         "4\n",
			wantReturnVal: 4,
			wantErr:       nil},
		"input has spaces": {
			input:         "2 \n",
			wantReturnVal: 2,
			wantErr:       nil},
		"input has spaces 2": {
			input:         " 2\n",
			wantReturnVal: 2,
			wantErr:       nil},
		"input is negative": {
			input:         "-2\n",
			wantReturnVal: -2,
			wantErr:       nil},
		"input has emojis": {
			input:   "😀\n",
			wantErr: strconv.ErrSyntax},
		"input has cjk": {
			input:   "접는 사람 テスト 試験 史诗般的\n",
			wantErr: strconv.ErrSyntax},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(test.input))
			userInput, err := getIntFromUser(r)

			if !errors.Is(err, test.wantErr) {
				t.Errorf("got '%v'; want '%v';\n", err, test.wantErr)
			}
			if userInput != test.wantReturnVal {
				t.Errorf("got '%v'; want '%v';\n", userInput, test.wantReturnVal)
			}
		})
	}
}

// Creates a directory and a file inside it, both are named with name + random numbers.
func setupAddGamePromptTest(t *testing.T, name string) *os.File {
	tempDir, err := os.MkdirTemp(testsDir, name)
	if err != nil {
		t.Errorf("error creating temp dir: %v\n", err)
	}

	tempFile, err := os.CreateTemp(tempDir, name)
	if err != nil {
		t.Errorf("error creating temp file: %v\n", err)
	}
	tempFile.Close()

	return tempFile
}
