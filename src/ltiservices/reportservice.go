package ltiservices

import (
	"1edtech/ap-demo/utils"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Report struct {
	AssetId            string  `json:"assetId"`
	Type               string  `json:"type"`
	Timestamp          string  `json:"timestamp"`
	Title              string  `json:"title"`
	Result             string  `json:"result"`
	IndicationColor    string  `json:"indicationColor"`
	IndicationAlt      string  `json:"indicationAlt"`
	Priority           int16   `json:"priority"`
	ProcessingProgress string  `json:"processingProgress"`
	Comment            *string `json:"comment"`
	ErrorCode          *string `json:"errorCode"`
}

func SendReports(issuer string, clientId string, deploymentId string, serviceUrl string, scopes []string, reports []Report, errs *utils.JsonErrors) bool {
	// Get service token
	token, ok := GetClientServiceToken(issuer, clientId, deploymentId, scopes, errs)
	if !ok {
		return false
	}
	// Send each report
	for _, report := range reports {
		b, err := json.Marshal(report)
		if err != nil {
			utils.AddError(errs, "Error marshalling report", err)
			errs.Code = 401
			continue
		}
		req, err := http.NewRequest("POST", serviceUrl, bytes.NewBuffer(b))
		if err != nil {
			utils.AddError(errs, "Error creating request", err)
			errs.Code = 401
			continue
		}
		req.Header.Add("Authorization", "Bearer "+token.AccessToken)
		req.Header.Set("Content-Type", "application/json")
		client := utils.HttpClient()
		resp, err := client.Do(req)
		if err != nil {
			utils.AddError(errs, "Error ending report", err)
			errs.Code = 401
			continue
		}
		if resp.StatusCode >= 400 {
			utils.AddError(errs, "Ending report returned a non-success response", resp)
			errBody, _ := io.ReadAll(resp.Body)
			utils.AddError(errs, "err body", string(errBody))
			errs.Code = resp.StatusCode
			continue
		}
	}
	return len(errs.Errors) == 0
}
