package templates

import _ "embed"

type Template struct{}

//go:embed files/deployment.tmpl
var deploymentTemplate []byte

func (t Template) Deployment() []byte {
	return deploymentTemplate
}

//go:embed files/service.tmpl
var serviceTemplate []byte

func (t Template) Service() []byte {
	return serviceTemplate
}

//go:embed files/httproute.tmpl
var httpRouteTemplate []byte

func (t Template) HttpRoute() []byte {
	return httpRouteTemplate
}

//go:embed files/secret.tmpl
var secretTemplate []byte

func (t Template) Secret() []byte {
	return secretTemplate
}

//go:embed files/configmap.tmpl
var configMapTemplate []byte

func (t Template) ConfigMap() []byte {
	return configMapTemplate
}

//go:embed files/hpa.tmpl
var hpaTemplate []byte

func (t Template) HPA() []byte {
	return hpaTemplate
}

//go:embed files/hcpolicy.tmpl
var hcPolocyTemplate []byte

func (t Template) HcPolicy() []byte {
	return hcPolocyTemplate
}
