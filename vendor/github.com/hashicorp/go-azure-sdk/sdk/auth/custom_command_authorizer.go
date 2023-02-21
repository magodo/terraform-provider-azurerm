package auth

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"golang.org/x/oauth2"
)

type CustomCommandAuthorizerOptions struct {
	// Api describes the Azure API being used
	Api environments.Api

	// TenantId is the tenant to authenticate against
	TenantId string

	// AuxTenantIds lists additional tenants to authenticate against, currently only
	// used for Resource Manager when auxiliary tenants are needed.
	// e.g. https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/authenticate-multi-tenant
	AuxTenantIds []string

	// TokenType specifies the type of the access token. By default, it is "Bearer".
	TokenType string

	// Command is the exec form of command used to retrieve the access token from the stdout. Any surrounding empty space and quotes around the token will be trimmed when being used.
	// Each command argument will be rendered as a Go template with an input object that has following fields:
	// - .ApiEndpoint: The Azure API endpoint if exists, otherwise is an empty string.
	// - .ApiResourceIdentifier: The Azure API resource identifier if exists, otherwise is an empty string.
	// - .ApiAppId: The application ID of the Azure API if exists, otherwise is an empty string.
	// - .ApiName: The name of the Azure API.
	// - .ApiScope: The scope of the Azure API if exists, otherwise an error will return.
	// - .TenantID: The tenant ID. For auxiliary tokens, it is set as one of each auxiliary token.
	// E.g. []string{"az", "account", "get-access-token", "--scope={{.ApiScope}}", "--query=accessToken"}"}
	Command []string
}

// NewCustomCommandAuthorizer returns an Authorizer which authenticates using the custom command.
func NewCustomCommandAuthorizer(ctx context.Context, options CustomCommandAuthorizerOptions) (Authorizer, error) {
	conf, err := newCustomCommandConfig(options.Api, options.TenantId, options.AuxTenantIds, options.TokenType, options.Command)
	if err != nil {
		return nil, err
	}
	return conf.TokenSource(ctx)
}

var _ Authorizer = &CustomCommandAuthorizer{}

// CustomCommandAuthorizer is an Authorizer which supports the custom command.
type CustomCommandAuthorizer struct {
	conf *customCommandConfig
}

// Token returns an access token using the Azure CLI as an authentication mechanism.
func (a *CustomCommandAuthorizer) Token(_ context.Context, _ *http.Request) (*oauth2.Token, error) {
	if a.conf == nil {
		return nil, fmt.Errorf("could not request token: conf is nil")
	}

	token, err := runCustomCommand(a.conf.PrimaryCommand)
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken: token,
		TokenType:   a.conf.TokenType,
	}, nil
}

// AuxiliaryTokens returns additional tokens for auxiliary tenant IDs, for use in multi-tenant scenarios
func (a *CustomCommandAuthorizer) AuxiliaryTokens(_ context.Context, _ *http.Request) ([]*oauth2.Token, error) {
	if a.conf == nil {
		return nil, fmt.Errorf("could not request token: conf is nil")
	}

	tokens := make([]*oauth2.Token, 0)
	for _, command := range a.conf.AuxiliaryCommands {
		token, err := runCustomCommand(command)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, &oauth2.Token{
			AccessToken: token,
			TokenType:   a.conf.TokenType,
		})
	}

	return tokens, nil
}

// customCommandConfig configures an CustomCommandAuthorizer.
type customCommandConfig struct {
	// Api describes the Azure API being used
	Api environments.Api

	// TenantId is the tenant to authenticate against
	TenantId string

	// AuxTenantIds lists additional tenants to authenticate against, currently only
	// used for Resource Manager when auxiliary tenants are needed.
	// e.g. https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/authenticate-multi-tenant
	AuxTenantIds []string

	// TokenType specifies the type of the access token. By default, it is "Bearer".
	TokenType string

	// PrimaryCommand is the rendered command for the primary tenant
	PrimaryCommand []string

	// AuxiliaryCommands are a list of rendered commands for each auxiliary tenant
	AuxiliaryCommands [][]string
}

// newCustomCommandConfig returns a new customCommandConfig.
func newCustomCommandConfig(api environments.Api, tenantId string, auxiliaryTenantIds []string, tokenType string, command []string) (*customCommandConfig, error) {
	config := customCommandConfig{
		Api:          api,
		TenantId:     tenantId,
		AuxTenantIds: auxiliaryTenantIds,
		TokenType:    tokenType,
	}

	buildCommand := func(api environments.Api, tenantID string, rawCommand []string) ([]string, error) {
		if len(rawCommand) == 0 {
			return nil, fmt.Errorf("missing command")
		}
		command := make([]string, len(rawCommand))
		copy(command, rawCommand)

		for i, arg := range rawCommand {
			// Not render the command to run
			if i == 0 {
				continue
			}

			var apiEndpoint string
			if v, ok := api.Endpoint(); ok {
				apiEndpoint = *v
			}

			var apiResourceIdentifier string
			if v, ok := api.ResourceIdentifier(); ok {
				apiResourceIdentifier = *v
			}

			var apiAppId string
			if v, ok := api.AppId(); ok {
				apiAppId = *v
			}

			apiScope, err := environments.Scope(api)
			if err != nil {
				return nil, err
			}

			apiName := api.Name()

			inputObj := struct {
				ApiEndpoint           string
				ApiResourceIdentifier string
				ApiAppId              string
				ApiName               string
				ApiScope              string
				TenantID              string
			}{
				ApiEndpoint:           apiEndpoint,
				ApiResourceIdentifier: apiResourceIdentifier,
				ApiAppId:              apiAppId,
				ApiName:               apiName,
				ApiScope:              *apiScope,
				TenantID:              tenantID,
			}
			tpl, err := template.New("arg").Parse(arg)
			if err != nil {
				return nil, fmt.Errorf("format of the %d-th argument is not valid: %v", i, err)
			}
			var buf bytes.Buffer
			if err := tpl.Execute(&buf, inputObj); err != nil {
				return nil, fmt.Errorf("format of the %d-th argument is not valid: %v", i, err)
			}
			command[i] = buf.String()
		}
		return command, nil
	}

	pcommand, err := buildCommand(config.Api, config.TenantId, command)
	if err != nil {
		return nil, err
	}
	config.PrimaryCommand = pcommand

	for _, tenantId := range config.AuxTenantIds {
		acommand, err := buildCommand(config.Api, tenantId, command)
		if err != nil {
			return nil, err
		}
		config.AuxiliaryCommands = append(config.AuxiliaryCommands, acommand)
	}

	return &config, nil
}

// TokenSource provides a source for obtaining access tokens using CustomCommandAuthorizer.
func (c *customCommandConfig) TokenSource(ctx context.Context) (Authorizer, error) {
	// Cache access tokens internally to avoid unnecessary custom command invocations
	return NewCachedAuthorizer(&CustomCommandAuthorizer{
		conf: c,
	})
}

// runCustomCommand executes the custom command and return the output access token with any quote/emptyspace trimmed.
func runCustomCommand(command []string) (string, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		err := fmt.Errorf("launching custom command: %+v", err)
		if stdErrStr := stderr.String(); stdErrStr != "" {
			err = fmt.Errorf("%s: %s", err, strings.TrimSpace(stdErrStr))
		}
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		err := fmt.Errorf("running custom command: %+v", err)
		if stdErrStr := stderr.String(); stdErrStr != "" {
			err = fmt.Errorf("%s: %s", err, strings.TrimSpace(stdErrStr))
		}
		return "", err
	}

	return strings.Trim(strings.TrimSpace(stdout.String()), `"`), nil
}
