package tracer

import (
	"go.opencensus.io/trace"
)

type FlushableExporter interface {
	trace.Exporter
	Flush()
}
