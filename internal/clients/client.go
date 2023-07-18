// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package clients

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
	"github.com/hashicorp/terraform-provider-azurerm/internal/features"
	network "github.com/hashicorp/terraform-provider-azurerm/internal/services/network/client"
)

type Client struct {
	autoClient

	// StopContext is used for propagating control from Terraform Core (e.g. Ctrl/Cmd+C)
	StopContext context.Context

	Account  *ResourceManagerAccount
	Features features.UserFeatures

	Network *network.Client
}

// NOTE: it should be possible for this method to become Private once the top level Client's removed

func (client *Client) Build(ctx context.Context, o *common.ClientOptions) error {
	autorest.Count429AsRetry = false
	// Disable the Azure SDK for Go's validation since it's unhelpful for our use-case
	validation.Disabled = true

	if err := buildAutoClients(&client.autoClient, o); err != nil {
		return fmt.Errorf("building auto-clients: %+v", err)
	}

	client.Features = o.Features
	client.StopContext = ctx

	var err error

	if client.Network, err = network.NewClient(o); err != nil {
		return fmt.Errorf("building clients for Network: %+v", err)
	}

	return nil
}
