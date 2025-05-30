package processors

import (
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProcessAssets(t *testing.T) {
	// Test the ProcessAssets function
	testInit(t)
	// Create a list of assets
	assets := []ltiservices.DownloadedAsset{
		{
			Asset: ltiservices.Asset{
				ContentType: "image",
			},
		},
		{
			Asset: ltiservices.Asset{
				ContentType: "text",
				Id:          "123",
				Url:         "http://example.com",
				Title:       "Example",
				FileName:    "example.txt",
				Timestamp:   "2025-01-01T00:00:00Z",
			},
			UserId: "123",
			Path:   "/tmp/123",
		},
		{
			Asset: ltiservices.Asset{
				ContentType: "unknown",
			},
		},
	}
	// Create err object
	testErrs := &utils.JsonErrors{}
	// Create a list of reports
	reports := ProcessAssets("", "", "", assets, testErrs)
	// Check the length of the reports
	if len(reports) != 1 {
		t.Errorf("Expected 1 reports, got %d", len(reports))
	}
}

func testInit(t *testing.T) {
	httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}
