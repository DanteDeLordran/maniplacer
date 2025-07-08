# Maniplacer

Simple YAML K8s replacer from JSON file

## How to install

```bash
curl -fsSL https://raw.github.com/dantedelordran/maniplacer/main/installer.sh | bash
```


## JSON structure

```javascript
{
    "configMapName": "",
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
    "maxReplicas": 1,
    "port": "",
    "reqCPU": "",
    "reqMemory": "",
    "timeoutSec": 1,
    "image": "",
    "timeoutLiveness": "",
    "timeoutReadiness": "",
    "hpaMaxReplicas": 3,
    "hpaAvgCPU": "",
    "hpaAvgMemory": ""
}
```