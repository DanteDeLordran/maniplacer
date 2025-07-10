# Maniplacer

Simple YAML K8s replacer from JSON file

## How to install

```bash
curl -fsSL https://raw.github.com/dantedelordran/maniplacer/main/installer.sh | bash
```

## How to use

1. Generate a JSON file with the structure from below
2. Fill with the required values
3. Type ```maniplacer new -f <path to your JSON file>``` to create a new YAML template
4. Template will be saved on ```$HOME/maniplacer```
5. Null ```pathBase``` will skip HTTPRoute component generation
6. Null ```secrets``` will skip Secrets component generation

## JSON structure

```javascript
{
    "gatewayGKE": "",
    "hostNameK8": "",
    "name": "",
    "nameSpace": "",
    "nameSpaceGateway": "",
    "pathBase": "", <---- OPTIONAL
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