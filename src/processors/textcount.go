package processors

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/utils"
	"bytes"
	"fmt"
	"html/template"
	"math"
	"os"
	"strings"
	"time"
)

type textCountProcessorType struct{}

var textCountProcessor IProcessor = textCountProcessorType{}

func (textCountProcessorType) GetName() string {
	return "textCountProcessor"
}

func (textCountProcessorType) GetType() string {
	return "textCount"
}

func (textCountProcessorType) CanBeUsed(asset ltiservices.Asset) bool {
	// Check if the asset is a text file
	return asset.ContentType == "text/plain"
}

func (textCountProcessorType) Process(registrationId string, deploymentId string, asset ltiservices.DownloadedAsset, errs *utils.JsonErrors) (bool, *ltiservices.Report) {
	assetFile, err := os.ReadFile(asset.Path)
	if err != nil {
		utils.AddError(errs, "Error reading asset file", err)
		errs.Code = 401
		return false, nil
	}
	given := float64(len(strings.Split(string(assetFile), " ")))
	msg := fmt.Sprintf("The word count is: %d", int(given))
	ok := datastore.AssetReportQueries.SaveAssetReport(asset.Id, registrationId, deploymentId, asset.Asset.Id, "textCount", msg)
	if !ok {
		utils.AddError(errs, "Error saving asset report", nil)
		errs.Code = 500
		return false, nil
	}
	return true, &ltiservices.Report{
		AssetId:            asset.Asset.Id,
		Type:               "textCount",
		Timestamp:          time.Now().Format(time.RFC3339),
		Title:              asset.Asset.Title,
		Result:             fmt.Sprintf("%d words", int(given)),
		IndicationColor:    fmt.Sprintf("#0000%02x", int(math.Min(given*40, 255))),
		IndicationAlt:      "Good",
		Priority:           0,
		ProcessingProgress: "Processed",
		Comment:            &msg,
		ErrorCode:          nil,
	}
}

func (textCountProcessorType) GetFileHtml(assetId string) (template.HTML, error) {
	// Get asset file as string
	assetFile, err := os.ReadFile("/tmp/" + assetId)
	if err != nil {
		return "Error reading asset file", err
	}
	t, err := template.New("text").Parse(`{{define "T"}}<pre>{{.}}</pre>{{end}}`)
	if err != nil {
		return "Error parsing template", err
	}
	out := new(bytes.Buffer)
	err = t.ExecuteTemplate(out, "T", string(assetFile))
	return template.HTML(out.String()), err
}
