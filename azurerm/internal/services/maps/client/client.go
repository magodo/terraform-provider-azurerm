package client

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/mgmt/2020-02-01-preview/maps"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	AccountsClient *maps.AccountsClient
	CreatorsClient *maps.CreatorsClient
}

func NewClient(o *common.ClientOptions) *Client {
	accountsClient := maps.NewAccountsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&accountsClient.Client, o.ResourceManagerAuthorizer)

	creatorsClient := maps.NewCreatorsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&creatorsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		AccountsClient: &accountsClient,
		CreatorsClient: &creatorsClient,
	}
}
