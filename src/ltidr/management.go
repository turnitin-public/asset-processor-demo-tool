package ltidr

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/utils"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

type registrationListItem struct {
	RegistrationID string
	Issuer         string
	ClientID       string
	DeploymentID   string
	CustomerID     string
	LoginURL       string
}

func Registrations(w http.ResponseWriter, r *http.Request) {
	registrations, err := datastore.RegistrationQueries.ListRegistrations()
	if err != nil {
		utils.UiError(w, 500, "Unable to load registrations")
		return
	}

	toolBaseURL := resolveToolBaseURL(r)
	items := make([]registrationListItem, 0, len(registrations))
	for _, registration := range registrations {
		items = append(items, registrationListItem{
			RegistrationID: registration.RegistrationID,
			Issuer:         registration.Issuer,
			ClientID:       registration.ClientID,
			DeploymentID:   registration.DeploymentID,
			CustomerID:     registration.CustomerID,
			LoginURL:       buildLoginURL(toolBaseURL, registration.Issuer, registration.RegistrationID),
		})
	}

	if isJSONRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
		return
	}

	t, err := template.ParseFiles("templates/registrations.html")
	if err != nil {
		utils.UiError(w, 500, "Unable to load registrations template")
		return
	}

	data := struct {
		Registrations []registrationListItem
	}{
		Registrations: items,
	}
	if err := t.Execute(w, data); err != nil {
		utils.UiError(w, 500, "Unable to render registrations template")
	}
}

func DeleteRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	registrationID, ok := registrationIDFromRequest(r)
	if !ok {
		if isJSONRequest(r) {
			utils.WriteJsonError(w, r, utils.JsonErrors{Errors: []utils.JsonError{{Message: "registration_id is required"}}, Code: 400})
			return
		}
		utils.UiError(w, 400, "registration_id is required")
		return
	}

	if err := datastore.RegistrationQueries.DeleteRegistration(registrationID); err != nil {
		if isJSONRequest(r) {
			utils.WriteJsonError(w, r, utils.JsonErrors{Errors: []utils.JsonError{{Message: "unable to delete registration"}}, Code: 500})
			return
		}
		utils.UiError(w, 500, "Unable to delete registration")
		return
	}

	if isJSONRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "registration_id": registrationID})
		return
	}

	http.Redirect(w, r, "/admin/registrations", http.StatusSeeOther)
}

func registrationIDFromRequest(r *http.Request) (string, bool) {
	if r.Method == http.MethodDelete || strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		var payload struct {
			RegistrationID string `json:"registration_id"`
		}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
				payload.RegistrationID = strings.TrimSpace(payload.RegistrationID)
				if payload.RegistrationID != "" {
					return payload.RegistrationID, true
				}
			}
		}
	}

	if err := r.ParseForm(); err == nil {
		registrationID := strings.TrimSpace(r.FormValue("registration_id"))
		if registrationID != "" {
			return registrationID, true
		}
	}

	registrationID := strings.TrimSpace(r.URL.Query().Get("registration_id"))
	if registrationID == "" {
		return "", false
	}
	return registrationID, true
}
