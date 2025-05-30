package ltimessages

import (
	"1edtech/ap-demo/utils"
	"html/template"
	"net/http"
)

func assetProcessorSettingsRequest(w http.ResponseWriter, r *http.Request, claims *LtiMessage) utils.JsonErrors {
	errs := utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}
	// Load Template
	t, err := template.ParseFiles("templates/settings.html")
	if err != nil {
		utils.AddError(&errs, "Unable to load template", err)
		errs.Code = 500
		return errs
	}
	data := struct {
		Title string
	}{
		Title: claims.Activity.Title,
	}
	err = t.Execute(w, data)
	if err != nil {
		utils.AddError(&errs, "Unable to render template", err)
		errs.Code = 500
	}

	return errs
}
