package ltimessages

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/processors"
	"1edtech/ap-demo/utils"
	"html/template"
	"net/http"
)

func reportReviewRequest(w http.ResponseWriter, r *http.Request, claims *LtiMessage) utils.JsonErrors {
	// Register for submission notices
	var errs = utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}
	// Get report for current asset
	id, report, ok := datastore.AssetReportQueries.GetAssetReport(claims.Issuer, claims.Audience, claims.DeploymentId, claims.Asset.Id, claims.ReportType)
	if !ok {
		utils.AddError(&errs, "Report not found", ok)
		errs.Code = 404
		return errs
	}
	fileEmbed, err := processors.GetProcessorByType(claims.ReportType).GetFileHtml(id)
	if err != nil {
		fileEmbed = "Unable to load report"
	}
	// Load Template
	t, err := template.ParseFiles("templates/report.html")
	if err != nil {
		utils.AddError(&errs, "Unable to load template", err)
		errs.Code = 500
		return errs
	}
	data := struct {
		Report    string
		AssetId   string
		FileEmbed template.HTML
	}{
		Report:    report,
		AssetId:   claims.Asset.Id,
		FileEmbed: fileEmbed,
	}
	err = t.Execute(w, data)
	if err != nil {
		utils.AddError(&errs, "Unable to render template", err)
		errs.Code = 500
	}

	return errs
}
