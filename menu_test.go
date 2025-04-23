package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func TestAddGameExecutableGoodEnding(t *testing.T) {
	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Error(err)
	}

	input := "Test Game Name\n" + f.Name() + "\n"
	r := bufio.NewReader(strings.NewReader(input))
	addGamePrompt(r)
}

func TestAddGameExecutablePathWithSpaces(t *testing.T) {
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
	addGamePrompt(r)
}

func TestAddGameExecutablePathIsADirectory(t *testing.T) {
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
		"-test.run=TestAddGameExecutablePathIsADirectory")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v; want exit status 1", err)
}

func TestAddGameExecutableExpandsEnvVariables(t *testing.T) {
	if os.Getenv("TEST_TMPDIR") != "" {
		f, err := createTempFileOnOsTempDir()
		if err != nil {
			t.Error("error creating temp file:", err)
		}

		fileName := filepath.Base(f.Name())

		input := "Test Game Name\n" +
			"$TEST_TMPDIR/" + fileName + "\n"
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
		return
	}

	cmd := exec.Command(os.Args[0],
		"-test.run=TestAddGameExecutableExpandsEnvVariables")
	cmd.Env = append(os.Environ(), "TEST_TMPDIR="+os.TempDir())
	err := cmd.Run()
	if err != nil {
		t.Fatalf("process ran with err %v; want exit status 0", err)
	}
}

func TestGetCleanPathExpandsTilde(t *testing.T) {
	if runtime.GOOS != "linux" {
		return
	}

	pathInput := "~/"
	s := getCleanPath(pathInput)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
	}

	if s != home {
		t.Error("tilde did not expand user home dir!")
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
