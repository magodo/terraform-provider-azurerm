package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hashicorp/go-azure-sdk/sdk/environments"
	"golang.org/x/oauth2"
)

// Copyright (c) HashiCorp Inc. All rights reserved.
// Licensed under the MIT License. See NOTICE.txt in the project root for license information.

type AccessTokenAuthorizerOptions struct {
	// Api describes the Azure API being used
	Api environments.Api

	// TokenMap is a map of access tokens (JWT) issued by the Microsoft Identity Platform, where the key is the api's name, and the value is the token.
	TokenMap map[string][]byte

	// AllowInvalidAuthorizer specifies that NewAccessTokenAuthorizer allows to create an problematic Authorizer, whose Token() method always fails.
	// This is useful to delay the error on actual calls on the Authorizer.
	AllowInvalidAuthorizer bool
}

// NewAccessTokenAuthorizer returns an Authorizer which authenticates using the Access Token.
func NewAccessTokenAuthorizer(ctx context.Context, options AccessTokenAuthorizerOptions) (auth Authorizer, err error) {
	defer func() {
		if err != nil {
			err = newTokenRefreshError(err)
			if options.AllowInvalidAuthorizer {
				auth = &invalidAccessTokenAuthorizer{err: err}
				err = nil
			}
		}
	}()

	token, ok := options.TokenMap[options.Api.Name()]
	if !ok {
		err = fmt.Errorf("no token configured for API name %s", options.Api.Name())
		return
	}
	tk, _, err := jwt.NewParser().ParseUnverified(string(token), jwt.MapClaims{})
	if err != nil {
		err = fmt.Errorf("parsing JWT token: %v", err)
		return
	}
	claims := tk.Claims.(jwt.MapClaims)
	exp, ok := claims["exp"]
	if !ok {
		err = fmt.Errorf(`no "exp" found in claim`)
		return
	}
	expireOn := time.Unix(int64(exp.(float64)), 0)
	if time.Now().After(expireOn) {
		err = fmt.Errorf("token has already expired")
		return
	}

	return &AccessTokenAuthorizer{
		token: oauth2.Token{
			AccessToken: string(token),
			Expiry:      expireOn,
		},
	}, nil
}

var _ Authorizer = &AccessTokenAuthorizer{}

// AccessTokenAuthorizer is an Authorizer which supports the Access Token.
type AccessTokenAuthorizer struct {
	token oauth2.Token
}

// Token returns an access token using the Access token as an authentication mechanism.
func (a *AccessTokenAuthorizer) Token(_ context.Context, _ *http.Request) (*oauth2.Token, error) {
	if time.Now().After(a.token.Expiry) {
		return nil, newTokenRefreshError(fmt.Errorf("token has already expired"))
	}
	return &a.token, nil
}

// AuxiliaryTokens returns additional tokens for auxiliary tenant IDs, for use in multi-tenant scenarios
func (a *AccessTokenAuthorizer) AuxiliaryTokens(_ context.Context, _ *http.Request) ([]*oauth2.Token, error) {
	// We are not returning error, but nil auxiliary tokens here since this method is always called.
	// See: https://github.com/manicminer/hamilton-autorest/blob/2e25d83affaf261180cad1156d2a8f30fe5e8a0a/auth/auth.go#L33
	return nil, nil
}

// invalidAccessTokenAuthorizer is an invalid Authorizer whose methods always return the pre-captured error
type invalidAccessTokenAuthorizer struct {
	err error
}

func (a *invalidAccessTokenAuthorizer) Token(_ context.Context, _ *http.Request) (*oauth2.Token, error) {
	return nil, a.err
}
func (a *invalidAccessTokenAuthorizer) AuxiliaryTokens(_ context.Context, _ *http.Request) ([]*oauth2.Token, error) {
	return nil, a.err
}

// tokenRefreshError is an internal type that implements adal.TokenRefreshError.
// This is the type of the returned error of the (Invalid)AccessTokenAuthorizer, so that these kinds of errors won't be retried (as it will always fail).
// See: https://github.com/Azure/go-autorest/blob/9038e4a609b1899f0eb382d03c3e823b70537125/autorest/sender.go#L331
type tokenRefreshError struct {
	err error
}

var _ adal.TokenRefreshError = tokenRefreshError{}

// Error implements the error interface which is part of the TokenRefreshError interface.
func (tre tokenRefreshError) Error() string {
	return tre.err.Error()
}

// Response implements the TokenRefreshError interface, it returns the raw HTTP response from the refresh operation.
func (tre tokenRefreshError) Response() *http.Response {
	return nil
}

func newTokenRefreshError(err error) error {
	return tokenRefreshError{err: err}
}
