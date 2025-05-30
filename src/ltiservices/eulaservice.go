package ltiservices

import (
	"1edtech/ap-demo/utils"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func AcceptEula(issuer string, clientId string, deploymentId string, serviceUrl string, scopes []string, userId string, errs *utils.JsonErrors) bool {
	// Get service token
	token, ok := GetClientServiceToken(issuer, clientId, deploymentId, scopes, errs)
	if !ok {
		return false
	}
	EulaAcceptanceRequest := struct {
		UserId    string `json:"userId"`
		Accepted  bool   `json:"accepted"`
		Timestamp string `json:"timestamp"`
	}{
		UserId:    userId,
		Accepted:  true,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	b, err := json.Marshal(EulaAcceptanceRequest)
	if err != nil {
		utils.AddError(errs, "Error marshalling EULA acceptance request", err)
		errs.Code = 401
		return false
	}
	// Create request to accept EULA
	req, err := http.NewRequest("POST", serviceUrl+"/user", bytes.NewBuffer(b))
	log.Printf("Making request to accept EULA: %s body: %s Authorization: Bearer %s", serviceUrl, string(b), token.AccessToken)
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
		utils.AddError(errs, "Accepting EULA", err)
		errs.Code = 401
		return false
	}
	if resp.StatusCode >= 400 {
		utils.AddError(errs, "Accept EULA returned a non-success response", resp)
		errs.Code = resp.StatusCode
		return false
	}
	return true
}
