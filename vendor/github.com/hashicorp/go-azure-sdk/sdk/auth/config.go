package auth

import (
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
)

// Copyright (c) HashiCorp Inc. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

// Credentials sets up NewAuthorizer to return an Authorizer based on the provided credentails.
type Credentials struct {
	// Specifies the national cloud environment to use
	Environment environments.Environment

	// AuxiliaryTenantIDs specifies the Auxiliary Tenant IDs for which to obtain tokens in a multi-tenant scenario.
	AuxiliaryTenantIDs []string
	// ClientID specifies the Client ID for the application used to authenticate the connection
	ClientID string
	// TenantID specifies the Azure Active Directory Tenant to connect to, which must be a valid UUID.
	TenantID string

	// EnableAuthenticatingUsingAzureCLI specifies whether Azure CLI authentication should be checked.
	EnableAuthenticatingUsingAzureCLI bool

	// EnableAuthenticatingUsingClientCertificate specifies whether Client Certificate authentication should be checked.
	EnableAuthenticatingUsingClientCertificate bool
	// ClientCertificateData specifies the contents of a Client Certificate PKCS#12 bundle.
	ClientCertificateData []byte
	// ClientCertificatePath specifies the path to a Client Certificate PKCS#12 bundle (.pfx file)
	ClientCertificatePath string
	// ClientCertificatePassword specifies the encryption password to unlock a Client Certificate.
	ClientCertificatePassword string

	// EnableAuthenticatingUsingClientSecret specifies whether Client Secret authentication should be used.
	EnableAuthenticatingUsingClientSecret bool
	// ClientSecret specifies the Secret used authenticate using Client Secret authentication.
	ClientSecret string

	// EnableAuthenticatingUsingManagedIdentity specifies whether Managed Identity authentication should be checked.
	EnableAuthenticatingUsingManagedIdentity bool
	// CustomManagedIdentityEndpoint specifies a custom endpoint which should be used for Managed Identity.
	CustomManagedIdentityEndpoint string

	// Enables OIDC authentication (federated client credentials).
	EnableAuthenticationUsingOIDC bool
	// OIDCAssertionToken specifies the OIDC Assertion Token to authenticate using Client Credentials.
	OIDCAssertionToken string

	// EnableAuthenticationUsingGitHubOIDC specifies whether GitHub OIDC
	EnableAuthenticationUsingGitHubOIDC bool
	// GitHubOIDCTokenRequestURL specifies the URL for GitHub's OIDC provider
	GitHubOIDCTokenRequestURL string
	// GitHubOIDCTokenRequestToken specifies the bearer token for the request to GitHub's OIDC provider
	GitHubOIDCTokenRequestToken string

	// EnableCustomCommand specifies whether custom command should be checked.
	EnableCustomCommand bool
	// CustomCommand is the exec form of command used to retrieve the access token from the stdout.
	// Each command argument will be rendered as a Go template with an input object that has following fields:
	// - .ApiEndpoint: The Azure API endpoint if exists, otherwise is an empty string.
	// - .ApiResourceIdentifier: The Azure API resource identifier if exists, otherwise is an empty string.
	// - .ApiAppId: The application ID of the Azure API if exists, otherwise is an empty string.
	// - .ApiName: The name of the Azure API.
	// - .ApiScope: The scope of the Azure API if exists, otherwise an error will return.
	// - .TenantID: The tenant ID. For auxiliary tokens, it is set as one of each auxiliary token.
	// E.g. []string{"az", "account", "get-access-token", "--scope={{.ApiScope}}", "--query=accessToken"}"}
	CustomCommand []string
	// CustomCommandTokenType specifies the type of the access token retrieved by the custom command. By default, it is "Bearer".
	CustomCommandTokenType string
}
