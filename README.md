# Maniplacer

Simple YAML K8s replacer from JSON file

## JSON structure

```javascript
{
    "configMapName": "",
    "gatewayGKE": "",
    "hostNameK8": "",
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