package ltiservices

import (
	"1edtech/ap-demo/utils"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type Asset struct {
	Id          string `json:"asset_id,omitempty"`
	Url         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	FileName    string `json:"filename,omitempty"`
	Checksum    string `json:"sha256_checksum,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	Size        int    `json:"size,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type DownloadedAsset struct {
	Id     string `json:"id"`
	Asset  Asset  `json:"asset"`
	UserId string `json:"userId"`
	Path   string `json:"path"`
}

func FetchAssets(issuer string, clientId string, deploymentId string, assets []Asset, userId string, scopes []string, errs *utils.JsonErrors) []DownloadedAsset {
	// Get service token
	token, ok := GetClientServiceToken(issuer, clientId, deploymentId, scopes, errs)
	if !ok {
		return nil
	}
	downloadedAssets := make([]DownloadedAsset, 0)
	for _, asset := range assets {
		req, err := http.NewRequest("GET", asset.Url, nil)
		if err != nil {
			utils.AddError(errs, "Error creating request", err)
			errs.Code = 401
			continue
		}
		req.Header.Add("Authorization", "Bearer "+token.AccessToken)
		client := utils.HttpClient()
		resp, err := client.Do(req)
		if err != nil {
			utils.AddError(errs, "Fetching asset", err)
			errs.Code = 401
			continue
		}
		if resp.StatusCode >= 400 {
			utils.AddError(errs, "Fetching asset returned a non-success response", resp)
			errBody, _ := io.ReadAll(resp.Body)
			utils.AddError(errs, "err body", string(errBody))
			errs.Code = resp.StatusCode
			continue
		}
		// Save asset
		id := uuid.New().String()
		assetPath := "/tmp/" + id
		out, err := os.Create(assetPath)
		if err != nil {
			utils.AddError(errs, "Error creating asset file", err)
			errs.Code = 401
			continue
		}
		defer out.Close()
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			utils.AddError(errs, "Error saving asset", err)
			errs.Code = 401
			continue
		}
		downloadedAssets = append(downloadedAssets, DownloadedAsset{Id: id, Asset: asset, Path: assetPath, UserId: userId})
	}

	return downloadedAssets
}
