package main

import (
	"context"
	"log"

	opencensusTrace "go.opencensus.io/trace"

	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/tracer"
)

func main() {
	// remove date and time stamp from log output as the plugin SDK already adds its own
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	if tracer.TracingEnabled() {
		tracer.Init()
		// create the first root span, this span has the same lifetime as the plugin server
		_, tracer.RootSpan = opencensusTrace.StartSpan(context.Background(), "ROOT SPAN")
		// [WORKAROUND]
		// If put this as the last command of main(), jaeger will complain about "invalid parent span",
		// not sure why it is.
		tracer.RootSpan.End()
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: azurerm.Provider,
	})

	if tracer.TracingEnabled() {
		// Some exporter (e.g. jaeger) works in a async manner. Hence need refresh.
		tracer.Exporter.Flush()
	}
}
