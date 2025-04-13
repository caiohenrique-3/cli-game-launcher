package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestAddGamePromptFileNotFound(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		input := "Test\nTest\n"
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
		return
	}
	fmt.Println(os.Stdout == nil)
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
	input := "Test Game Name\n./menu_test.go\n"
	r := bufio.NewReader(strings.NewReader(input))
	addGamePrompt(r)
}

func TestAddGameExecutablePathWithSpaces(t *testing.T) {
	input := "Test Game Name\n./tests/Dir With Spaces/game\n"
	r := bufio.NewReader(strings.NewReader(input))
	addGamePrompt(r)
}

func TestAddGameExecutablePathWithBackslash(t *testing.T) {
	input := "Test Game Name\n./tests/Dir\\ With\\ Spaces/game\n"
	r := bufio.NewReader(strings.NewReader(input))
	addGamePrompt(r)
}

func TestAddGameExecutablePathIsADirectory(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		input := "Test Game Name\n./tests/\n"
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
		return
	}
	fmt.Println(os.Stdout == nil)
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
	input := "Test Game Name\n$HOME/.zshrc\n"
	r := bufio.NewReader(strings.NewReader(input))
	addGamePrompt(r)
}

func TestAddGameExecutableExpandsTilde(t *testing.T) {
	input := "Test Game Name\n~/.zshrc\n"
	r := bufio.NewReader(strings.NewReader(input))
	addGamePrompt(r)
}

func TestFileExistsTrue(t *testing.T) {
	b := fileExists("./menu_test.go")
	if b != true {
		t.Errorf("error: %v; want true", b)
	}
}

func TestExistsFalse(t *testing.T) {
	b := fileExists("/mint/dragon/path/three")
	if b != false {
		t.Errorf("error: %v; want false", b)
	}
}

func BenchmarkAddGameExecutable(b *testing.B) {
	os.Stdout = nil
	os.Stderr = nil
	input := "Test\nmenu_test.go\n"
	for b.Loop() {
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
	}
}

func BenchmarkAddGamePathWithEnvVariable(b *testing.B) {
	os.Stdout = nil
	os.Stderr = nil
	input := "Test\n$HOME/.zshrc\n"
	for b.Loop() {
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
	}
}

func BenchmarkAddGamePathWithTilde(b *testing.B) {
	os.Stdout = nil
	os.Stderr = nil
	input := "Test\n~/.zshrc\n"
	for b.Loop() {
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
	}
}

func BenchmarkAddGamePathWithSpaces(b *testing.B) {
	os.Stdout = nil
	os.Stderr = nil
	input := "Test\n./tests/Dir With Spaces/game\n"
	for b.Loop() {
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
	}
}

func BenchmarkAddGamePathWithSpacesAndBackslashes(b *testing.B) {
	os.Stdout = nil
	os.Stderr = nil
	input := "Test\n./tests/Dir\\ With\\ Spaces/game\n"
	for b.Loop() {
		r := bufio.NewReader(strings.NewReader(input))
		addGamePrompt(r)
	}
}

func BenchmarkExists(b *testing.B) {
	os.Stdout = nil
	os.Stderr = nil
	for b.Loop() {
		fileExists("./menu_test.go")
	}
}
