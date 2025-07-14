package commands

import "fmt"

const VERSION = "1.1.0"

func Help() {
	fmt.Println("Usage: maniplacer new -f <path to json>")
}

func Version() {
	fmt.Println(VERSION)
}
