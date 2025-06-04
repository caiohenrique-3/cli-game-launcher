package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

// Executes a command and prints the output.
func runNative(command string, reader *bufio.Reader, writer io.Writer) error {
	cmd := exec.Command(command)
	cmd.Stdout = writer
	cmd.Stdin = reader
	cmd.Stderr = cmd.Stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}
