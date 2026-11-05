// Copyright (c) HashiCorp Inc. All rights reserved.
// Licensed under the MPL-2.0 License. See NOTICE.txt in the project root for license information.

package auth

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"os"

	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"golang.org/x/oauth2"
)

type ManagedIdentityOIDCAuthorizerOptions struct {
	// Environment is the Azure environment/cloud being targeted.
	Environment environments.Environment

	// Api describes the Azure API being used.
	Api environments.Api

	// TenantId is the tenant to authenticate against.
	TenantId string

	// AuxiliaryTenantIds lists additional tenants to authenticate against.
	AuxiliaryTenantIds []string

	// AppClientId is the client ID of the application with the federated identity credential.
	AppClientId string

	// ManagedIdentityClientId is the client ID of the user-assigned managed identity trusted by the application.
	ManagedIdentityClientId string

	// CustomManagedIdentityEndpoint optionally overrides the default metadata endpoint.
	CustomManagedIdentityEndpoint string

	// CustomManagedIdentityAPIVersion optionally overrides the metadata API version.
	CustomManagedIdentityAPIVersion string

	// CustomManagedIdentityHeaders are additional headers for metadata requests.
	CustomManagedIdentityHeaders map[string][]string
}

// NewManagedIdentityOIDCAuthorizer returns an authorizer which uses a managed identity token
// as a client assertion to authenticate an application.
func NewManagedIdentityOIDCAuthorizer(ctx context.Context, options ManagedIdentityOIDCAuthorizerOptions) (Authorizer, error) {
	_, err := environments.Scope(options.Api)
	if err != nil {
		return nil, fmt.Errorf("determining scope for %q: %+v", options.Api.Name(), err)
	}

	// For Aztfy-api, it's hosted on Azure Container App.
	// To keep the effort when merge upstream we handle these environment variables here.
	customManagedIdentityEndpoint := options.CustomManagedIdentityEndpoint
	customManagedIdentityAPIVersion := options.CustomManagedIdentityAPIVersion
	customManagedIdentityHeaders := maps.Clone(options.CustomManagedIdentityHeaders)

	// On ACS, the endpoint and header environment are always set.
	if identityEndpoint, identityHeader := os.Getenv("IDENTITY_ENDPOINT"), os.Getenv("IDENTITY_HEADER"); identityEndpoint != "" && identityHeader != "" {
		if customManagedIdentityEndpoint == "" {
			customManagedIdentityEndpoint = identityEndpoint
		}
		if customManagedIdentityAPIVersion == "" {
			customManagedIdentityAPIVersion = "2019-08-01"
		}
		if customManagedIdentityHeaders == nil {
			customManagedIdentityHeaders = make(map[string][]string)
		}
		if _, ok := customManagedIdentityHeaders["X-IDENTITY-HEADER"]; !ok {
			customManagedIdentityHeaders["X-IDENTITY-HEADER"] = []string{identityHeader}
		}
	}

	managedIdentityAuthorizer, err := NewManagedIdentityAuthorizer(ctx, ManagedIdentityAuthorizerOptions{
		Api:                             environments.NewApiEndpoint("AzureADTokenExchange", "api://AzureADTokenExchange", nil),
		ClientId:                        options.ManagedIdentityClientId,
		CustomManagedIdentityEndpoint:   customManagedIdentityEndpoint,
		CustomManagedIdentityAPIVersion: customManagedIdentityAPIVersion,
		CustomManagedIdentityHeaders:    customManagedIdentityHeaders,
	})
	if err != nil {
		return nil, fmt.Errorf("building managed identity authorizer: %w", err)
	}

	return NewCachedAuthorizer(&ManagedIdentityOIDCAuthorizer{
		managedIdentityAuthorizer: managedIdentityAuthorizer,
		oidcOptions: OIDCAuthorizerOptions{
			Environment:        options.Environment,
			Api:                options.Api,
			TenantId:           options.TenantId,
			AuxiliaryTenantIds: append([]string(nil), options.AuxiliaryTenantIds...),
			ClientId:           options.AppClientId,
		},
	})
}

var _ Authorizer = &ManagedIdentityOIDCAuthorizer{}

// ManagedIdentityOIDCAuthorizer authenticates an application using a federated managed identity.
type ManagedIdentityOIDCAuthorizer struct {
	managedIdentityAuthorizer Authorizer
	oidcOptions               OIDCAuthorizerOptions
}

func (a *ManagedIdentityOIDCAuthorizer) tokenSource(ctx context.Context, req *http.Request) (Authorizer, error) {
	if a.managedIdentityAuthorizer == nil {
		return nil, fmt.Errorf("ManagedIdentityOIDCAuthorizer is not configured")
	}

	assertion, err := a.managedIdentityAuthorizer.Token(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ManagedIdentityOIDCAuthorizer: acquiring managed identity assertion: %w", err)
	}
	if assertion == nil || assertion.AccessToken == "" {
		return nil, fmt.Errorf("ManagedIdentityOIDCAuthorizer: received an empty managed identity assertion")
	}

	options := a.oidcOptions
	options.FederatedAssertion = assertion.AccessToken
	return NewOIDCAuthorizer(ctx, options)
}

// Token returns an access token for the application and configured API.
func (a *ManagedIdentityOIDCAuthorizer) Token(ctx context.Context, req *http.Request) (*oauth2.Token, error) {
	source, err := a.tokenSource(ctx, req)
	if err != nil {
		return nil, err
	}
	return source.Token(ctx, req)
}

// AuxiliaryTokens returns application tokens for the configured auxiliary tenants.
func (a *ManagedIdentityOIDCAuthorizer) AuxiliaryTokens(ctx context.Context, req *http.Request) ([]*oauth2.Token, error) {
	if len(a.oidcOptions.AuxiliaryTenantIds) == 0 {
		return []*oauth2.Token{}, nil
	}
	source, err := a.tokenSource(ctx, req)
	if err != nil {
		return nil, err
	}
	return source.AuxiliaryTokens(ctx, req)
}
