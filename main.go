package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type ManifestConfig struct {
	ConfigMapName    string `json:"configMapName"`
	GatewayGKE       string `json:"gatewayGKE"`
	HostNameK8       string `json:"hostNameK8"`
	NameSpace        string `json:"nameSpace"`
	NameSpaceGateway string `json:"nameSpaceGateway"`
	PathBase         string `json:"pathBase"`
	PathLiveness     string `json:"pathLiveness"`
	PathReadiness    string `json:"pathReadiness"`
	PortService      string `json:"portService"`
	Replicas         string `json:"replicas"`
	MaxReplicas      int64  `json:"maxReplicas"`
	Port             string `json:"port"`
	ReqCPU           string `json:"reqCPU"`
	ReqMemory        string `json:"reqMemory"`
	TimeoutSEC       int64  `json:"timeoutSec"`
	Image            string `json:"image"`
	TimeoutLiveness  string `json:"timeoutLiveness"`
	TimeoutReadiness string `json:"timeoutReadiness"`
	HpaMaxReplicas   int64  `json:"hpaMaxReplicas"`
	HpaAvgCPU        string `json:"hpaAvgCPU"`
	HpaAvgMemory     string `json:"hpaAvgMemory"`
}

func main() {
	fmt.Println("Lordran maniplacer")

	if len(os.Args) < 2 {
		help()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "new":
		newManifest()
	case "help", "-h", "--help":
		help()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		help()
		os.Exit(1)
	}

}

func help() {
	fmt.Println("Usage: maniplacer new -f <path to json>")
}

func newManifest() {
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

func loadConfig(path string) (*ManifestConfig, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("Failed to load file with specified path")
		return nil, err
	}

	var config ManifestConfig
	err = json.Unmarshal(data, &config)
	return &config, nil
}
