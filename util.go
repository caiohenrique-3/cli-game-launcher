package main

import (
	"fmt"
	"os"
	"path/filepath"
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
func getAbsolutePath(s string) (string, error) {
	if !filepath.IsAbs(s) {
		s, err := filepath.Abs(s)
		if err != nil {
			return s, fmt.Errorf("get absolute path failed: %w", err)
		}

		return s, nil
	}
	return s, nil
}
