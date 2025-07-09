# Maniplacer

Simple YAML K8s replacer from JSON file

## How to install

```bash
curl -fsSL https://raw.github.com/dantedelordran/maniplacer/main/installer.sh | bash
```


## JSON structure

```javascript
{
    "gatewayGKE": "",
    "hostNameK8": "",
    "name": "",
    "nameSpace": "",
    "nameSpaceGateway": "",
    "pathBase": "",
    "pathLiveness": "",
    "pathReadiness": "",
    "portService": "",
    "replicas": "",
    "maxReplicas": "",
    "port": "",
    "reqCPU": "",
    "reqMemory": "",
    "timeoutSec": "",
    "image": "",
    "timeoutLiveness": "",
    "timeoutReadiness": "",
    "hpaAvgCPU": "",
    "hpaAvgMemory": "",
    "secrets": {   <---- OPTIONAL
        "SOME_SECRET": "some value",
    }
}
```