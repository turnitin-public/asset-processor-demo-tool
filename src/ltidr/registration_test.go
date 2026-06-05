package ltidr

import "testing"

func TestValidatePlatformConfiguration(t *testing.T) {
	authorizationServer := "https://platform.example.com/authz"
	tests := []struct {
		name      string
		configURL string
		config    platformOpenIDConfiguration
		wantErr   bool
	}{
		{
			name:      "accepts matching issuer path",
			configURL: "https://platform.example.com/lti/.well-known/openid-configuration?registration=123",
			config: platformOpenIDConfiguration{
				Issuer:                "https://platform.example.com/lti",
				AuthorizationEndpoint: "https://platform.example.com/lti/auth",
				RegistrationEndpoint:  "https://platform.example.com/lti/register?token=abc",
				JWKSURI:               "https://platform.example.com/lti/jwks",
				TokenEndpoint:         "https://platform.example.com/lti/token",
				AuthorizationServer:   &authorizationServer,
			},
		},
		{
			name:      "rejects mismatched issuer path",
			configURL: "https://platform.example.com/.well-known/openid-configuration",
			config: platformOpenIDConfiguration{
				Issuer:                "https://platform.example.com/lti",
				AuthorizationEndpoint: "https://platform.example.com/lti/auth",
				RegistrationEndpoint:  "https://platform.example.com/lti/register",
				JWKSURI:               "https://platform.example.com/lti/jwks",
				TokenEndpoint:         "https://platform.example.com/lti/token",
			},
			wantErr: true,
		},
		{
			name:      "rejects issuer query string",
			configURL: "https://platform.example.com/lti/.well-known/openid-configuration",
			config: platformOpenIDConfiguration{
				Issuer:                "https://platform.example.com/lti?foo=bar",
				AuthorizationEndpoint: "https://platform.example.com/lti/auth",
				RegistrationEndpoint:  "https://platform.example.com/lti/register",
				JWKSURI:               "https://platform.example.com/lti/jwks",
				TokenEndpoint:         "https://platform.example.com/lti/token",
			},
			wantErr: true,
		},
		{
			name:      "rejects non-https token endpoint",
			configURL: "https://platform.example.com/lti/.well-known/openid-configuration",
			config: platformOpenIDConfiguration{
				Issuer:                "https://platform.example.com/lti",
				AuthorizationEndpoint: "https://platform.example.com/lti/auth",
				RegistrationEndpoint:  "https://platform.example.com/lti/register",
				JWKSURI:               "https://platform.example.com/lti/jwks",
				TokenEndpoint:         "http://platform.example.com/lti/token",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePlatformConfiguration(test.configURL, &test.config)
			if test.wantErr && err == nil {
				t.Fatalf("expected error but got nil")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("expected success but got error: %v", err)
			}
		})
	}
}

func TestBuildToolRegistrationPayload(t *testing.T) {
	payload := buildToolRegistrationPayload(
		"Brightspace Demo Tool Deployment",
		"https://tool.example.com",
		"https://tool.example.com/lti/launch",
		"https://tool.example.com/oidc/login?iss=https%3A%2F%2Fplatform.example.com&reg_id=abc",
	)

	if payload.ApplicationType != "web" {
		t.Fatalf("unexpected application type: %s", payload.ApplicationType)
	}
	if len(payload.GrantTypes) != 2 || payload.GrantTypes[0] != "client_credentials" || payload.GrantTypes[1] != "implicit" {
		t.Fatalf("unexpected grant types: %#v", payload.GrantTypes)
	}
	if payload.InitiateLoginURI == "" || len(payload.RedirectURIs) != 1 || payload.RedirectURIs[0] != "https://tool.example.com/lti/launch" {
		t.Fatalf("unexpected launch URLs in payload: %#v", payload)
	}
	if payload.JwksURI != "https://tool.example.com/.well-known/jwks.json" {
		t.Fatalf("unexpected jwks_uri: %s", payload.JwksURI)
	}
	if payload.TokenEndpointAuthMethod != "private_key_jwt" {
		t.Fatalf("unexpected token auth method: %s", payload.TokenEndpointAuthMethod)
	}
	if payload.Scope != "openid" {
		t.Fatalf("unexpected scope: %s", payload.Scope)
	}
	if payload.LtiToolConfiguration.Domain != "tool.example.com" {
		t.Fatalf("unexpected domain: %s", payload.LtiToolConfiguration.Domain)
	}
	if payload.LtiToolConfiguration.TargetLinkURI != "https://tool.example.com/lti/launch" {
		t.Fatalf("unexpected target_link_uri: %s", payload.LtiToolConfiguration.TargetLinkURI)
	}
	if len(payload.LtiToolConfiguration.Messages) != 1 || payload.LtiToolConfiguration.Messages[0].Type != "LtiResourceLinkRequest" {
		t.Fatalf("unexpected messages: %#v", payload.LtiToolConfiguration.Messages)
	}
}
