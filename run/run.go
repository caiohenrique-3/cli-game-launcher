// Contains functions related to running native programs.
package run

import (
	"fmt"
	"os"
	"os/exec"
)

// Executes a command and prints the output.
func RunNative() error {
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
		tmp := make([]byte, 32)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}

	return nil
}
