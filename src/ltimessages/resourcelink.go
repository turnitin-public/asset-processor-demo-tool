package ltimessages

import (
	"1edtech/ap-demo/utils"
	"net/http"
)

func resourceLinkRequest(w http.ResponseWriter, r *http.Request, claims *LtiMessage) utils.JsonErrors {
	var errs = utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}

	data := struct {
		Title string
	}{
		Title: claims.ResourceLink.Title,
	}

	utils.TemplateLoader("templates/resource.html", data, w, &errs)

	return errs
}
