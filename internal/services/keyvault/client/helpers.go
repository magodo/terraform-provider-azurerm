package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2021-10-01/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2021-03-01/resourcegraph"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault/parse"
	resourceGraphClient "github.com/hashicorp/terraform-provider-azurerm/internal/services/resourcegraph/client"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var (
	keyVaultsCache = map[string]keyVaultDetails{}
	keysmith       = &sync.RWMutex{}
	lock           = map[string]*sync.RWMutex{}
)

type keyVaultDetails struct {
	keyVaultId       string
	dataPlaneBaseUri string
	resourceGroup    string
}

func (c *Client) AddToCache(keyVaultId parse.VaultId, dataPlaneUri string) {
	cacheKey := c.cacheKeyForKeyVault(keyVaultId.Name)
	keysmith.Lock()
	keyVaultsCache[cacheKey] = keyVaultDetails{
		keyVaultId:       keyVaultId.ID(),
		dataPlaneBaseUri: dataPlaneUri,
		resourceGroup:    keyVaultId.ResourceGroup,
	}
	keysmith.Unlock()
}

func (c *Client) BaseUriForKeyVault(ctx context.Context, keyVaultId parse.VaultId) (*string, error) {
	cacheKey := c.cacheKeyForKeyVault(keyVaultId.Name)
	keysmith.Lock()
	if lock[cacheKey] == nil {
		lock[cacheKey] = &sync.RWMutex{}
	}
	keysmith.Unlock()
	lock[cacheKey].Lock()
	defer lock[cacheKey].Unlock()

	if keyVaultId.SubscriptionId != c.VaultsClient.SubscriptionID {
		c.VaultsClient = c.KeyVaultClientForSubscription(keyVaultId.SubscriptionId)
	}

	resp, err := c.VaultsClient.Get(ctx, keyVaultId.ResourceGroup, keyVaultId.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return nil, fmt.Errorf("%s was not found", keyVaultId)
		}
		return nil, fmt.Errorf("retrieving %s: %+v", keyVaultId, err)
	}

	if resp.Properties == nil || resp.Properties.VaultURI == nil {
		return nil, fmt.Errorf("`properties` was nil for %s", keyVaultId)
	}

	return resp.Properties.VaultURI, nil
}

func (c *Client) Exists(ctx context.Context, keyVaultId parse.VaultId) (bool, error) {
	cacheKey := c.cacheKeyForKeyVault(keyVaultId.Name)
	keysmith.Lock()
	if lock[cacheKey] == nil {
		lock[cacheKey] = &sync.RWMutex{}
	}
	keysmith.Unlock()
	lock[cacheKey].Lock()
	defer lock[cacheKey].Unlock()

	if _, ok := keyVaultsCache[cacheKey]; ok {
		return true, nil
	}

	resp, err := c.VaultsClient.Get(ctx, keyVaultId.ResourceGroup, keyVaultId.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return false, nil
		}
		return false, fmt.Errorf("retrieving %s: %+v", keyVaultId, err)
	}

	if resp.Properties == nil || resp.Properties.VaultURI == nil {
		return false, fmt.Errorf("`properties` was nil for %s", keyVaultId)
	}

	c.AddToCache(keyVaultId, *resp.Properties.VaultURI)

	return true, nil
}

func (c *Client) KeyVaultIDFromBaseUrl(ctx context.Context, argClient *resourceGraphClient.Client, keyVaultBaseUrl string) (*string, error) {
	keyVaultName, err := c.parseNameFromBaseUrl(keyVaultBaseUrl)
	if err != nil {
		return nil, err
	}

	cacheKey := c.cacheKeyForKeyVault(*keyVaultName)
	keysmith.Lock()
	if lock[cacheKey] == nil {
		lock[cacheKey] = &sync.RWMutex{}
	}
	keysmith.Unlock()
	lock[cacheKey].Lock()
	defer lock[cacheKey].Unlock()

	if v, ok := keyVaultsCache[cacheKey]; ok {
		return &v.keyVaultId, nil
	}

	query := fmt.Sprintf("resources | where type =~ 'Microsoft.KeyVault/vaults' and name =~ '%s'", *keyVaultName)
	request := resourcegraph.QueryRequest{
		Subscriptions: &[]string{c.VaultsClient.SubscriptionID},
		Query:         &query,
		Options: &resourcegraph.QueryRequestOptions{
			ResultFormat: resourcegraph.ResultFormatObjectArray,
		},
	}

	results, err := argClient.Client.Resources(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("listing resources matching %q: %+v", query, err)
	}

	data, err := json.Marshal(results.Data)
	if err != nil {
		return nil, fmt.Errorf("marshaling the ARG result data for query %q: %+v", query, err)
	}

	var kvs []keyvault.Vault
	if err := json.Unmarshal(data, &kvs); err != nil {
		return nil, fmt.Errorf("unmarshalling the ARG result data for query %q as key vaults: %+v", query, err)
	}

	for _, kv := range kvs {
		if kv.ID == nil {
			continue
		}
		if kv.Properties == nil || kv.Properties.VaultURI == nil {
			continue
		}
		id, err := parse.VaultID(*kv.ID)
		if err != nil {
			return nil, fmt.Errorf("parsing %q: %+v", *kv.ID, err)
		}
		c.AddToCache(*id, *kv.Properties.VaultURI)
		return utils.String(id.ID()), nil
	}

	// we haven't found it, but Data Sources and Resources need to handle this error separately
	return nil, nil
}

func (c *Client) Purge(keyVaultId parse.VaultId) {
	cacheKey := c.cacheKeyForKeyVault(keyVaultId.Name)
	keysmith.Lock()
	if lock[cacheKey] == nil {
		lock[cacheKey] = &sync.RWMutex{}
	}
	keysmith.Unlock()
	lock[cacheKey].Lock()
	delete(keyVaultsCache, cacheKey)
	lock[cacheKey].Unlock()
}

func (c *Client) cacheKeyForKeyVault(name string) string {
	return strings.ToLower(name)
}

func (c *Client) parseNameFromBaseUrl(input string) (*string, error) {
	uri, err := url.Parse(input)
	if err != nil {
		return nil, err
	}

	// https://the-keyvault.vault.azure.net
	// https://the-keyvault.vault.microsoftazure.de
	// https://the-keyvault.vault.usgovcloudapi.net
	// https://the-keyvault.vault.cloudapi.microsoft
	// https://the-keyvault.vault.azure.cn

	segments := strings.Split(uri.Host, ".")
	if len(segments) < 3 || segments[1] != "vault" {
		return nil, fmt.Errorf("expected a URI in the format `the-keyvault-name.vault.**` but got %q", uri.Host)
	}
	return &segments[0], nil
}
