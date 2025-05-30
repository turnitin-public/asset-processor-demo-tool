package ltimessages

import (
	"1edtech/ap-demo/utils"
	"html/template"
	"net/http"
)

func resourceLinkRequest(w http.ResponseWriter, r *http.Request, claims *LtiMessage) utils.JsonErrors {
	var errs = utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}

	// Load Template
	t, err := template.ParseFiles("templates/resource.html")
	if err != nil {
		utils.AddError(&errs, "Unable to load template", err)
		errs.Code = 500
		return errs
	}
	data := struct {
		Title string
	}{
		Title: claims.ResourceLink.Title,
	}
	err = t.Execute(w, data)
	if err != nil {
		utils.AddError(&errs, "Unable to render template", err)
		errs.Code = 500
	}

	return errs
}
