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
	"time"
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

func Test_ListGamesWithPlaytime_HappyPath(t *testing.T) {
	setupTestsDir(t)
	games := []Game{{
		Name:             "Super Race 2",
		PathToExecutable: "",
		TimeSpentPlaying: "2h0m0s"}}

	configFileData, err := json.Marshal(games)
	if err != nil {
		t.Errorf("error marshaling config file data: %v\n", err)
	}
	configFileParentDir := createTempDirWithConfigFile(t, configFileData)

	var b bytes.Buffer
	err = listGamesWithPlaytime(configFileParentDir, &b)
	if err != nil {
		t.Errorf("got '%v'; want nil;\n", err)
	}

	gotOutput := b.String()
	if !strings.Contains(gotOutput, "Super Race 2") {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, "Super Race 2")
	}
	if !strings.Contains(gotOutput, "2h0m0s") {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, "2h0m0s")
	}
}

func Test_ListGamesWithPlaytime_ChecksForTimeSpentEmptyString(t *testing.T) {
	setupTestsDir(t)
	games := []Game{{
		Name:             "Super Race 2",
		PathToExecutable: "",
		TimeSpentPlaying: ""}}

	configFileData, err := json.Marshal(games)
	if err != nil {
		t.Errorf("error marshaling config file data: %v\n", err)
	}
	configFileParentDir := createTempDirWithConfigFile(t, configFileData)

	var b bytes.Buffer
	err = listGamesWithPlaytime(configFileParentDir, &b)
	if err != nil {
		t.Errorf("got '%v'; want nil;\n", err)
	}

	gotOutput := b.String()
	if !strings.Contains(gotOutput, "0h0m0s") {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, "0h0m0s")
	}
}

func Test_ListGamesWithLastPlayed_HappyPath(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToTempFile := createTempFileForTests(t)
	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")

	for i := range 3 {
		err := saveGameToConfig("Test Game"+strconv.Itoa(i), pathToTempFile, configFileParentDir)
		if err != nil {
			t.Errorf("[%d] error saving game to config: %v\n", i, err)
		}
	}

	// Creating logs.json file with necessary data
	time6DaysAgo := time.Now().AddDate(0, 0, -6)
	timeToday := time.Now()

	time6DaysAgoString := fmt.
		Sprintf("[%s] Played '%s' from 12:24:30 to 12:30:00.\n",
			time6DaysAgo.Format(time.DateOnly),
			"Test Game1")
	timeTodayString := fmt.
		Sprintf("[%s] Played '%s' from 12:00:09 to 12:30:59.\n",
			timeToday.Format(time.DateOnly),
			"Test Game2")
	// Removed/hidden because the game name is not in the config file.
	// For testing to see if the function will include this in the output
	// (It should not).
	hiddenGameString := fmt.
		Sprintf("[%s] Played '%s' from 12:00:00 to 12:30:00.\n",
			timeToday.Format(time.DateOnly),
			"Test Game5")

	var sb strings.Builder
	// Not checking for sb errors here because the test will fail
	// if these are not right anyway.
	sb.WriteString(time6DaysAgoString)
	sb.WriteString(timeTodayString)
	sb.WriteString(hiddenGameString)

	logFileDataString := sb.String()

	err := os.WriteFile(pathToLogFile, []byte(logFileDataString), os.ModePerm)
	if err != nil {
		t.Errorf("error writing to log file: %v\n", err)
	}

	var b bytes.Buffer
	err = listGamesWithLastPlayed(configFileParentDir, &b)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	gotOutput := b.String()
	if !strings.Contains(gotOutput, "Today") {
		t.Errorf("got: \n'%s'; must contain: '%s';\n", gotOutput, "Today")
	}
	if !strings.Contains(gotOutput, "6 days ago") {
		t.Errorf("got: \n'%s'; must contain: '%s';\n", gotOutput, "6 days ago")
	}
	if strings.Contains(gotOutput, "Test Game5") {
		t.Errorf("got: \n'%s'; must NOT contain: '%s';\n", gotOutput, "Test Game5")
	}
}

func Test_ListGamesWithLastPlayed_NoGamesFound(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")

	err := os.WriteFile(pathToLogFile, nil, os.ModePerm)
	if err != nil {
		t.Errorf("error writing to log file: %v\n", err)
	}

	var b bytes.Buffer
	err = listGamesWithLastPlayed(configFileParentDir, &b)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	gotOutput := b.String()
	wantOutput := "No games found in log file.\n"
	if gotOutput != wantOutput {
		t.Errorf("got: '%s'; want: '%s';\n", gotOutput, wantOutput)
	}
}

func Test_ListGamesWithLastPlayed_Prints1DayAgo(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToTempFile := createTempFileForTests(t)
	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")

	err := saveGameToConfig("Test Game", pathToTempFile, configFileParentDir)
	if err != nil {
		t.Errorf("error saving game to config: %v\n", err)
	}

	// Creating logs.json file with necessary data
	time1DayAgo := time.Now().AddDate(0, 0, -1)

	time1DayAgoString := fmt.
		Sprintf("[%s] Played '%s' from 12:00:00 to 12:30:20.\n",
			time1DayAgo.Format(time.DateOnly),
			"Test Game")

	err = os.WriteFile(pathToLogFile, []byte(time1DayAgoString), os.ModePerm)
	if err != nil {
		t.Errorf("error writing to log file: %v\n", err)
	}

	var b bytes.Buffer
	err = listGamesWithLastPlayed(configFileParentDir, &b)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	gotOutput := b.String()
	if !strings.Contains(gotOutput, "1 day ago") {
		t.Errorf("got: \n'%s'; must contain: '%s';\n",
			gotOutput, "1 day ago")
	}
}

func Test_ListGamesWithPlaytimeLastTwoWeeks_HappyPath(t *testing.T) {
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToLogFile := filepath.Join(configFileParentDir, "logs.json")
	pathToTempFile := createTempFileForTests(t)

	for i := 1; i < 4; i++ {
		err := saveGameToConfig(
			"Test Game "+strconv.Itoa(i),
			pathToTempFile,
			configFileParentDir)
		if err != nil {
			t.Errorf("[%d] error saving game to config: %v\n", i, err)
		}
	}

	// These dates have to be in the last two weeks.
	timeNow := time.Now().UTC()
	time1DayLater := timeNow.AddDate(0, 0, 1)

	logEntry4HoursPlayed := fmt.
		Sprintf("[%s] Played 'Test Game 1' from 08:00:30 to 12:00:59.\n",
			timeNow.Format(time.DateOnly))
	logEntry8HoursPlayed := fmt.
		Sprintf("[%s] Played 'Test Game 2' from 08:00:09 to 16:00:20.\n",
			timeNow.Format(time.DateOnly))
	logEntry24HoursPlayed := fmt.
		Sprintf("[%s] Played 'Test Game 3' from 08:00:21 to %s 08:00:34.\n",
			timeNow.Format(time.DateOnly),
			time1DayLater.Format(time.DateOnly))

	// Not checking for write errors, if this fails the test will fail too.
	var sb strings.Builder
	sb.WriteString(logEntry4HoursPlayed)
	sb.WriteString(logEntry8HoursPlayed)
	sb.WriteString(logEntry24HoursPlayed)

	err := os.WriteFile(pathToLogFile, []byte(sb.String()), os.ModePerm)
	if err != nil {
		t.Errorf("error writing to log file: %v\n", err)
	}

	var b bytes.Buffer
	err = listGamesWithPlaytimeLastTwoWeeks(configFileParentDir, &b)
	if err != nil {
		t.Errorf("got '%v'; want nil;\n", err)
	}

	gotOutput := b.String()
	totalHours := "36h0m53s spent playing last two weeks."
	if !strings.Contains(gotOutput, totalHours) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, totalHours)
	}

	game1 := "Test Game 1"
	playtime1 := "4h0m29s (11.1%)"
	if !strings.Contains(gotOutput, game1) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, game1)
	}
	if !strings.Contains(gotOutput, playtime1) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, playtime1)
	}

	game2 := "Test Game 2"
	playtime2 := "8h0m11s (22.2%)"
	if !strings.Contains(gotOutput, game2) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, game2)
	}
	if !strings.Contains(gotOutput, playtime2) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, playtime2)
	}

	game3 := "Test Game 3"
	playtime3 := "24h0m13s (66.6%)"
	if !strings.Contains(gotOutput, game3) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, game3)
	}
	if !strings.Contains(gotOutput, playtime3) {
		t.Errorf("got: '%s'; must contain: '%s';\n", gotOutput, playtime3)
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
