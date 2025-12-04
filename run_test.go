package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func Test_RunNative_HappyPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		fmt.Println("GNU/Linux and Bash is required for this test. Skipping.")
		return
	}

	configFileParentDir := t.TempDir()
	pathToBashScript := filepath.Join(configFileParentDir, "script.bash")
	bashScriptString := "#!/bin/bash\n" +
		"echo \"Hello from test!\"\n" +
		"read -p \"Please enter some text: \" user_input\n" +
		"echo \"You entered: '$user_input'\"\n"
	bashScriptData := []byte(bashScriptString)

	err := os.WriteFile(pathToBashScript, bashScriptData, os.ModePerm)
	if err != nil {
		t.Fatalf("write file failed: '%v'\n", err)
	}

	err = os.Chmod(pathToBashScript, 0700)
	if err != nil {
		t.Fatalf("chmod failed: '%v'\n", err)
	}

	// The bash script is the executable of the added game
	err = saveGameToConfig("Test Game", pathToBashScript, configFileParentDir)
	if err != nil {
		t.Fatalf("save game failed: '%v'\n", err)
	}

	var writer bytes.Buffer
	userInput := bufio.NewReader(strings.NewReader("Test user input\n"))
	_, err = runNative(pathToBashScript, userInput, &writer)
	if err != nil {
		t.Fatalf("got: '%v'; want nil;\n", err)
	}

	capturedOutput := writer.String()
	wantOutputString := "Hello from test!\nYou entered: 'Test user input'\n"
	if capturedOutput != wantOutputString {
		t.Fatalf("got: \n%s\nwant: \n%s\n", capturedOutput, wantOutputString)
	}
}
