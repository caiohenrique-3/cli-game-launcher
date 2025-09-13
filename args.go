package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

type ListCmdOptions struct {
	showHidden          bool
	onlyShowHidden      bool
	hiddenGameIndicator bool // Shows [HIDDEN] besides the name of the game
}

type RunCmdOptions struct {
	showHidden bool
}

func handleCommandLine() error {
	expectedSubcommandsMsg := "expected 'add', 'remove', 'list', 'hide', 'run' or 'help' subcommand\n"

	if len(os.Args) < 2 {
		fmt.Print(expectedSubcommandsMsg)
		os.Exit(1)
	}

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listShowsHiddenGames := listCmd.Bool("show-hidden", false, "Show hidden games in output")
	listOnlyShowsHiddenGames := listCmd.Bool("only-hidden", false, "Only show hidden games in output")

	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	runShowsHiddenGames := runCmd.Bool("show-hidden", false, "Show hidden games in output")

	switch os.Args[1] {
	case "--help", "-h", "help":
		showHelp()
		return nil
	case "add":
		err := addGamePrompt(bufio.NewReader(os.Stdin))
		if err != nil {
			return err
		}
	case "remove":
		err := removeGamePrompt(bufio.NewReader(os.Stdin))
		if err != nil {
			return err
		}
	case "list":
		err := listCmd.Parse(os.Args[2:])
		if err != nil {
			return fmt.Errorf("parse failed: %v", err)
		}

		cmdOptions := &ListCmdOptions{
			showHidden:     *listShowsHiddenGames,
			onlyShowHidden: *listOnlyShowsHiddenGames,
		}

		err = listGamesNumbered(programHome, os.Stdout, *cmdOptions)
		if err != nil {
			return err
		}
	case "run":
		err := runCmd.Parse(os.Args[2:])
		if err != nil {
			return fmt.Errorf("parse failed: %v", err)
		}

		cmdOptions := &RunCmdOptions{
			showHidden: *runShowsHiddenGames,
		}

		err = runGamePrompt(bufio.NewReader(os.Stdin), *cmdOptions)
		if err != nil {
			return err
		}
	case "hide":
		err := hideGamePrompt(bufio.NewReader(os.Stdin))
		if err != nil {
			return err
		}
	default:
		fmt.Print(expectedSubcommandsMsg)
		os.Exit(1)
	}

	return nil
}
