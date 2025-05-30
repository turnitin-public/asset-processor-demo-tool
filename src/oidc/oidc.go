package oidc

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/utils"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// Login
func Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Print(err)
	}
	if !r.Form.Has("reg_id") {
		utils.UiError(w, 400, "No reg id")
		return
	}

	if !r.Form.Has("iss") {
		utils.UiError(w, 400, "No iss")
		return
	}
	// Find reg
	reg, err := datastore.RegistrationQueries.GetRegistration(r.Form.Get("iss"), r.Form.Get("reg_id"))
	if err != nil {
		utils.UiError(w, 400, "Failed to find registration")
		return
	}

	// Set state
	auth_params := map[string]string{
		"client_id":     reg.ClientId,        // Registered client id.
		"redirect_uri":  reg.ToolRedirectUri, // URL to return to after login.
		"scope":         "openid",            // Requested scope.
		"response_type": "id_token",          // Response type.
		"response_mode": "form_post",         // Response mode.
		"prompt":        "none",              // Prompt.
		"state":         uuid.New().String(), // State.
		"nonce":         uuid.New().String(), // Nonce.
	}
	if r.Form.Has("login_hint") {
		auth_params["login_hint"] = r.Form.Get("login_hint") // Login hint to identify platform session.
	}
	if r.Form.Has("lti_message_hint") {
		auth_params["lti_message_hint"] = r.Form.Get("lti_message_hint") // Login hint to identify platform session.
	}

	t, err := template.ParseFiles("templates/autopost.html")
	if err != nil {
		utils.UiError(w, 500, "Failed to parse template")
		return
	}
	data := struct {
		Action string
		Params map[string]string
	}{
		Action: reg.PlatformLoginAuthEndpoint,
		Params: auth_params,
	}
	t.Execute(w, data)
}
