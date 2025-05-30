package ltiservices

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/utils"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type AccessTokenResp struct {
	AccessToken string `json:"access_token,omitempty"`
}

// GetClientServiceToken makes a call to an LTI platform to fetch a service token with the specified scopes.
func GetClientServiceToken(issuer string, clientId string, deploymentId string, scopes []string, errs *utils.JsonErrors) (*AccessTokenResp, bool) {
	privateKey, registration, ok := datastore.RegistrationQueries.GetPrivateKeyAndRegForClient(issuer, clientId, errs)
	if !ok {
		return nil, false
	}

	// Build request JWT
	authProvider := registration.PlatformServiceAuthEndpoint
	if registration.PlatformAuthProvider != nil {
		authProvider = *registration.PlatformAuthProvider
	}
	type serviceAuthClaims struct {
		DeploymentId string `json:"https://purl.imsglobal.org/spec/lti/claim/deployment_id,omitempty"`
		jwt.StandardClaims
	}
	serviceAuthJwt := serviceAuthClaims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    registration.ClientId,
			Subject:   registration.ClientId,
			Audience:  authProvider,
			IssuedAt:  time.Now().Unix() - 5,
			ExpiresAt: time.Now().Unix() + 60,
			Id:        uuid.New().String(),
		},
		DeploymentId: deploymentId,
	}
	t := jwt.New(jwt.SigningMethodRS256)

	t.Claims = &serviceAuthJwt
	t.Header = map[string]interface{}{
		"kid": registration.Kid,
		"alg": registration.Alg,
	}

	signedAuthJwt, err := t.SignedString(privateKey)
	if err != nil {
		utils.AddError(errs, "Unable to sign service token request", err)
		errs.Code = 500
		return nil, false
	}

	// Fetch client token
	data := url.Values{
		"grant_type":            {"client_credentials"},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
		"client_assertion":      {signedAuthJwt},
		"scope":                 {strings.Join(scopes, " ")},
	}

	log.Printf("%+v", data)

	resp, err := http.PostForm(registration.PlatformServiceAuthEndpoint, data)
	if err != nil {
		utils.AddError(errs, "Error Fetching service token from platform", err)
		errs.Code = 401
		return nil, false
	}

	accessTokenRespString, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.AddError(errs, "Invalid response from platform service token endpoint", err)
		errs.Code = 401
		return nil, false
	}

	log.Printf("%+v", string(accessTokenRespString))

	var accessTokenResp AccessTokenResp

	err = json.Unmarshal(accessTokenRespString, &accessTokenResp)
	if err != nil {
		utils.AddError(errs, "Invalid json in response from platform service token endpoint", err)
		return nil, false
	}

	return &accessTokenResp, true

}
