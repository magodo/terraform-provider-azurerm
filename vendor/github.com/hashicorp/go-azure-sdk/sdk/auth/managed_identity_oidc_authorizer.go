// Copyright (c) HashiCorp Inc. All rights reserved.
// Licensed under the MPL-2.0 License. See NOTICE.txt in the project root for license information.

package auth

import (
	"context"
	"fmt"
	"net/http"

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

	managedIdentityAuthorizer, err := NewManagedIdentityAuthorizer(ctx, ManagedIdentityAuthorizerOptions{
		Api:                             environments.NewApiEndpoint("AzureADTokenExchange", "api://AzureADTokenExchange", nil),
		ClientId:                        options.ManagedIdentityClientId,
		CustomManagedIdentityEndpoint:   options.CustomManagedIdentityEndpoint,
		CustomManagedIdentityAPIVersion: options.CustomManagedIdentityAPIVersion,
		CustomManagedIdentityHeaders:    options.CustomManagedIdentityHeaders,
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
