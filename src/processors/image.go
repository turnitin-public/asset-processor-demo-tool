package processors

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/utils"
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"time"

	"github.com/Kagami/go-face"
)

type imageProcessorType struct{}

var imageProcessor IProcessor = imageProcessorType{}

func (imageProcessorType) GetName() string {
	return "imageProcessor"
}

func (imageProcessorType) GetType() string {
	return "image"
}

func (imageProcessorType) CanBeUsed(asset ltiservices.Asset) bool {
	// Check if the asset is an image
	return asset.ContentType == "image/jpeg"
}

func (imageProcessorType) Process(registrationId string, deploymentId string, asset ltiservices.DownloadedAsset, errs *utils.JsonErrors) (bool, *ltiservices.Report) {
	// Process image
	// Initialize recognizer
	rec, err := face.NewRecognizer("/modules")
	if err != nil {
		utils.AddError(errs, "Can't init face recognizer", err)
	}
	// Free the resources when you're finished.
	defer rec.Close()

	// Now let's try to classify some not yet known image.
	foundFaces, err := rec.RecognizeFile(asset.Path)
	if err != nil {
		utils.AddError(errs, "Can't recognize", err)
	}
	if foundFaces == nil {
		utils.AddError(errs, "Could not find faces", nil)
	}

	// Save report to DB
	if err != nil {
		utils.AddError(errs, "Error decoding llm response", err)
		errs.Code = 401
		return false, nil
	}
	comment := fmt.Sprintf("There are %d people in attendance", len(foundFaces))
	log.Println(comment)
	ok := datastore.AssetReportQueries.SaveAssetReport(asset.Id, registrationId, deploymentId, asset.Asset.Id, "image", comment)
	if !ok {
		utils.AddError(errs, "Error saving asset report", nil)
		errs.Code = 500
		return false, nil
	}

	given := float64(len(foundFaces))

	return true, &ltiservices.Report{
		AssetId:            asset.Asset.Id,
		Type:               "image",
		Timestamp:          time.Now().Format(time.RFC3339),
		Title:              asset.Asset.Title,
		Result:             fmt.Sprintf("%d people", len(foundFaces)),
		IndicationColor:    fmt.Sprintf("#00%02x00", int(math.Min(given*40, 255))),
		IndicationAlt:      "Good",
		Priority:           0,
		ProcessingProgress: "Processed",
		Comment:            &comment,
		ErrorCode:          nil,
	}
}

func (imageProcessorType) GetFileHtml(assetId string) (template.HTML, error) {
	// Get asset file as string
	assetFile, err := os.ReadFile("/tmp/" + assetId)
	if err != nil {
		return "Error reading asset file", err
	}
	b64Image := base64.StdEncoding.EncodeToString(assetFile)
	t, err := template.New("image").Parse(`{{define "T"}}<img src="data:image/png;base64,{{.}}" />{{end}}`)
	if err != nil {
		return "Error parsing template", err
	}
	out := new(bytes.Buffer)
	err = t.ExecuteTemplate(out, "T", b64Image)
	return template.HTML(out.String()), err
}
