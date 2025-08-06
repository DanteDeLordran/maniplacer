package commands

import "fmt"

const VERSION = "1.3.2"

func Help() {
	fmt.Print(`maniplacer - A CLI tool for generating Kubernetes manifests

Usage:
  maniplacer <command> [options]

Commands:
  new         Generate a manifest from a JSON config
  version     Show the version of the tool
  help        Show this help message

Options for 'new':
  -f string    Path to the JSON configuration file (required)
  -t string    Target directory to save the manifest (optional, default: ~/maniplacer)

Examples:
  maniplacer new -f config.json
  maniplacer new -f config.json -t ~/custom/path
  maniplacer version
  maniplacer help

JSON Config Format:
  {
    "Name": "my-app",
    "NameSpace": "default",
    ...
  }

Manifest Template:
  Uses an embedded YAML file with Go template syntax.
  Base64 encoding available via 'b64enc' function in template.

Website & Docs:
  https://github.com/dantedelordran/maniplacer
`)
}

func Version() {
	fmt.Println(VERSION)
}
