# syntax=docker/dockerfile:1

FROM golang:1.23 AS build-stage

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY ./ ./

# Build
RUN cd otelcol-dev && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /apim-otel-collector

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /apim-otel-collector /apim-otel-collector

 # user id 65532
USER nonroot:nonroot

# Run
ENTRYPOINT ["/apim-otel-collector"]
