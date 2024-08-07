// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	"github.com/hashicorp/terraform-provider-azurerm/internal/provider"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-provider-azurerm/internal/provider/framework"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	// remove date and time stamp from log output as the plugin SDK already adds its own
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	var debugMode bool

	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	ctx := context.Background()

	if features.FourPointOhBeta() {
		providerServer, _, err := framework.ProtoV5ProviderServerFactory(ctx)
		if err != nil {
			log.Fatalf("creating AzureRM Provider Server: %+v", err)
		}

		var serveOpts []tf5server.ServeOpt

		if debugMode {
			serveOpts = append(serveOpts, tf5server.WithManagedDebug())
		}

		err = tf5server.Serve("registry.terraform.io/hashicorp/azurerm", providerServer, serveOpts...)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if debugMode {
			//nolint:staticcheck
			err := plugin.Debug(context.Background(), "registry.terraform.io/hashicorp/azurerm",
				&plugin.ServeOpts{
					ProviderFunc: provider.AzureProvider,
				})
			if err != nil {
				log.Println(err.Error())
			}
		} else {
			plugin.Serve(&plugin.ServeOpts{
				ProviderFunc: provider.AzureProvider,
			})
		}
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
