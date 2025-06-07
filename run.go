package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// Executes a command and prints the output.
func runNative(command string, reader *bufio.Reader, writer io.Writer) (time.Duration, error) {
	cmd := exec.Command(command)
	cmd.Stdout = writer
	cmd.Stdin = reader
	cmd.Stderr = cmd.Stdout

	timeStart := time.Now()
	err := cmd.Run()
	if err != nil {
		return time.Duration(0), fmt.Errorf("command failed: %w", err)
	}
	timeSpentPlaying := time.Since(timeStart)

	return timeSpentPlaying, nil
}
