package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func (s *SmD) IsUsingAuthentication() bool {
	return s.jwksURL != ""
}

func (s *SmD) VerifyClaims(testClaims []string, r *http.Request) (bool, error) {
	// extract claims from JWT
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return false, fmt.Errorf("failed to get claims(s) from token: %v", err)
	}

	// verify that each one of the test claims are included
	for _, testClaim := range testClaims {
		_, ok := claims[testClaim]
		if !ok {
			return false, fmt.Errorf("failed to verify claim(s) from token: %s", testClaim)
		}
	}
	return true, nil
}

func (s *SmD) VerifyScope(testScopes []string, r *http.Request) (bool, error) {
	// extract the scopes from JWT
	var scopes []string
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return false, fmt.Errorf("failed to get claim(s) from token: %v", err)
	}

	appendScopes := func(slice []string, scopeClaim any) []string {
		switch scopeClaim.(type) {
		case []any:
			// convert all scopes to str and append
			for _, s := range scopeClaim.([]any) {
				switch s.(type) {
				case string:
					slice = append(slice, s.(string))
				}
			}
		case []string:
			slice = append(slice, scopeClaim.([]string)...)
		}
		return slice
	}
	v, ok := claims["scp"]
	if ok {
		scopes = appendScopes(scopes, v)
	}
	v, ok = claims["scope"]
	if ok {
		scopes = appendScopes(scopes, v)
	}

	// check for both 'scp' and 'scope' claims for scope
	scopeClaim, ok := claims["scp"]
	if ok {
		switch scopeClaim.(type) {
		case []any:
			// convert all scopes to str and append
			for _, s := range scopeClaim.([]any) {
				switch s.(type) {
				case string:
					scopes = append(scopes, s.(string))
				}
			}
		case []string:
			scopes = append(scopes, scopeClaim.([]string)...)
		}
	}
	scopeClaim, ok = claims["scope"]
	if ok {
		scopes = append(scopes, scopeClaim.([]string)...)
	}

	// verify that each of the test scopes are included
	for _, testScope := range testScopes {
		index := slices.Index(scopes, testScope)
		if index < 0 {
			return false, fmt.Errorf("invalid or missing scope")
		}
	}
	// NOTE: should this be ok if no scopes were found?
	return true, nil
}

func (s *SmD) fetchPublicKeyFromURL(url string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	set, err := jwk.Fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	jwks, err := json.Marshal(set)
	if err != nil {
		return fmt.Errorf("failed to marshal JWKS: %v", err)
	}
	s.tokenAuth, err = jwtauth.NewKeySet(jwks)
	if err != nil {
		return fmt.Errorf("failed to initialize JWKS: %v", err)
	}

	return nil
}
