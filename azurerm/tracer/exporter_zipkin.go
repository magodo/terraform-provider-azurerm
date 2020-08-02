package tracer

import (
	"fmt"
	"os"

	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
)

const (
	envTfAzureTraceZipkinEndpointUri     = "TF_AZURE_TRACE_ZIPKIN_ENDPOINT_URI"
	envTfAzureTraceZipkinHttpReporterUri = "TF_AZURE_TRACE_ZIPKIN_HTTP_REPORTER_URI"
)

type zipkinExporter struct {
	*zipkin.Exporter
}

func (e *zipkinExporter) Flush() {}

func buildZipkinExporter(serviceName string) (FlushableExporter, error) {
	endpointUri := os.Getenv(envTfAzureTraceZipkinEndpointUri)
	if endpointUri == "" {
		return nil, fmt.Errorf("please specify %s as environemnt variable", envTfAzureTraceZipkinEndpointUri)
	}
	reporterUri := os.Getenv(envTfAzureTraceZipkinHttpReporterUri)
	if reporterUri == "" {
		return nil, fmt.Errorf("please specify %s as environemnt variable", envTfAzureTraceZipkinHttpReporterUri)
	}

	localEndpoint, err := openzipkin.NewEndpoint(serviceName, endpointUri)
	if err != nil {
		return nil, fmt.Errorf("failed to create the local zipkinEndpoint: %w", err)
	}
	reporter := zipkinHTTP.NewReporter(reporterUri)
	return &zipkinExporter{zipkin.NewExporter(reporter, localEndpoint)}, nil
}
