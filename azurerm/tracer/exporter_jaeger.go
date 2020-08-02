package tracer

import (
	"fmt"
	"os"

	"contrib.go.opencensus.io/exporter/jaeger"
)

const (
	envTfAzureTraceJaegerCollectionEndpointUri = "TF_AZURE_TRACE_JAEGER_COLLECTION_ENDPOINT_URI"
	envTfAzureTraceJaegerAgentEndpointUri      = "TF_AZURE_TRACE_JAEGER_AGENT_ENDPOINT_URI"
)

func buildJaegerExporter(serviceName string) (FlushableExporter, error) {
	agentUri := os.Getenv(envTfAzureTraceJaegerAgentEndpointUri)
	if agentUri == "" {
		return nil, fmt.Errorf("please specify %s as environemnt variable", envTfAzureTraceJaegerAgentEndpointUri)
	}
	collectionUri := os.Getenv(envTfAzureTraceJaegerCollectionEndpointUri)
	if collectionUri == "" {
		return nil, fmt.Errorf("please specify %s as environemnt variable", envTfAzureTraceJaegerCollectionEndpointUri)
	}

	e, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint:     agentUri,
		CollectorEndpoint: collectionUri,
		Process: jaeger.Process{
			ServiceName: serviceName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize exporter: %w", err)
	}

	return e, nil
}
