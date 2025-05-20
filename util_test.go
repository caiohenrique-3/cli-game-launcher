package main

import "testing"

func TestFileExistsTrue(t *testing.T) {
	f, err := createTempFileOnOsTempDir()
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}

	b := fileExists(f.Name())
	if !b {
		t.Errorf("got '%t'; want '%t'\n", b, true)
	}
}

func TestExistsFalse(t *testing.T) {
	b := fileExists("/mint/dragon/path/three")
	if b {
		t.Errorf("got: '%t'; want '%t'\n", b, false)
	}
}
