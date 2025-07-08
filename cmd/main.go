package main

import (
	"fmt"
	"os"

	"github.com/dantedelordran/maniplacer/internal/commands"
)

func main() {

	if len(os.Args) < 2 {
		commands.Help()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		commands.Help()
	case "version", "-v":
		commands.Version()
	case "new":
		commands.NewManifest()
	case "update":
		commands.AutoUpdate()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		commands.Help()
		os.Exit(1)
	}

}
