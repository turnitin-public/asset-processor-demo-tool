package ltidr

type toolRegistrationPayload struct {
	ApplicationType         string                   `json:"application_type"`
	GrantTypes              []string                 `json:"grant_types"`
	ResponseTypes           []string                 `json:"response_types"`
	InitiateLoginURI        string                   `json:"initiate_login_uri"`
	RedirectURIs            []string                 `json:"redirect_uris"`
	ClientName              string                   `json:"client_name"`
	JwksURI                 string                   `json:"jwks_uri"`
	TokenEndpointAuthMethod string                   `json:"token_endpoint_auth_method"`
	Scope                   string                   `json:"scope"`
	LtiToolConfiguration    toolConfigurationPayload `json:"https://purl.imsglobal.org/spec/lti-tool-configuration"`
}

type toolConfigurationPayload struct {
	Domain        string                  `json:"domain"`
	TargetLinkURI string                  `json:"target_link_uri"`
	Messages      []toolMessageDefinition `json:"messages"`
}

type toolMessageDefinition struct {
	Type string `json:"type"`
}

func buildToolRegistrationPayload(customerID string, toolBaseURL string, toolRedirectURI string, initiateLoginURI string) toolRegistrationPayload {
	return toolRegistrationPayload{
		ApplicationType:         "web",
		GrantTypes:              []string{"client_credentials", "implicit"},
		ResponseTypes:           []string{"id_token"},
		InitiateLoginURI:        initiateLoginURI,
		RedirectURIs:            []string{toolRedirectURI},
		ClientName:              resolveToolName(customerID),
		JwksURI:                 toolBaseURL + "/.well-known/jwks.json",
		TokenEndpointAuthMethod: "private_key_jwt",
		Scope:                   "openid",
		LtiToolConfiguration: toolConfigurationPayload{
			Domain:        hostWithoutPort(toolBaseURL),
			TargetLinkURI: toolRedirectURI,
			Messages: []toolMessageDefinition{
				{Type: "LtiResourceLinkRequest"},
			},
		},
	}
}
