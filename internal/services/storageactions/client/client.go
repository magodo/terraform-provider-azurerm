package client

import (
	"fmt"

	storageactions "github.com/hashicorp/go-azure-sdk/resource-manager/storageactions/2023-01-01"
	"github.com/hashicorp/go-azure-sdk/sdk/client/resourcemanager"
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
)

type Client struct {
	*storageactions.Client
}

func NewClient(o *common.ClientOptions) (*Client, error) {
	client, err := storageactions.NewClientWithBaseURI(o.Environment.ResourceManager, func(c *resourcemanager.Client) {
		o.Configure(c, o.Authorizers.ResourceManager)
	})
	if err != nil {
		return nil, fmt.Errorf("building Storage Actions Client: %+v", err)
	}
	return &Client{Client: client}, nil
}
