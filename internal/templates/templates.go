package templates

import _ "embed"

// The embedded component scaffolds. They are exposed through TemplateRegistry
// in components.go, which is what the add command looks components up in.

//go:embed files/deployment.tmpl
var deploymentTemplate []byte

//go:embed files/service.tmpl
var serviceTemplate []byte

//go:embed files/httproute.tmpl
var httpRouteTemplate []byte

//go:embed files/secret.tmpl
var secretTemplate []byte

//go:embed files/configmap.tmpl
var configMapTemplate []byte

//go:embed files/hpa.tmpl
var hpaTemplate []byte

//go:embed files/hcpolicy.tmpl
var hcPolicyTemplate []byte
