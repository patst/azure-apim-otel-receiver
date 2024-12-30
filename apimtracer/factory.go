package apimtracer

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

// Real world example: https://github.com/open-telemetry/opentelemetry-collector/blob/main/receiver/otlpreceiver/factory.go
var (
	typeStr = component.MustNewType("apimtracer")
)

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTracesReceiver(_ context.Context, params receiver.Settings, baseCfg component.Config, consumer consumer.Traces) (receiver.Traces, error) {
	logger := params.Logger
	apimtracerCfg := baseCfg.(*Config)

	traceRcvr := &apimtracerReceiver{
		logger:       logger,
		nextConsumer: consumer,
		config:       apimtracerCfg,
	}

	return traceRcvr, nil
}

// NewFactory creates a factory for apimtracer receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithTraces(createTracesReceiver, component.StabilityLevelAlpha))
}
