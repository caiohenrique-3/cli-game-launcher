package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddGamePromptFileNotFound(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		input := "Test1234567890\nTest Game\n"
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
		return
	}

	cmd := exec.Command(os.Args[0],
		"-test.run=TestAddGamePromptFileNotFound")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v; want exit status 1", err)
}

func TestAddGamePromptGoodEnding(t *testing.T) {
	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Error(err)
	}

	input := "Test Game Name\n" + f.Name() + "\n"
	r := bufio.NewReader(strings.NewReader(input))

	orig := programHome
	programHome = os.TempDir()
	addGamePrompt(r)
	programHome = orig
}

func TestAddGamePromptPathWithSpaces(t *testing.T) {
	testDir, err := os.MkdirTemp("", "Dir With Spaces")
	if err != nil {
		t.Error(err)
	}

	f, err := os.CreateTemp(testDir, "TESTFILE")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	input := "Test Game Name\n" + f.Name() + "\n"
	r := bufio.NewReader(strings.NewReader(input))
	orig := programHome
	programHome = os.TempDir()
	addGamePrompt(r)
	programHome = orig
}

func TestAddGamePromptPathIsADirectory(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		testDir, err := os.MkdirTemp("", "TESTDIR")
		if err != nil {
			t.Error(err)
		}

		input := "Test Game Name\n" + testDir + "\n"
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
		return
	}
	cmd := exec.Command(os.Args[0],
		"-test.run=TestAddGamePromptPathIsADirectory")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v; want exit status 1", err)
}

func TestAddGamePromptExpandsEnvVariables(t *testing.T) {
	if os.Getenv("TEST_TMPDIR") != "" {
		f, err := createTempFileOnOsTempDir()
		if err != nil {
			t.Error("error creating temp file:", err)
		}

		fileName := filepath.Base(f.Name())

		input := "Test Game Name\n" +
			"$TEST_TMPDIR/" + fileName + "\n"
		r := bufio.NewReader(strings.NewReader(input))
		orig := programHome
		programHome = os.TempDir()
		addGamePrompt(r)
		programHome = orig
		return
	}

	cmd := exec.Command(os.Args[0],
		"-test.run=TestAddGamePromptExpandsEnvVariables")
	cmd.Env = append(os.Environ(), "TEST_TMPDIR="+os.TempDir())
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v; want exit status 0", err)
	}
}

func TestFileExistsTrue(t *testing.T) {
	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Error(err)
	}

	b := fileExists(f.Name())
	if !b {
		t.Errorf("error: %v; want true", b)
	}
}

func TestExistsFalse(t *testing.T) {
	b := fileExists("/mint/dragon/path/three")
	if b {
		t.Errorf("error: %v; want false", b)
	}
}

func TestListGamesNumbered(t *testing.T) {
	f, testDir := setupTest(t)

	err := deleteFile(filepath.Join(testDir, "config.json"))
	if err != nil {
		t.Error(err)
	}

	createEmptyConfigFileAt(testDir)

	for range 3 {
		saveGameToConfig("Test Game", f.Name(), testDir)
	}

	games := []Game{}
	err = json.Unmarshal(loadConfigFile(testDir), &games)
	if err != nil {
		t.Error(err)
	}

	var want string
	for i, game := range games {
		want = want + fmt.Sprintf("[%v] %v\n", i, game.Name)
	}

	got := captureOutput(func() {
		listGamesNumbered(testDir)
	})

	if got != want {
		t.Errorf("wanted: %v; got: %v", want, got)
	}
}

func captureOutput(f func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestRemoveGamePrompt(t *testing.T) {
	f, testDir := setupTest(t)

	err := deleteFile(filepath.Join(testDir, "config.json"))
	if err != nil {
		t.Error(err)
	}

	saveGameToConfig("Test Game", f.Name(), testDir)
	input := "0\n"
	r := bufio.NewReader(strings.NewReader(input))
	orig := programHome
	programHome = testDir

	removeGamePrompt(r)
	programHome = orig

	games := []Game{}
	json.Unmarshal(loadConfigFile(testDir), &games)
	got := len(games)
	want := 0
	if got != want {
		t.Errorf("want length=%v, got length=%v", want, got)
	}
}
