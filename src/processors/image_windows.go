//go:build windows
// +build windows

package processors

import (
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/utils"
	"html/template"
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
	return asset.ContentType == "image/jpeg"
}

func (imageProcessorType) Process(registrationId string, deploymentId string, asset ltiservices.DownloadedAsset, errs *utils.JsonErrors) (bool, *ltiservices.Report) {
	return false, nil
}

func (imageProcessorType) GetFileHtml(assetId string) (template.HTML, error) {
	return template.HTML(""), nil
}
