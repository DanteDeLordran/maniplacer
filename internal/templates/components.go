package templates

type component string

const (
	deployment component = "deployment"
	service    component = "service"
	httpRoute  component = "httproute"
	secret     component = "secret"
	configmap  component = "configmap"
	hpa        component = "hpa"
	hcpolicy   component = "hcpolicy"
)

var AllowedComponents = []string{string(deployment), string(service), string(httpRoute), string(secret), string(configmap), string(hpa), string(hcpolicy)}

var TemplateRegistry = map[string][]byte{
	string(deployment): deploymentTemplate,
	string(service):    serviceTemplate,
	string(httpRoute):  httpRouteTemplate,
	string(secret):     secretTemplate,
	string(configmap):  configMapTemplate,
	string(hpa):        hpaTemplate,
	string(hcpolicy):   hcPolocyTemplate,
}
