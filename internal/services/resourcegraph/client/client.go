package client

import (
	"github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
)

type Client struct {
	Client *resourcegraph.BaseClient
}

func NewClient(o *common.ClientOptions) *Client {
	client := resourcegraph.NewWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&client.Client, o.ResourceManagerAuthorizer)

	return &Client{
		Client: &client,
	}
}
