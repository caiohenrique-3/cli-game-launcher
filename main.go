package main

import (
	"fmt"

	"github.com/caiohenrique-3/cli-game-launcher/run"
)

func main() {
	// Get executable path from somewhere
	// Then pass it to runNative()
	err := run.RunNative()
	if err != nil {
		fmt.Println(err)
	}
}
