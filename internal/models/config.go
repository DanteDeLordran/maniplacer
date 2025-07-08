package models

type ManifestConfig struct {
	ConfigMapName    string `json:"configMapName"`
	GatewayGKE       string `json:"gatewayGKE"`
	HostNameK8       string `json:"hostNameK8"`
	Name             string `json:"name"`
	NameSpace        string `json:"nameSpace"`
	NameSpaceGateway string `json:"nameSpaceGateway"`
	PathBase         string `json:"pathBase"`
	PathLiveness     string `json:"pathLiveness"`
	PathReadiness    string `json:"pathReadiness"`
	PortService      string `json:"portService"`
	Replicas         string `json:"replicas"`
	MaxReplicas      uint8  `json:"maxReplicas"`
	Port             string `json:"port"`
	ReqCPU           string `json:"reqCPU"`
	ReqMemory        string `json:"reqMemory"`
	TimeoutSec       uint8  `json:"timeoutSec"`
	Image            string `json:"image"`
	TimeoutLiveness  string `json:"timeoutLiveness"`
	TimeoutReadiness string `json:"timeoutReadiness"`
	HpaMaxReplicas   uint8  `json:"hpaMaxReplicas"`
	HpaAvgCPU        string `json:"hpaAvgCPU"`
	HpaAvgMemory     string `json:"hpaAvgMemory"`
}
