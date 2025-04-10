package main

import "os"

func main() {
	// Get executable path from somewhere
	// Then pass it to runNative()
	err := handleCommandLine()
	if err != nil {
		os.Exit(1)
	}

	// Pass a full comand to RunNative here,
	// eg. "sbx run instance"
	// then split it and pass
	/* 	err := run.RunNative()
	   	if err != nil {
	   		fmt.Println(err)
	   	} */
}
