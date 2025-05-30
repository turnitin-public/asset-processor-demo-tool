package processors

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/utils"
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type textProcessorType struct{}

var textProcessor IProcessor = textProcessorType{}

func (textProcessorType) GetName() string {
	return "textProcessor"
}

func (textProcessorType) GetType() string {
	return "text"
}

func (textProcessorType) CanBeUsed(asset ltiservices.Asset) bool {
	// Check if the asset is a text file
	return asset.ContentType == "text/plain"
}

func (textProcessorType) Process(registrationId string, deploymentId string, asset ltiservices.DownloadedAsset, errs *utils.JsonErrors) (bool, *ltiservices.Report) {
	// Process text
	// Get asset file as string
	assetFile, err := os.ReadFile(asset.Path)
	if err != nil {
		utils.AddError(errs, "Error reading asset file", err)
		errs.Code = 401
		return false, nil
	}
	prompt := "You will be given a block of text, the text block will begin with <TEXT_START> and end with <TEXT_END>. These tags are not included as part of the overall text block. Summarize the text in the following text block, giving a short, single sentence to describe what the text is about in an easily digestible manner. You will not ask any follow up questions or make any statements after the summary.\n\n<TEXT_START>" + string(assetFile) + "<TEXT_END>\nHere is a short summary:\n"

	llmRequest := struct {
		Prompt   string `json:"prompt"`
		NPredict int    `json:"n_predict"`
	}{
		Prompt:   prompt,
		NPredict: 500,
	}

	// Call the local llm service
	b, err := json.Marshal(llmRequest)
	if err != nil {
		utils.AddError(errs, "Error marshalling llm request", err)
		errs.Code = 401
		return false, nil
	}

	req, err := http.NewRequest("POST", os.Getenv("LLM_SERVER_URL")+"/completion", bytes.NewBuffer(b))
	if err != nil {
		utils.AddError(errs, "Error creating request", err)
		errs.Code = 401
		return false, nil
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := utils.HttpClient().Do(req)
	if err != nil {
		utils.AddError(errs, "Error calling llm service", err)
		errs.Code = 401
		return false, nil
	}
	if resp.StatusCode >= 400 {
		utils.AddError(errs, "llm service returned a non-success response", resp)
		errBody, _ := io.ReadAll(resp.Body)
		utils.AddError(errs, "err body", string(errBody))
		errs.Code = resp.StatusCode
		return false, nil
	}

	// Save report to DB
	var llmResponse datastore.LlmResponse
	err = json.NewDecoder(resp.Body).Decode(&llmResponse)
	if err != nil {
		utils.AddError(errs, "Error decoding llm response", err)
		errs.Code = 401
		return false, nil
	}
	ok := datastore.AssetReportQueries.SaveAssetReport(asset.Id, registrationId, deploymentId, asset.Asset.Id, "text", llmResponse.Content)
	if !ok {
		utils.AddError(errs, "Error saving asset report", nil)
		errs.Code = 500
		return false, nil
	}

	log.Printf("llm response: %v", llmResponse)

	return true, &ltiservices.Report{
		AssetId:            asset.Asset.Id,
		Type:               "text",
		Timestamp:          time.Now().Format(time.RFC3339),
		Title:              asset.Asset.Title,
		Result:             "Processed",
		IndicationColor:    "#009900",
		IndicationAlt:      "Good",
		Priority:           0,
		ProcessingProgress: "Processed",
		Comment:            &llmResponse.Content,
		ErrorCode:          nil,
	}
}

func (textProcessorType) GetFileHtml(assetId string) (template.HTML, error) {
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
