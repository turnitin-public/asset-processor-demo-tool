package ltimessages

import (
	"1edtech/ap-demo/ltiservices"
	"1edtech/ap-demo/utils"
	"html/template"
	"net/http"
)

func eulaRequest(w http.ResponseWriter, r *http.Request, claims *LtiMessage) utils.JsonErrors {
	var errs = utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 200}

	// Accept eula (this should happen after the user accepts, but we are doing it automatically here for demo purposes)
	ok := ltiservices.AcceptEula(claims.Issuer, claims.Audience, claims.DeploymentId, claims.Eula.Url, claims.Eula.Scope, claims.Subject, &errs)
	if !ok {
		utils.AddError(&errs, "Unable to accept EULA", nil)
		errs.Code = 500
		return errs
	}

	// Load Template
	t, err := template.ParseFiles("templates/eula.html")
	if err != nil {
		utils.AddError(&errs, "Unable to load template", err)
		errs.Code = 500
		return errs
	}
	data := struct{}{}
	err = t.Execute(w, data)
	if err != nil {
		utils.AddError(&errs, "Unable to render template", err)
		errs.Code = 500
	}

	return errs
}
