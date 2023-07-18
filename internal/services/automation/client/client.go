// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"fmt"

	// nolint: staticcheck

	// use new sdk once https://github.com/hashicorp/pandora/issues/2756 fixed

	// hybridrunbookworkergroup v2022-08-08 issue: https://github.com/Azure/azure-rest-api-specs/issues/24740

	"github.com/hashicorp/go-azure-sdk/resource-manager/automation/2022-08-08/schedule"
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
)

type Client struct {
	ScheduleClient *schedule.ScheduleClient
}

func NewClient(o *common.ClientOptions) (*Client, error) {
	scheduleClient, err := schedule.NewScheduleClientWithBaseURI(o.Environment.ResourceManager)
	if err != nil {
		return nil, fmt.Errorf("build scheduleClient: %+v", err)
	}
	o.Configure(scheduleClient.Client, o.Authorizers.ResourceManager)

	return &Client{
		ScheduleClient: scheduleClient,
	}, nil
}
