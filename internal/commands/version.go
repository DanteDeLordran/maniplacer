package commands

import "fmt"

const VERSION = "1.2.2"

func Help() {
	fmt.Println("Usage: maniplacer new -f <path to json>")
}

func Version() {
	fmt.Println(VERSION)
}
