# OpenTelemetry Trace Receiver

see https://opentelemetry.io/docs/collector/building/receiver/


Consume messages from EventHub using Golang:
https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-go-get-started-send 


Install the opentelemetry builder cli:

```bash
curl --proto '=https' --tlsv1.2 -fL -o ocb \
https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/cmd%2Fbuilder%2Fv0.115.0/ocb_0.115.0_darwin_arm64
chmod +x ocb
```

## Build

Either use Paketo buildpacks:

```bash
pack build apim-trace-collector --buildpack paketo-buildpacks/go --env BP_GO_WORK_USE=./apimtracer:./otelcol-dev --env BP_GO_TARGETS=./otelcol-dev --env BP_GO_VERSION=1.23.3
```

Or a plain Dockerfile:
```bash
docker build -t azure-apim-otel-reciever:latest .
```

## Deploy

Use the kubernetes manifests in the `manifesst` folder:
```bash
kubectl apply -f manifests/
```

## Configure Azure API-Management policy

The policy definition is in the `api-policy.xml` file and must be configured for the API in Azure API-Management.
