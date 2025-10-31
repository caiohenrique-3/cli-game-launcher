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
	timeSpentPlaying := time.Since(timeStart)
	if err != nil {
		return timeSpentPlaying, fmt.Errorf("command failed: %w", err)
	}

	return timeSpentPlaying, nil
}
