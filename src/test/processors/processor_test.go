package processors_test

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/processors"
	"1edtech/ap-demo/utils"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"html/template"
	"os"
	"testing"
)

type mockRegistrationQueries struct{}

func (mockRegistrationQueries) GetRegistration(i string, r string) (*datastore.ToolRegistration, error) {
	return nil, nil
}

func (mockRegistrationQueries) GetRegistrationByClient(i string, c string) (*datastore.ToolRegistration, error) {
	return &datastore.ToolRegistration{Id: "registration-1"}, nil
}

func (mockRegistrationQueries) GetPrivateKeyAndRegForClient(i string, c string, errs *utils.JsonErrors) (*rsa.PrivateKey, *datastore.RegistrationWithKey, bool) {
	return nil, nil, false
}

func (mockRegistrationQueries) GetAllKeys() ([]datastore.Key, error) {
	return nil, nil
}

func (mockRegistrationQueries) GetDefaultKeySetID() (string, error) {
	return "", nil
}

func (mockRegistrationQueries) CreateRegistrationAndDeployment(reg datastore.DynamicRegistrationInsert) error {
	return nil
}

func (mockRegistrationQueries) ListRegistrations() ([]datastore.RegistrationSummary, error) {
	return nil, nil
}

func (mockRegistrationQueries) DeleteRegistration(registrationID string) error {
	return nil
}

type mockProcessor struct{}

func (mockProcessor) GetName() string { return "mockProcessor" }
func (mockProcessor) GetType() string { return "mock" }
func (mockProcessor) CanBeUsed(asset ltiservices.Asset) bool {
	return asset.ContentType == "mock/type"
}
func (mockProcessor) Process(registrationId, deploymentId string, asset ltiservices.DownloadedAsset, errs *utils.JsonErrors) (bool, *ltiservices.Report) {
	return true, &ltiservices.Report{AssetId: asset.Asset.Id, Type: "mock", ProcessingProgress: "Processed"}
}
func (mockProcessor) GetFileHtml(internalAssetId string) (template.HTML, error) {
	return template.HTML("<div>mock</div>"), nil
}

func TestProcessAssets(t *testing.T) {
	t.Cleanup(func() {
		datastore.RegistrationQueries = nil
		processors.SetProcessorsForTest(nil)
	})

	datastore.RegistrationQueries = mockRegistrationQueries{}
	processors.SetProcessorsForTest([]processors.IProcessor{mockProcessor{}})

	content := []byte("hello world")
	checksum := sha256.Sum256(content)
	encodedChecksum := base64.StdEncoding.EncodeToString(checksum[:])
	assetFile := t.TempDir() + string(os.PathSeparator) + "example.txt"
	if err := os.WriteFile(assetFile, content, 0o600); err != nil {
		t.Fatalf("failed to write temp asset: %v", err)
	}

	assets := []ltiservices.DownloadedAsset{{
		Asset: ltiservices.Asset{
			Id:          "asset-1",
			ContentType: "mock/type",
			Checksum:    encodedChecksum,
			Title:       "Example",
		},
		Path: assetFile,
	}}

	testErrs := &utils.JsonErrors{}
	reports := processors.ProcessAssets("issuer-1", "client-1", "deployment-1", assets, testErrs)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].Type != "mock" {
		t.Fatalf("unexpected report type: %s", reports[0].Type)
	}
}