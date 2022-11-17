package main

// Run/Build with -gcflags="-l"

import (
	"fmt"

	"bou.ke/monkey"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/provider"
)

var FakeError error

func main() {
	monkey.Patch(validation.StringInSlice, func(valid []string, ignoreCase bool) schema.SchemaValidateFunc {
		return func(i interface{}, k string) (warnings []string, errors []error) {
			return valid, []error{FakeError}
		}
	})

	for _, service := range provider.SupportedTypedServices() {
		for _, resource := range service.Resources() {
			walkSchemaMap(resource.ResourceType(), resource.Arguments())
		}
	}
	for _, service := range provider.SupportedUntypedServices() {
		for rt, resource := range service.SupportedResources() {
			walkSchemaMap(rt, resource.Schema)
		}
	}
}

func walkSchemaMap(base string, sm map[string]*schema.Schema) {
	processValidateFunc := func(addr string, f schema.SchemaValidateFunc) {
		if f != nil {
			// Some validation function might panic if passed by a nil interface{}
			defer func() {
				recover()
			}()
			warns, errors := f(nil, "")
			if len(errors) == 1 && errors[0] == FakeError {
				fmt.Printf("%s: %v\n", addr, warns)
			}
		}
	}
	for k, v := range sm {
		addr := base + "." + k
		processValidateFunc(addr, v.ValidateFunc)

		switch elem := v.Elem.(type) {
		case *schema.Schema:
			addr += ".[]"
			processValidateFunc(addr, elem.ValidateFunc)
		case *schema.Resource:
			walkSchemaMap(addr, elem.Schema)
		}
	}
}
