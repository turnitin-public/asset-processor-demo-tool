package ltinotices

import (
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/oidc"
	"1edtech/ap-demo/utils"
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
)

type LtiNotice struct {
	Notice *struct {
		Type string `json:"type,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/notice,omitempty"`
	DeploymentId string `json:"https://purl.imsglobal.org/spec/lti/claim/deployment_id,omitempty"`
	Context      *struct {
		Id string `json:"id,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/context,omitempty"`
	Activity *struct {
		Id string `json:"id,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/activity,omitempty"`
	ForUser *struct {
		UserId string `json:"user_id,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/for_user,omitempty"`
	Submission *struct {
		Id string `json:"id,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/submission,omitempty"`
	AssetService *struct {
		Scope  []string            `json:"scope,omitempty"`
		Assets []ltiservices.Asset `json:"assets,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/assetservice,omitempty"`
	AssetReport *struct {
		Scope     []string `json:"scope,omitempty"`
		ReportUrl string   `json:"report_url,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/assetreport,omitempty"`
	TargetLinkUri string `json:"https://purl.imsglobal.org/spec/lti/claim/target_link_uri,omitempty"`
	Nonce         string `json:"nonce,omitempty"`
	jwt.StandardClaims
}

type BatchNotice struct {
	Notices *[]struct {
		Jwt string `json:"jwt,omitempty"`
	} `json:"notices,omitempty"`
}

func BatchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decodeErrs := utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 400}
	// Get json body
	decoder := json.NewDecoder(r.Body)
	var request_data BatchNotice
	err := decoder.Decode(&request_data)
	if err != nil {
		log.Print(err)
		utils.AddError(&decodeErrs, "Unable to parse request", err)
		utils.WriteJsonError(w, r, decodeErrs)
		return
	}
	allErrs := utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 400}
	// Loop through notices
	for _, notice := range *request_data.Notices {
		errs := handler(w, r, notice.Jwt)
		allErrs.Errors = append(allErrs.Errors, errs.Errors...)
	}
	if len(allErrs.Errors) > 0 {
		utils.WriteJsonError(w, r, allErrs)
		return
	}
	w.WriteHeader(201)
}

func handler(w http.ResponseWriter, r *http.Request, idToken string) utils.JsonErrors {
	errs, c := oidc.Validator().ValidateIdToken(idToken, &LtiNotice{})
	if len(errs.Errors) > 0 || c == nil {
		// Log all errors
		for _, err := range errs.Errors {
			log.Print(err)
		}
		return errs
	}
	// Switch on message type
	claims := c.(*LtiNotice)
	switch claims.Notice.Type {
	case "LtiAssetProcessorSubmissionNotice":
		// Handle asset processor submission notice
		errs = assetProcessorSubmissionNotice(w, r, claims)
	}
	if len(errs.Errors) > 0 {
		// Log all errors
		for _, err := range errs.Errors {
			log.Print(err)
		}
	}
	return errs
}
