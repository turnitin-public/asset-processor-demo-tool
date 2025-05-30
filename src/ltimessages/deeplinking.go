package ltimessages

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/utils"
	"html/template"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func deepLinkingRequest(w http.ResponseWriter, r *http.Request, claims *LtiMessage) utils.JsonErrors {
	// Register for submission notices
	var errs = utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}
	ltiservices.RegisterSubmissionNotice(claims.Issuer, claims.Audience, claims.DeploymentId, claims.Pns.ServiceUrl, []string{"https://purl.imsglobal.org/spec/lti/scope/noticehandlers"}, &errs)
	if len(errs.Errors) > 0 {
		return errs
	}
	// Load Template
	t, err := template.ParseFiles("templates/deeplinking.html")
	if err != nil {
		utils.AddError(&errs, "Unable to load template", err)
		errs.Code = 500
		return errs
	}

	// Get first supported accept type
	acceptType := "ltiResourceLink"
	acceptedTypes := []string{"ltiResourceLink", "ltiAssetProcessor"}
	for _, at := range claims.Dl.AcceptTypes {
		for _, supportedType := range acceptedTypes {
			if at == supportedType {
				acceptType = at
				break
			}
		}
	}

	data := struct {
		Issuer          string
		ClientId        string
		DeploymentId    string
		DlReturnUrl     string
		Data            *string
		ContentItemType string
	}{
		Issuer:          claims.Issuer,
		ClientId:        claims.Audience,
		DeploymentId:    claims.DeploymentId,
		DlReturnUrl:     claims.Dl.DeepLinkReturnUrl,
		Data:            claims.Dl.Data,
		ContentItemType: acceptType,
	}
	err = t.Execute(w, data)
	if err != nil {
		utils.AddError(&errs, "Unable to render template", err)
		errs.Code = 500
	}

	return errs
}

func DeepLinkingResponse(w http.ResponseWriter, r *http.Request) {
	var errs = utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}

	if err := r.ParseForm(); err != nil {
		utils.AddError(&errs, "Error reading deep link request", err)
		errs.Code = 400
		utils.WriteJsonError(w, r, errs)
		return
	}

	type ContentItemReport struct {
		Indicator      bool     `json:"indicator"`
		SupportedTypes []string `json:"supportedTypes"`
	}
	type LineItem struct {
		ScoreMaximum float64 `json:"scoreMaximum"`
	}
	type ContentItem struct {
		Type     string            `json:"type"`
		Title    string            `json:"title"`
		Custom   map[string]string `json:"custom"`
		Report   ContentItemReport `json:"report"`
		LineItem LineItem          `json:"lineItem"`
	}
	var deepLinkRequest struct {
		jwt.StandardClaims
		Nonce        string        `json:"nonce,omitempty"`
		DeploymentId string        `json:"https://purl.imsglobal.org/spec/lti/claim/deployment_id,omitempty"`
		Data         *string       `json:"https://purl.imsglobal.org/spec/lti-dl/claim/data,omitempty"`
		MessageType  string        `json:"https://purl.imsglobal.org/spec/lti/claim/message_type"`
		Version      string        `json:"https://purl.imsglobal.org/spec/lti/claim/version"`
		ContentItems []ContentItem `json:"https://purl.imsglobal.org/spec/lti-dl/claim/content_items"`
	}

	dlData := r.Form.Get("data")

	deepLinkResponseData := struct {
		Issuer       string  `json:"iss"`
		Audience     string  `json:"aud"`
		DeploymentId string  `json:"deployment_id"`
		Data         *string `json:"data"`
		ResponseUrl  string  `json:"response_url"`
	}{
		Issuer:       r.Form.Get("iss"),
		Audience:     r.Form.Get("aud"),
		DeploymentId: r.Form.Get("deployment_id"),
		Data:         &dlData,
		ResponseUrl:  r.Form.Get("response_url"),
	}
	privateKey, registration, ok := datastore.RegistrationQueries.GetPrivateKeyAndRegForClient(deepLinkResponseData.Issuer, deepLinkResponseData.Audience, &errs)
	if !ok {
		utils.WriteJsonError(w, r, errs)
		return
	}
	deepLinkRequest.Issuer = registration.ClientId
	deepLinkRequest.Audience = registration.Issuer
	deepLinkRequest.DeploymentId = deepLinkResponseData.DeploymentId
	deepLinkRequest.Version = "1.3.0"
	deepLinkRequest.MessageType = "LtiDeepLinkingResponse"
	deepLinkRequest.ContentItems = make([]ContentItem, 0)

	if r.Form.Get("content_item_type") == "ltiAssetProcessor" {
		deepLinkRequest.ContentItems = append(deepLinkRequest.ContentItems, ContentItem{
			Type:  "ltiAssetProcessor",
			Title: "Example Content",
			Custom: map[string]string{
				"process_text":   r.Form.Get("process_text"),
				"process_images": r.Form.Get("process_images"),
			},
			Report: ContentItemReport{
				SupportedTypes: []string{
					"text",
					"image",
				},
			},
		})
	}
	if r.Form.Get("content_item_type") == "ltiResourceLink" {
		deepLinkRequest.ContentItems = append(deepLinkRequest.ContentItems, ContentItem{
			Type:  "ltiResourceLink",
			Title: "Example Resource Link",
			Custom: map[string]string{
				"something_custom": "some value",
			},
			LineItem: LineItem{
				ScoreMaximum: 100.0,
			},
		})
	}
	deepLinkRequest.IssuedAt = time.Now().Unix() - 5
	deepLinkRequest.ExpiresAt = time.Now().Unix() + 120
	deepLinkRequest.Nonce = uuid.New().String()

	if deepLinkResponseData.Data != nil {
		deepLinkRequest.Data = deepLinkResponseData.Data
	}

	t := jwt.New(jwt.SigningMethodRS256)

	t.Claims = &deepLinkRequest
	t.Header = map[string]interface{}{
		"kid": registration.Kid,
		"alg": registration.Alg,
	}

	signedDlJwt, err := t.SignedString(privateKey)
	if err != nil {
		utils.AddError(&errs, "Unable to sign service token request", err)
		errs.Code = 500
		utils.WriteJsonError(w, r, errs)
		return
	}

	data := struct {
		Action string
		Params map[string]string
	}{
		Action: deepLinkResponseData.ResponseUrl,
		Params: map[string]string{"JWT": signedDlJwt},
	}
	te, err := template.ParseFiles("templates/autopost.html")
	if err != nil {
		utils.AddError(&errs, "Failed to parse template", err)
		errs.Code = 500
		utils.WriteJsonError(w, r, errs)
		return
	}
	te.Execute(w, data)

}
