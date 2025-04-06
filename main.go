package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Executes a command and prints the output.
func runNative() error {
	cmd := exec.Command("./game.sh")
	stdout, err := cmd.StdoutPipe()
	cmd.Stdin = os.Stdin
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	// https://stackoverflow.com/a/62630988
	// Hmmm...
	for {
		tmp := make([]byte, 1)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	return nil
}

func main() {
	// Get executable path from somewhere
	// Then pass it to runNative()
	runNative()
}
