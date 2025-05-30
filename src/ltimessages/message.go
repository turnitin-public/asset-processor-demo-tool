package ltimessages

import (
	"1edtech/ap-demo/oidc"
	"1edtech/ap-demo/utils"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt"
)

type LtiMessage struct {
	MessageType  string `json:"https://purl.imsglobal.org/spec/lti/claim/message_type,omitempty"`
	DeploymentId string `json:"https://purl.imsglobal.org/spec/lti/claim/deployment_id,omitempty"`
	Context      *struct {
		Id string `json:"id,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/context,omitempty"`
	Asset *struct {
		Id string `json:"id,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/asset,omitempty"`
	ReportType string `json:"https://purl.imsglobal.org/spec/lti/claim/assetreport_type"`
	Submission *struct {
		Id string `json:"id,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/submission,omitempty"`
	ForUser *struct {
		UserId string `json:"user_id,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/for_user,omitempty"`
	Activity *struct {
		Id    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/activity,omitempty"`
	ResourceLink *struct {
		Id    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/resource_link,omitempty"`
	Roles []string `json:"https://purl.imsglobal.org/spec/lti/claim/roles,omitempty"`
	Dl    *struct {
		Data              *string  `json:"data,omitempty"`
		DeepLinkReturnUrl string   `json:"deep_link_return_url,omitempty"`
		AcceptTypes       []string `json:"accept_types,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti-dl/claim/deep_linking_settings,omitempty"`
	Pns *struct {
		ServiceUrl           string   `json:"platform_notification_service_url,omitempty"`
		NoticeTypesSupported []string `json:"notice_types_supported,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/platformnotificationservice,omitempty"`
	Eula *struct {
		Url   string   `json:"url,omitempty"`
		Scope []string `json:"scope,omitempty"`
	} `json:"https://purl.imsglobal.org/spec/lti/claim/eulaservice,omitempty"`
	TargetLinkUri string `json:"https://purl.imsglobal.org/spec/lti/claim/target_link_uri,omitempty"`
	Nonce         string `json:"nonce,omitempty"`
	jwt.StandardClaims
}

func Handler(w http.ResponseWriter, r *http.Request) {
	errs, c := oidc.Validator().ValidateIdToken(getIdToken(w, r), &LtiMessage{})
	claims := c.(*LtiMessage)
	if len(errs.Errors) > 0 {
		// Log all errors
		for _, err := range errs.Errors {
			log.Print(err)
		}
		utils.UiError(w, errs.Code, "Invalid id token")
		return
	}
	// Switch on message type
	switch claims.MessageType {
	case "LtiDeepLinkingRequest":
		// Handle deep linking request
		errs = deepLinkingRequest(w, r, claims)
	case "LtiResourceLinkRequest":
		// Handle resource link request
		errs = resourceLinkRequest(w, r, claims)
	case "LtiAssetProcessorSettingsRequest":
		// Handle asset processor settings request
		errs = assetProcessorSettingsRequest(w, r, claims)
	case "LtiReportReviewRequest":
		// Handle report review request
		errs = reportReviewRequest(w, r, claims)
	case "LtiEulaRequest":
		// Handle EULA request
		errs = eulaRequest(w, r, claims)
	}

	if len(errs.Errors) > 0 {
		// Log all errors
		for _, err := range errs.Errors {
			log.Print(err)
		}
		utils.UiError(w, errs.Code, "Error processing message")
	}
}

// Get id token from request
func getIdToken(w http.ResponseWriter, r *http.Request) string {
	// Parse form parameters
	if err := r.ParseForm(); err != nil {
		utils.UiError(w, 400, "Unable to parse form")
		return ""
	}
	// Get id token
	id_token := r.Form.Get("id_token")
	if len(id_token) == 0 {
		utils.UiError(w, 400, "Id token cannot be empty")
		return ""
	}
	return id_token
}
