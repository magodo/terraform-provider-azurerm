// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/network"
)

//go:generate go run ../tools/generator-services/main.go -path=../../

func SupportedTypedServices() []sdk.TypedServiceRegistration {
	services := []sdk.TypedServiceRegistration{
		network.Registration{},
	}
	services = append(services, autoRegisteredTypedServices()...)
	return services
}

func SupportedUntypedServices() []sdk.UntypedServiceRegistration {
	return func() []sdk.UntypedServiceRegistration {
		out := []sdk.UntypedServiceRegistration{}
		return out
	}()
}
