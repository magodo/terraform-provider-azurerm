package client

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/mgmt/2020-02-01-preview/maps"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	AccountsClient *maps.AccountsClient
}

func NewClient(o *common.ClientOptions) *Client {
	accountsClient := maps.NewAccountsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&accountsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		AccountsClient: &accountsClient,
	}
}
