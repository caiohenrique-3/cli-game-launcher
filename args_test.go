package main

import (
	"os"
	"strings"
	"testing"
)

var cleanArgs = getCleanOsArgs()

// FIXME: EOF error because of "add" subcommand
func TestHandleCommandLine(t *testing.T) {
	valid := []string{"--help", "-h", "add"}
	invalid := []string{"--lpdasl", "mbmbd"}

	// Dont print menu output
	os.Stdout = nil

	err := checkCommandLine(valid, true)
	if err != nil {
		t.Errorf("ARGS='%s' ERROR: %s", os.Args, err)
	}

	err = checkCommandLine(invalid, false)
	if err != nil {
		t.Errorf("ARGS='%s' ERROR: %s", os.Args, err)
	}
}

// Removes test flags from os.Args so that readArgs() can work.
func getCleanOsArgs() []string {
	// https://stackoverflow.com/a/20551116
	newArgs := os.Args
	i := 0
	for _, argStr := range newArgs {
		if !strings.Contains(argStr, "-test") {
			newArgs[i] = argStr
			i++
		}
	}
	newArgs = newArgs[:i]
	return newArgs
}

func checkCommandLine(s []string, isValid bool) error {
	for _, optionStr := range s {
		os.Args = cleanArgs
		os.Args = append(os.Args, optionStr)

		err := handleCommandLine()
		if err != nil && isValid {
			return err
		}
	}
	return nil
}
