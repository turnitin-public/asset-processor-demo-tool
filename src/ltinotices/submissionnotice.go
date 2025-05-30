package ltinotices

import (
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/processors"
	"1edtech/ap-demo/utils"
	"log"
	"net/http"
)

func assetProcessorSubmissionNotice(w http.ResponseWriter, r *http.Request, claims *LtiNotice) utils.JsonErrors {
	// Register for submission notices
	var errs = utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}

	if claims.AssetService == nil {
		utils.AddError(&errs, "Asset service claim not found", nil)
		errs.Code = 400
	}

	if claims.ForUser == nil {
		utils.AddError(&errs, "For user not found", nil)
		errs.Code = 400
	}

	if claims.AssetReport == nil {
		utils.AddError(&errs, "Asset report service claim not found", nil)
		errs.Code = 400
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	// Filter processable assets
	assets, reports := processors.FilterAssets(claims.AssetService.Assets, &errs)

	// Call asset service
	downloadedAssets := ltiservices.FetchAssets(claims.Issuer, claims.Audience, claims.DeploymentId, assets, claims.ForUser.UserId, claims.AssetService.Scope, &errs)

	// Do processing here
	processorReports := processors.ProcessAssets(claims.Issuer, claims.Audience, claims.DeploymentId, downloadedAssets, &errs)
	reports = append(reports, processorReports...)

	if len(reports) == 0 {
		utils.AddError(&errs, "No reports generated", nil)
		errs.Code = 400
		return errs
	}

	log.Printf("reports: %v", reports)

	// Call report service
	_ = ltiservices.SendReports(claims.Issuer, claims.Audience, claims.DeploymentId, claims.AssetReport.ReportUrl, claims.AssetReport.Scope, reports, &errs)

	return errs
}
