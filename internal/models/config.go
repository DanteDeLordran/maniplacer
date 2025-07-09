package models

type ManifestConfig struct {
	GatewayGKE       string            `json:"gatewayGKE"`
	HostNameK8       string            `json:"hostNameK8"`
	Name             string            `json:"name"`
	NameSpace        string            `json:"nameSpace"`
	NameSpaceGateway string            `json:"nameSpaceGateway"`
	PathBase         string            `json:"pathBase"`
	PathLiveness     string            `json:"pathLiveness"`
	PathReadiness    string            `json:"pathReadiness"`
	PortService      string            `json:"portService"`
	Replicas         string            `json:"replicas"`
	MaxReplicas      string            `json:"maxReplicas"`
	Port             string            `json:"port"`
	ReqCPU           string            `json:"reqCPU"`
	ReqMemory        string            `json:"reqMemory"`
	TimeoutSec       string            `json:"timeoutSec"`
	Image            string            `json:"image"`
	TimeoutLiveness  string            `json:"timeoutLiveness"`
	TimeoutReadiness string            `json:"timeoutReadiness"`
	HpaMaxReplicas   string            `json:"hpaMaxReplicas"`
	HpaAvgCPU        string            `json:"hpaAvgCPU"`
	HpaAvgMemory     string            `json:"hpaAvgMemory"`
	Secrets          map[string]string `json:"secrets,omitempty"`
}
