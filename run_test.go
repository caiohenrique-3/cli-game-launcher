package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
)

func Test_RunNative_HappyPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		fmt.Println("GNU/Linux and Bash is required for this test. Skipping.")
		return
	}
	setupTestsDir(t)
	configFileParentDir := createTempDirWithConfigFile(t, nil)
	pathToBashScript := createTempFileForTests(t)
	bashScriptString := "#!/bin/bash\n" +
		"echo \"Hello from test!\"\n" +
		"read -p \"Please enter some text: \" user_input\n" +
		"echo \"You entered: '$user_input'\"\n"
	bashScriptData := []byte(bashScriptString)

	err := os.WriteFile(pathToBashScript, bashScriptData, os.ModePerm)
	if err != nil {
		t.Errorf("error writing to bash script: %v\n", err)
	}

	err = os.Chmod(pathToBashScript, 0700)
	if err != nil {
		t.Errorf("error changing permissions of bash script: %v\n", err)
	}

	// The bash script is the executable of the added game
	err = saveGameToConfig("Test Game", pathToBashScript, configFileParentDir)
	if err != nil {
		t.Errorf("error saving game to config: %v\n", err)
	}

	var writer bytes.Buffer
	reader := bufio.NewReader(strings.NewReader("Test data\n"))
	err = runNative(pathToBashScript, reader, &writer)
	if err != nil {
		t.Errorf("got: '%v'; want nil;\n", err)
	}

	capturedOutput := writer.String()
	wantOutputString := "Hello from test!\n" +
		"You entered: 'Test data'\n"
	if capturedOutput != wantOutputString {
		t.Errorf("got: \n%s\nwant: \n%s\n", capturedOutput, wantOutputString)
	}
}
