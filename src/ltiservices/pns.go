package ltiservices

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/utils"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func RegisterSubmissionNotice(issuer string, clientId string, deploymentId string, serviceUrl string, scopes []string, errs *utils.JsonErrors) bool {
	// Register submission notice

	// Get service token
	token, ok := GetClientServiceToken(issuer, clientId, deploymentId, scopes, errs)
	if !ok {
		return false
	}
	reg, err := datastore.RegistrationQueries.GetRegistrationByClient(issuer, clientId)
	if err != nil {
		utils.AddError(errs, "Error getting registration", err)
		errs.Code = 401
		return false
	}

	tunnelUrl := os.Getenv("TUNNEL_URL")
	parsedUrl, err := url.Parse(tunnelUrl)
	if err != nil {
		utils.AddError(errs, "Error parsing TUNNEL_URL", err)
		errs.Code = 401
		return false
	}
	redirectUrl, err := url.Parse(reg.ToolRedirectUri)
	if err != nil {
		utils.AddError(errs, "Error parsing ToolRedirectUri", err)
		errs.Code = 401
		return false
	}
	handlerUrl := redirectUrl.Scheme + "://" + redirectUrl.Host + "/lti/notice"
	if os.Getenv("HANDLER_PROXY") != "" {
		handlerUrl = "https://" + os.Getenv("HANDLER_PROXY") + "/" + parsedUrl.Hostname()
	}
	pnsRequest := struct {
		NoticeType string `json:"notice_type"`
		Handler    string `json:"handler"`
	}{
		NoticeType: "LtiAssetProcessorSubmissionNotice",
		Handler:    handlerUrl,
	}
	b, err := json.Marshal(pnsRequest)
	if err != nil {
		utils.AddError(errs, "Error marshalling pns request", err)
		errs.Code = 401
		return false
	}
	req, err := http.NewRequest("PUT", serviceUrl, bytes.NewBuffer(b))
	if err != nil {
		utils.AddError(errs, "Error creating request", err)
		errs.Code = 401
		return false
	}
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	client := utils.HttpClient()
	resp, err := client.Do(req)
	if err != nil {
		utils.AddError(errs, "Registering for submission notices", err)
		errs.Code = 401
		return false
	}
	if resp.StatusCode >= 400 {
		utils.AddError(errs, "Register for submission notices returned a non-success response", resp)
		errBody, _ := io.ReadAll(resp.Body)
		utils.AddError(errs, "err body", string(errBody))
		errs.Code = resp.StatusCode
		return false
	}
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Registered for submission notices: %s", string(body))

	return true
}
