package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Creates a dir and a file inside it with name + random numbers
func setupAddGamePromptTest(t *testing.T, name string) *os.File {
	d, err := os.MkdirTemp("", name)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.CreateTemp(d, name)
	if err != nil {
		t.Fatal("creating temp file on dir with emojis:", err)
	}
	f.Close()

	return f
}

func TestAddGamePrompt(t *testing.T) {
	deleteConfigFileFromTempDir(t)

	f1, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Fatal(err)
	} // f1.Close already called

	fileWithSpaces := setupAddGamePromptTest(t, "With Some Spaces")
	fileWithEmojis := setupAddGamePromptTest(t, "😀😃😄😁🤣🥲🥹☺️")
	fileWithCJK := setupAddGamePromptTest(t, "史诗 テスト 파일")

	os.Setenv("TEST_TMPDIR", os.TempDir())

	orig := programHome
	programHome = os.TempDir()
	defer func() {
		programHome = orig
	}()

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
		"good ending": {
			input:  fmt.Sprintf("test\n%s\n", f1.Name()),
			result: nil},
		"path with spaces": {
			input:  fmt.Sprintf("test\n%s\n", fileWithSpaces.Name()),
			result: nil},
		"path is a directory": {
			input:  fmt.Sprintf("test\n%s\n", os.TempDir()),
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

			sout := os.Stdout
			serr := os.Stderr
			os.Stdout = nil
			os.Stderr = nil

			got := addGamePrompt(r)
			want := test.result

			os.Stdout = sout
			os.Stderr = serr

			if !errors.Is(got, want) {
				t.Fatalf("got %v; want %v;", got, want)
			}
		})
	}
}

func TestListGamesNumbered(t *testing.T) {
	deleteConfigFileFromTempDir(t)
	err := createEmptyConfigFileAt(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	emptyDir, err := os.MkdirTemp("", "emptydir")
	if err != nil {
		t.Fatal(err)
	}

	badJsonDir, err := os.MkdirTemp("", "badJson")
	if err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(badJsonDir, "config.json")
	err = os.WriteFile(f, []byte("TEST1"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	emptyConfigDir, err := os.MkdirTemp("", "holdingEmptyConfigFile")
	if err != nil {
		t.Fatal(err)
	}

	err = createEmptyConfigFileAt(emptyConfigDir)
	if err != nil {
		t.Fatal(err)
	}

	// Reusing created file for game executable
	err = saveGameToConfig("Test Game", f, os.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	errTests := map[string]struct {
		input  string
		result error
	}{
		"config file not found": {
			input:  emptyDir,
			result: ErrDoesNotExistOrIsADirectory},
		"bad json": {
			input:  badJsonDir,
			result: &json.SyntaxError{Offset: 0},
		},
	}

	outTests := map[string]struct {
		input  string
		result string
	}{
		"good ending": {
			input:  os.TempDir(),
			result: fmt.Sprintf("[0] Test Game\n"),
		},
		"empty config": {
			input:  emptyConfigDir,
			result: "No games found\n",
		},
	}

	for name, test := range errTests {
		t.Run(name, func(t *testing.T) {
			sout := os.Stdout
			serr := os.Stderr
			os.Stdout = nil
			os.Stderr = nil

			got := listGamesNumbered(test.input)
			want := test.result

			os.Stdout = sout
			os.Stderr = serr

			// HACK: Only checks the message, but not the type
			if name == "bad json" {
				msg := "invalid character 'T' looking for beginning of value"
				if !strings.Contains(got.Error(), msg) {
					t.Errorf("got '%v'; want '%v';\n", got, want)
				}
			} else {
				if !errors.Is(got, want) {
					t.Errorf("got '%v'; want '%v';\n", got, want)
				}
			}

		})
	}

	for name, test := range outTests {
		t.Run(name, func(t *testing.T) {
			got := captureOutput(test.input, listGamesNumbered)
			want := test.result

			if got != want {
				t.Errorf("got '%v'; want '%v';\n", got, want)
			}

		})
	}
}

// Only for listGamesNumbered
func captureOutput(input string, f func(s string) error) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f(input)
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestRemoveGamePrompt(t *testing.T) {
	deleteConfigFileFromTempDir(t)
	f, d := setupTest(t)
	for range 10 {
		err := saveGameToConfig("Test", f.Name(), d)
		if err != nil {
			t.Error(err)
		}
	}

	orig := programHome
	programHome = os.TempDir()
	defer func() {
		programHome = orig
	}()

	tests := map[string]struct {
		input  string
		result error
	}{
		"empty string": {
			input:  "\n",
			result: strconv.ErrSyntax},
		"good ending": {
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
		// TODO: should be ErrInvalidOption from config.go
		// "less than zero": {
		// 	input:  "-1\n",
		// 	result: nil},
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

			sout := os.Stdout
			serr := os.Stderr
			os.Stdout = nil
			os.Stderr = nil

			got := removeGamePrompt(r)
			want := test.result

			os.Stdout = sout
			os.Stderr = serr

			if !errors.Is(got, want) {
				t.Fatalf("got %v; want %v;", got, want)
			}
		})
	}
}
