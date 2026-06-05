package ltidr

import (
	"fmt"
	"net/url"
	"strings"
)

func validatePlatformConfiguration(openIDConfigurationURL string, config *platformOpenIDConfiguration) error {
	issuerURL, err := validateIssuerURL(config.Issuer)
	if err != nil {
		return err
	}

	if err := validateOpenIDConfigurationURL(openIDConfigurationURL, issuerURL); err != nil {
		return err
	}

	if err := validateHTTPSURL(config.AuthorizationEndpoint, false, "authorization_endpoint"); err != nil {
		return err
	}
	if err := validateHTTPSURL(config.RegistrationEndpoint, true, "registration_endpoint"); err != nil {
		return err
	}
	if err := validateHTTPSURL(config.JWKSURI, true, "jwks_uri"); err != nil {
		return err
	}
	if err := validateHTTPSURL(config.TokenEndpoint, true, "token_endpoint"); err != nil {
		return err
	}
	if config.AuthorizationServer != nil && strings.TrimSpace(*config.AuthorizationServer) != "" {
		if err := validateHTTPSURL(*config.AuthorizationServer, false, "authorization_server"); err != nil {
			return err
		}
	}

	return nil
}

func validateIssuerURL(rawIssuer string) (*url.URL, error) {
	issuerURL, err := url.Parse(rawIssuer)
	if err != nil {
		return nil, fmt.Errorf("issuer is not a valid URL")
	}
	if !strings.EqualFold(issuerURL.Scheme, "https") {
		return nil, fmt.Errorf("issuer must use https")
	}
	if issuerURL.Host == "" {
		return nil, fmt.Errorf("issuer must include a host")
	}
	if issuerURL.RawQuery != "" || issuerURL.Fragment != "" {
		return nil, fmt.Errorf("issuer must not include query or fragment")
	}
	return issuerURL, nil
}

func validateOpenIDConfigurationURL(rawConfigURL string, issuerURL *url.URL) error {
	configURL, err := url.Parse(rawConfigURL)
	if err != nil {
		return fmt.Errorf("openid_configuration is not a valid URL")
	}
	if !strings.EqualFold(configURL.Scheme, "https") {
		return fmt.Errorf("openid_configuration must use https")
	}
	if configURL.Host != issuerURL.Host {
		return fmt.Errorf("openid_configuration host must match issuer host exactly")
	}
	if configURL.Fragment != "" {
		return fmt.Errorf("openid_configuration must not include a fragment")
	}

	issuerPath := strings.TrimSuffix(issuerURL.EscapedPath(), "/")
	configPath := configURL.EscapedPath()
	if issuerPath != "" {
		if !strings.HasPrefix(configPath, issuerPath+"/") {
			return fmt.Errorf("openid_configuration path must extend the issuer path")
		}
	} else if !strings.HasPrefix(configPath, "/") {
		return fmt.Errorf("openid_configuration must include an absolute path")
	}

	return nil
}

func validateHTTPSURL(raw string, allowQuery bool, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s is not a valid URL", field)
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return fmt.Errorf("%s must use https", field)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%s must include a host", field)
	}
	if !allowQuery && parsed.RawQuery != "" {
		return fmt.Errorf("%s must not include a query", field)
	}
	if parsed.Fragment != "" {
		return fmt.Errorf("%s must not include a fragment", field)
	}
	return nil
}
