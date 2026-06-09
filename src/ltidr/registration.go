package ltidr

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
)

const ltiToolConfigurationClaim = "https://purl.imsglobal.org/spec/lti-tool-configuration"

type registrationRequest struct {
	OpenIDConfiguration string `json:"openid_configuration"`
	RegistrationToken   string `json:"registration_token"`
	CustomerID          string `json:"customer_id"`
}

type platformOpenIDConfiguration struct {
	Issuer                string  `json:"issuer"`
	AuthorizationEndpoint string  `json:"authorization_endpoint"`
	RegistrationEndpoint  string  `json:"registration_endpoint"`
	JWKSURI               string  `json:"jwks_uri"`
	TokenEndpoint         string  `json:"token_endpoint"`
	ScopesSupported       []string                `json:"scopes_supported"`
	ClaimsSupported       []string                `json:"claims_supported"`
	MessagesSupported     []toolMessageDefinition `json:"messages_supported"`
	AuthorizationServer   *string `json:"authorization_server"`
}

type platformRegistrationResponse struct {
	ClientID string `json:"client_id"`
}

func Initiate(w http.ResponseWriter, r *http.Request) {
	openidConfiguration := html.EscapeString(r.URL.Query().Get("openid_configuration"))
	registrationToken := html.EscapeString(r.URL.Query().Get("registration_token"))
	customerID := html.EscapeString(r.URL.Query().Get("customer_id"))

	if customerID == "" {
		customerID = "Dynamic Registration"
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>LTI Dynamic Registration</title>
  </head>
  <body>
    <h2>LTI Dynamic Registration</h2>
    <form method="post" action="/lti-dr/register">
      <label>OpenID Configuration URL</label><br/>
	<input type="url" name="openid_configuration" value="%s" required style="width:100%%" /><br/><br/>
      <label>Registration Token (optional)</label><br/>
      <input type="text" name="registration_token" value="%s" style="width:600px" /><br/><br/>
      <label>Customer ID</label><br/>
      <input type="text" name="customer_id" value="%s" required style="width:600px" /><br/><br/>
      <button type="submit">Register</button>
    </form>
  </body>
</html>`, openidConfiguration, registrationToken, customerID)
}

func Register(w http.ResponseWriter, r *http.Request) {
	log.Printf("dynamic registration: start method=%s path=%s", r.Method, r.URL.Path)

	errs := utils.JsonErrors{Errors: []utils.JsonError{}, Code: 400}
	log.Print("dynamic registration: parsing request payload")
	payload, ok := parseRegistrationRequest(r, &errs)
	if !ok {
		log.Print("dynamic registration: failed to parse request payload")
		utils.WriteJsonError(w, r, errs)
		return
	}
	log.Printf("dynamic registration: request payload=%+v", payload)
	log.Printf("dynamic registration: request parsed openid_configuration=%s customer_id=%s has_registration_token=%t", payload.OpenIDConfiguration, payload.CustomerID, payload.RegistrationToken != "")

	log.Print("dynamic registration: fetching platform openid configuration")
	platformConfig, ok := fetchPlatformOpenIDConfiguration(payload.OpenIDConfiguration, &errs)
	if !ok {
		log.Print("dynamic registration: failed to fetch platform openid configuration")
		utils.WriteJsonError(w, r, errs)
		return
	}
	log.Printf("dynamic registration: platform config=%+v", platformConfig)
	log.Printf("dynamic registration: fetched platform configuration issuer=%s registration_endpoint=%s", platformConfig.Issuer, platformConfig.RegistrationEndpoint)

	log.Print("dynamic registration: validating platform openid configuration")
	if err := validatePlatformConfiguration(payload.OpenIDConfiguration, platformConfig); err != nil {
		log.Printf("dynamic registration: platform configuration validation failed: %v", err)
		utils.AddError(&errs, err.Error(), err)
		errs.Code = 400
		utils.WriteJsonError(w, r, errs)
		return
	}
	log.Print("dynamic registration: platform openid configuration validation succeeded")

	log.Print("dynamic registration: generating registration and deployment identifiers")
	registrationID := uuid.New().String()
	deploymentID := uuid.New().String()
	toolBaseURL := resolveToolBaseURL(r)
	toolRedirectURI := toolBaseURL + "/lti/launch"
	initiateLoginURI := buildLoginURL(toolBaseURL, registrationID)
	log.Printf("dynamic registration: generated identifiers registration_id=%s deployment_id=%s", registrationID, deploymentID)
	log.Printf("dynamic registration: resolved tool urls base=%s launch=%s login=%s", toolBaseURL, toolRedirectURI, initiateLoginURI)

	log.Print("dynamic registration: sending registration request to platform")
	clientID, ok := registerOnPlatform(platformConfig, payload.RegistrationToken, payload.CustomerID, toolBaseURL, toolRedirectURI, initiateLoginURI, &errs)
	if !ok {
		log.Print("dynamic registration: platform registration request failed")
		utils.WriteJsonError(w, r, errs)
		return
	}
	log.Printf("dynamic registration: platform registration succeeded client_id=%s", clientID)

	log.Print("dynamic registration: loading default key set id")
	keySetID, err := datastore.RegistrationQueries.GetDefaultKeySetID()
	if err != nil {
		log.Printf("dynamic registration: unable to load default key set id: %v", err)
		utils.AddError(&errs, "unable to find key set for new registration", err)
		errs.Code = 500
		utils.WriteJsonError(w, r, errs)
		return
	}
	log.Printf("dynamic registration: loaded key set id=%s", keySetID)

	log.Print("dynamic registration: resolving customer id for persistence")
	customerID := payload.CustomerID
	if strings.TrimSpace(customerID) == "" {
		log.Printf("dynamic registration: customer id not provided, falling back to issuer=%s", platformConfig.Issuer)
		customerID = platformConfig.Issuer
	}
	log.Printf("dynamic registration: using customer id=%s", customerID)

	log.Print("dynamic registration: storing registration and deployment")
	err = datastore.RegistrationQueries.CreateRegistrationAndDeployment(datastore.DynamicRegistrationInsert{
		RegistrationID:              registrationID,
		DeploymentID:                deploymentID,
		Issuer:                      platformConfig.Issuer,
		ClientID:                    clientID,
		PlatformLoginAuthEndpoint:   platformConfig.AuthorizationEndpoint,
		PlatformServiceAuthEndpoint: platformConfig.TokenEndpoint,
		PlatformJwksEndpoint:        platformConfig.JWKSURI,
		PlatformAuthProvider:        platformConfig.AuthorizationServer,
		ToolRedirectURI:             toolRedirectURI,
		KeySetID:                    keySetID,
		CustomerID:                  customerID,
	})
	if err != nil {
		log.Printf("dynamic registration: unable to store registration: %v", err)
		utils.AddError(&errs, "unable to store registration", err)
		errs.Code = 500
		utils.WriteJsonError(w, r, errs)
		return
	}
	log.Print("dynamic registration: registration and deployment stored successfully")

	result := map[string]string{
		"registration_id": registrationID,
		"deployment_id":   deploymentID,
		"issuer":          platformConfig.Issuer,
		"client_id":       clientID,
		"login_url":       initiateLoginURI,
		"launch_url":      toolRedirectURI,
	}
	log.Print("dynamic registration: building success response payload")

	if isJSONRequest(r) {
		log.Print("dynamic registration: responding with JSON")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(result)
		log.Printf("dynamic registration: complete registration_id=%s deployment_id=%s", registrationID, deploymentID)
		return
	}

	log.Print("dynamic registration: responding with HTML completion page")
	renderRegistrationComplete(w, result)
	log.Printf("dynamic registration: complete registration_id=%s deployment_id=%s", registrationID, deploymentID)
}

func parseRegistrationRequest(r *http.Request, errs *utils.JsonErrors) (registrationRequest, bool) {
	request := registrationRequest{}
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			utils.AddError(errs, "invalid JSON body", err)
			errs.Code = 400
			return registrationRequest{}, false
		}
	} else {
		if err := r.ParseForm(); err != nil {
			utils.AddError(errs, "unable to parse form body", err)
			errs.Code = 400
			return registrationRequest{}, false
		}
		request.OpenIDConfiguration = r.FormValue("openid_configuration")
		request.RegistrationToken = r.FormValue("registration_token")
		request.CustomerID = r.FormValue("customer_id")
	}

	request.OpenIDConfiguration = strings.TrimSpace(request.OpenIDConfiguration)
	request.RegistrationToken = strings.TrimSpace(request.RegistrationToken)
	request.CustomerID = strings.TrimSpace(request.CustomerID)

	if request.OpenIDConfiguration == "" {
		utils.AddError(errs, "openid_configuration is required", "missing openid_configuration")
		errs.Code = 400
		return registrationRequest{}, false
	}

	return request, true
}

func fetchPlatformOpenIDConfiguration(openIDConfigurationURL string, errs *utils.JsonErrors) (*platformOpenIDConfiguration, bool) {
	req, err := http.NewRequest("GET", openIDConfigurationURL, nil)
	if err != nil {
		utils.AddError(errs, "unable to create openid configuration request", err)
		errs.Code = 400
		return nil, false
	}

	resp, err := utils.HttpClient().Do(req)
	log.Printf("%+v", resp)
	if err != nil {
		utils.AddError(errs, "unable to fetch openid configuration", err)
		errs.Code = 400
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		utils.AddError(errs, "platform openid configuration returned non-success status", resp.Status)
		errs.Code = 400
		return nil, false
	}

	var config platformOpenIDConfiguration
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		utils.AddError(errs, "unable to decode platform openid configuration", err)
		errs.Code = 400
		return nil, false
	}

	if strings.TrimSpace(config.Issuer) == "" || strings.TrimSpace(config.RegistrationEndpoint) == "" || strings.TrimSpace(config.AuthorizationEndpoint) == "" || strings.TrimSpace(config.TokenEndpoint) == "" || strings.TrimSpace(config.JWKSURI) == "" {
		utils.AddError(errs, "platform openid configuration is missing required fields", config)
		errs.Code = 400
		return nil, false
	}

	return &config, true
}

func registerOnPlatform(
	platformConfig *platformOpenIDConfiguration,
	registrationToken string,
	customerID string,
	toolBaseURL string,
	toolRedirectURI string,
	initiateLoginURI string,
	errs *utils.JsonErrors,
) (string, bool) {
	registrationPayload := buildToolRegistrationPayload(customerID, toolBaseURL, toolRedirectURI, initiateLoginURI, platformConfig.ClaimsSupported, platformConfig.MessagesSupported, platformConfig.ScopesSupported)

	body, err := json.Marshal(registrationPayload)
	if err != nil {
		utils.AddError(errs, "unable to serialize registration payload", err)
		errs.Code = 500
		return "", false
	}

	req, err := http.NewRequest("POST", platformConfig.RegistrationEndpoint, bytes.NewBuffer(body))
	if err != nil {
		utils.AddError(errs, "unable to create platform registration request", err)
		errs.Code = 400
		return "", false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if registrationToken != "" {
		req.Header.Set("Authorization", "Bearer "+registrationToken)
	}

	resp, err := utils.HttpClient().Do(req)
	if err != nil {
		utils.AddError(errs, "platform registration request failed", err)
		errs.Code = 400
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		utils.AddError(errs, "platform registration returned non-success status", resp.Status)
		errs.Code = 400
		return "", false
	}

	var response platformRegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		utils.AddError(errs, "unable to decode platform registration response", err)
		errs.Code = 400
		return "", false
	}
	if strings.TrimSpace(response.ClientID) == "" {
		utils.AddError(errs, "platform registration response did not include client_id", response)
		errs.Code = 400
		return "", false
	}

	return response.ClientID, true
}

func resolveToolBaseURL(r *http.Request) string {
	if configured := strings.TrimSpace(os.Getenv("TOOL_BASE_URL")); configured != "" {
		return strings.TrimSuffix(configured, "/")
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = forwarded
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func resolveToolName(customerID string) string {
	if configured := strings.TrimSpace(os.Getenv("TOOL_NAME")); configured != "" {
		return configured
	}
	if customerID != "" {
		return "Asset Processor - " + customerID
	}
	return "Asset Processor"
}

func hostWithoutPort(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func appendQueryValue(rawURL string, key string, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}

func buildLoginURL(toolBaseURL string, registrationID string) string {
	return appendQueryValue(toolBaseURL+"/oidc/login", "reg_id", registrationID)
}

func isJSONRequest(r *http.Request) bool {
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	accept := strings.ToLower(strings.TrimSpace(r.Header.Get("Accept")))
	return strings.Contains(contentType, "application/json") || strings.Contains(accept, "application/json")
}

func renderRegistrationComplete(w http.ResponseWriter, result map[string]string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(201)
	fmt.Fprintf(w, `<!doctype html>
<html>
	<head>
		<meta charset="utf-8" />
		<title>Registration Complete</title>
	</head>
	<body>
		<h2>Registration Complete</h2>
		<p>Registration and deployment were created successfully.</p>
		<p><strong>Registration ID:</strong> %s</p>
		<p><strong>Deployment ID:</strong> %s</p>
		<p><strong>Client ID:</strong> %s</p>
		<p><strong>Login URL:</strong> %s</p>
		<p>The platform window will be asked to close now.</p>
		<script>
			(window.opener || window.parent).postMessage({subject:'org.imsglobal.lti.close'}, '*');
			setTimeout(function () { window.close(); }, 300);
		</script>
	</body>
</html>`,
		html.EscapeString(result["registration_id"]),
		html.EscapeString(result["deployment_id"]),
		html.EscapeString(result["client_id"]),
		html.EscapeString(result["login_url"]),
	)
}
