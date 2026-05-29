package ltimessages

import (
	"1edtech/ap-demo/utils"
	"net/http"
)

func assetProcessorSettingsRequest(w http.ResponseWriter, r *http.Request, claims *LtiMessage) utils.JsonErrors {
	errs := utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}
	data := struct {
		Title string
	}{
		Title: claims.Activity.Title,
	}

	utils.TemplateLoader("templates/settings.html", data, w, &errs)

	return errs
}
