package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Executes a command and prints the output.
func runNative(command string) error {
	cmd := exec.Command(command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("command stdout pipe failed: %w", err)
	}

	cmd.Stdin = os.Stdin
	cmd.Stderr = cmd.Stdout

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	for {
		tmp := make([]byte, 128)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	return nil
}
