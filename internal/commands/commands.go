package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/dantedelordran/maniplacer/internal/models"
)

func Help() {
	fmt.Println("Usage: maniplacer new -f <path to json>")
}

func NewManifest() {
	cmd := flag.NewFlagSet("new", flag.ExitOnError)
	file := cmd.String("f", "", "Path to JSON config file")

	cmd.Parse(os.Args[2:])

	if *file == "" {
		fmt.Println("Error: -f flag is required")
		cmd.Usage()
		os.Exit(1)
	}

	config, err := loadConfig(*file)

	if err != nil {
		fmt.Println("Error loading config due to ", err)
		os.Exit(1)
	}

	fmt.Println(config)

}

func loadConfig(path string) (*models.ManifestConfig, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("Failed to load file with specified path")
		return nil, err
	}

	var config models.ManifestConfig
	err = json.Unmarshal(data, &config)
	return &config, nil
}
