package utils

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// Error handling code

// uiError will load an html error template to print out an error message for the end user.
func UiError(w http.ResponseWriter, s int, e string) {
	t, err := template.ParseFiles("templates/error.html")
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "Failed to parse template: "+err.Error())
		return
	}
	data := struct {
		Code    int
		Message string
	}{
		Code:    s,
		Message: e,
	}

	w.WriteHeader(s)
	if err := t.Execute(w, data); err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "Failed to parse template: "+err.Error())
		return
	}
}

// A JsonError represents a single error with a message.
// The message should be suitable to display to the end user.
type JsonError struct {
	Message string `json:"message"`
}
type JsonErrors struct {
	Errors []JsonError `json:"errors"`
	Code   int         `json:"code"`
}

// WriteJsonError will return a JsonErrors object formatted as JSON to the end user.
// This erroring mechanism is used for returning errors for API endpoints.
func WriteJsonError(w http.ResponseWriter, r *http.Request, e JsonErrors) {
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("origin"))
	w.WriteHeader(e.Code)
	json.NewEncoder(w).Encode(e)
}

func AddError(e *JsonErrors, m string, f interface{}) {
	e.Errors = append(e.Errors, JsonError{Message: m})
	log.Print(e)
	log.Printf("%+v", f)
}
