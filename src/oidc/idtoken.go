package oidc

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
)

var validator IdTokenValidator

func Validator() IdTokenValidator {
	if validator == nil {
		validator = &DefaultIdTokenValidator{}
	}
	return validator
}

func SetValidator(v IdTokenValidator) {
	validator = v
}

type IdTokenValidator interface {
	ValidateIdToken(idToken string, claimInterface jwt.Claims) (utils.JsonErrors, jwt.Claims)
}

type DefaultIdTokenValidator struct{}

// Validate LTI launch
func (DefaultIdTokenValidator) ValidateIdToken(idToken string, claimInterface jwt.Claims) (utils.JsonErrors, jwt.Claims) {
	errs := utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 400}

	// Split id token
	parts := strings.Split(idToken, ".")

	// Header and Body structures
	type UnsignedBody struct {
		Issuer   string      `json:"iss"`
		ClientID interface{} `json:"aud"`
		IssuedAt int64       `json:"iat"`
		Expires  int64       `json:"exp"`
	}
	var body_json UnsignedBody
	type JwtHead struct {
		KeyId string `json:"kid"`
	}
	var header_json JwtHead

	// Decode body
	if body_string, err := jwt.DecodeSegment(parts[1]); err != nil {
		utils.AddError(&errs, "Unable to decode id token body.", err)
		// Get issuer and client id
	} else if err := json.Unmarshal(body_string, &body_json); err != nil {
		utils.AddError(&errs, "Unable to unmarshal id token body.", err)
	}

	// Decode header
	if header_string, err := jwt.DecodeSegment(parts[0]); err != nil {
		utils.AddError(&errs, "Unable to decode id token header.", err)
		// Get KID
	} else if err := json.Unmarshal(header_string, &header_json); err != nil {
		utils.AddError(&errs, "Unable to unmarshal id token header.", err)
	}

	// Check we haven't hit any errors yet
	if len(errs.Errors) > 0 {
		return errs, nil
	}

	// Check token hasn't expired
	if body_json.IssuedAt == 0 || body_json.IssuedAt > (time.Now().Unix()+5) {
		utils.AddError(&errs, "Token is issued in the future.", body_json)
	}

	if body_json.Expires == 0 || body_json.Expires < (time.Now().Unix()-5) {
		utils.AddError(&errs, "Token has expired.", body_json)
	}

	// Validate issuer
	if len(body_json.Issuer) == 0 {
		utils.AddError(&errs, "Issue cannot be empty.", body_json)
	}

	// Client id array or string
	switch obj := body_json.ClientID.(type) {
	case string:
		body_json.ClientID = obj
	case []interface{}:
		if len(obj) != 1 {
			utils.AddError(&errs, "Invalid audience claim, array must only contain one value.", body_json)
			return errs, nil
		}
		ok := true
		body_json.ClientID, ok = obj[0].(string)
		if !ok {
			utils.AddError(&errs, "Invalid audience claim, must be a string or array.", body_json)
			return errs, nil
		}
	}

	// Validate client id
	if len(body_json.ClientID.(string)) == 0 {
		utils.AddError(&errs, "Client id cannot be empty.", body_json)
	}

	// Validate key id
	if len(header_json.KeyId) == 0 {
		utils.AddError(&errs, "Key id cannot be empty.", header_json)
	}

	// Check we haven't hit any errors yet
	if len(errs.Errors) > 0 {
		return errs, nil
	}

	// Find registration
	registration, err := datastore.RegistrationQueries.GetRegistrationByClient(body_json.Issuer, body_json.ClientID.(string))
	if err != nil {
		utils.AddError(&errs, "Unable to find registration", err)
		return errs, nil
	}
	// Fetch JWKS
	log.Printf("registration: %v", registration)
	resp, err := http.Get(registration.PlatformJwksEndpoint)
	if err != nil {
		utils.AddError(&errs, "Unable to fetch JWKS endpoint", err)
		return errs, nil
	}
	jwks_string, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.AddError(&errs, "Unable to read JWKS response", err)
		return errs, nil
	}
	// Parse JWKS
	jwks, err := jwk.Parse(jwks_string)
	if err != nil {
		utils.AddError(&errs, "Unable to parse JWKS", err)
		return errs, nil
	}
	// find JWK
	platJwk, ok := jwks.LookupKeyID(header_json.KeyId)
	if !ok {
		utils.AddError(&errs, "Key id not found", err)
		return errs, nil
	}
	// Validate JWT signature
	token, err := jwt.ParseWithClaims(idToken, claimInterface, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		var rawKey interface{}
		platJwk.Raw(&rawKey)
		return rawKey, nil
	})
	if err != nil {
		utils.AddError(&errs, "Invalid signature on id token", err)
		errs.Code = 401
		return errs, nil
	}

	return errs, token.Claims
}
