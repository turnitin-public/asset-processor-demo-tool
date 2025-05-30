package processors

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/utils"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
	"time"
)

// Interface for types of processors.
// Each processor will be run by each asset in the submission so that it can first identify if it is capable of processing the asset.
// Then it will process the asset and return a report.
type IProcessor interface {
	// The user friendly name of the processor
	GetName() string
	// The type of the processor returned in reports by that processor
	GetType() string
	// A lightweight check to identify from an asset if the processor is able to process it.
	// This usually checks the content type of the asset to allow it to be processed.
	CanBeUsed(asset ltiservices.Asset) bool
	// The main bulk of the processing done by the processor.
	// It must handle the processing of the asset along with the generation and saving of the report.
	Process(registrationId, deploymentId string, asset ltiservices.DownloadedAsset, errs *utils.JsonErrors) (bool, *ltiservices.Report)
	// A snippet of HTML that can be used to embed the original submitted file into the report.
	GetFileHtml(internalAssetId string) (template.HTML, error)
}

var processors []IProcessor

func FilterAssets(assets []ltiservices.Asset, errs *utils.JsonErrors) ([]ltiservices.Asset, []ltiservices.Report) {
	filteredAssets := make([]ltiservices.Asset, 0)
	reports := make([]ltiservices.Report, 0)
	// Validate each asset
	for _, asset := range assets {
		processable := false
		for _, processor := range getProcessors() {
			if processor.CanBeUsed(asset) {
				processable = true
				filteredAssets = append(filteredAssets, asset)
				break
			}
		}
		if !processable {
			reports = append(reports, *processUnsupportedError(asset, errs))
		}
	}
	return filteredAssets, reports
}

func ProcessAssets(issuer string, clientId string, deploymentId string, assets []ltiservices.DownloadedAsset, errs *utils.JsonErrors) []ltiservices.Report {
	reports := make([]ltiservices.Report, 0)
	// Get registrationId
	reg, err := datastore.RegistrationQueries.GetRegistrationByClient(issuer, clientId)
	if err != nil {
		utils.AddError(errs, "Error getting registration", err)
		errs.Code = 404
		return reports
	}
	// Process each asset
	for _, asset := range assets {
		// Validate asset
		ok, report := validateAsset(asset, errs)
		if !ok {
			reports = append(reports, *report)
			continue
		}
		// Process asset
		for _, processor := range getProcessors() {
			if processor.CanBeUsed(asset.Asset) {
				ok, report := processor.Process(reg.Id, deploymentId, asset, errs)
				if !ok && report == nil {
					// Process error
					report = processError(asset.Asset, processor.GetType(), "Error processing asset", "UNKNOWN_PROCESSING_ERROR", errs)
				}
				reports = append(reports, *report)
			}
		}
	}
	return reports
}

func validateAsset(asset ltiservices.DownloadedAsset, errs *utils.JsonErrors) (bool, *ltiservices.Report) {
	f, err := os.Open(asset.Path)
	if err != nil {
		utils.AddError(errs, "Error opening asset file", err)
		errs.Code = 400
		return false, processError(asset.Asset, "validateAsset", "Error opening asset file", "UNKNOWN_PROCESSING_ERROR", errs)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		utils.AddError(errs, "Error reading asset file", err)
		errs.Code = 400
		return false, processError(asset.Asset, "validateAsset", "Error reading asset file", "UNKNOWN_PROCESSING_ERROR", errs)
	}

	// Base 64 encode the hash
	hash := h.Sum(nil)
	encodedHash := base64.StdEncoding.EncodeToString(hash)
	// Compare the hash with the asset's hash
	if asset.Asset.Checksum != encodedHash {
		msg := fmt.Sprintf("Asset hash does not match: expected %s, got %s", asset.Asset.Checksum, encodedHash)
		errCode := "INVALID_ASSET_HASH"
		return false, processError(asset.Asset, "validateAsset", msg, errCode, errs)
	}
	return true, nil
}

func processError(asset ltiservices.Asset, processorType string, message string, errCode string, errs *utils.JsonErrors) *ltiservices.Report {
	return &ltiservices.Report{
		AssetId:            asset.Id,
		Type:               processorType,
		Timestamp:          time.Now().Format(time.RFC3339),
		Title:              asset.Title,
		IndicationColor:    "#990000",
		IndicationAlt:      "Error",
		Priority:           0,
		ProcessingProgress: "Failed",
		Comment:            &message,
		ErrorCode:          &errCode,
	}
}

func processUnsupportedError(asset ltiservices.Asset, errs *utils.JsonErrors) *ltiservices.Report {
	msg := fmt.Sprintf("Unsupported file type: %s", asset.ContentType)
	errCode := "UNSUPPORTED_ASSET_TYPE"
	return processError(asset, "unsupported", msg, errCode, errs)
}

func getProcessors() []IProcessor {
	if processors == nil {
		processors = []IProcessor{
			textCountProcessor,
			textProcessor,
			imageProcessor,
		}
	}
	return processors
}

func GetProcessorByType(processorType string) IProcessor {
	for _, processor := range getProcessors() {
		if processor.GetType() == processorType {
			return processor
		}
	}
	return nil
}
