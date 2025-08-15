package templates

import _ "embed"

type DeploymentTemplate struct{}

//go:embed files/deployment.tmpl
var deploymentTemplate []byte

func (d DeploymentTemplate) Deployment() []byte {
	return deploymentTemplate
}
