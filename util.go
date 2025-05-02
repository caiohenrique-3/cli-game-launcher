package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Returns false if the path is a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// If file exists, deletes it.
func deleteFile(path string) error {
	// TODO: Test when dir is not empty
	if fileExists(path) {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

// Removes backslashes from path (if not Windows), expands variables,
// expands tilde and makes the path absolute.
func getCleanPath(s string) string {
	if strings.Contains(s, "$") {
		s = os.ExpandEnv(s)
	}
	if strings.HasPrefix(s, "~/") {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		s = filepath.Join(userHomeDir, s[1:])

	}

	if !filepath.IsAbs(s) {
		s, err := filepath.Abs(s)
		if err != nil {
			log.Fatal(err)
		}

		return s
	}
	return s
}

