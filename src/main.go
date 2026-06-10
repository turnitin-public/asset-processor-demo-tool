package main

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltidr"
	"1edtech/ap-demo/ltimessages"
	"1edtech/ap-demo/ltinotices"
	"1edtech/ap-demo/oidc"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("new request: method=%s path=%s remote=%s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	type serviceStatus struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	}
	type healthResponse struct {
		Status   string                   `json:"status"`
		Services map[string]serviceStatus `json:"services"`
	}

	services := map[string]serviceStatus{}
	overall := "ok"

	// Check database
	if err := datastore.Ping(); err != nil {
		services["database"] = serviceStatus{Status: "unavailable", Error: err.Error()}
		overall = "degraded"
	} else {
		services["database"] = serviceStatus{Status: "ok"}
	}

	// Check LLM server
	llmURL := os.Getenv("LLM_SERVER_URL")
	if llmURL == "" {
		services["llm"] = serviceStatus{Status: "unconfigured"}
	} else {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(llmURL + "/health")
		if err != nil {
			services["llm"] = serviceStatus{Status: "unavailable", Error: err.Error()}
			overall = "degraded"
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				services["llm"] = serviceStatus{Status: "ok"}
			} else {
				services["llm"] = serviceStatus{Status: "unavailable", Error: resp.Status}
				overall = "degraded"
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if overall != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(healthResponse{Status: overall, Services: services})
}

func main() {
	mux := http.NewServeMux()
	// Health check
	mux.Handle("/health", http.HandlerFunc(healthHandler))

	// OIDC Login
	mux.Handle("/oidc/login", http.HandlerFunc(oidc.Login))

	// LTI Message Handler
	mux.Handle("/lti/launch", http.HandlerFunc(ltimessages.Handler))

	// LTI Batch Notice Handler
	mux.Handle("/lti/notice", http.HandlerFunc(ltinotices.BatchHandler))

	// JWKS endpoint
	mux.Handle("/.well-known/jwks.json", http.HandlerFunc(oidc.PrintJwks))

	// Deep Linking Response
	mux.Handle("/lti/deeplink/return", http.HandlerFunc(ltimessages.DeepLinkingResponse))

	// LTI Dynamic Registration
	mux.Handle("/lti-dr/initiate", http.HandlerFunc(ltidr.Initiate))
	mux.Handle("/lti-dr/register", http.HandlerFunc(ltidr.Register))
	mux.Handle("/admin/registrations", http.HandlerFunc(ltidr.Registrations))
	mux.Handle("/admin/registrations/delete", http.HandlerFunc(ltidr.DeleteRegistration))

	// http call Test
	mux.Handle("/client", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, _ := http.NewRequest("POST", os.Getenv("LLM_SERVER_URL")+"/completion", bytes.NewBuffer([]byte(`{"prompt":"write me a story about cheese rolling down a hill."}`)))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Failed to make request: "+err.Error())
			return
		}
		w.WriteHeader(200)
		resp.Write(w)
		//fmt.Fprint(w, "called")
	}))

	datastore.DBInit()

	srv := &http.Server{
		Addr:              ":8000",
		ReadTimeout:       0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		ReadHeaderTimeout: 0,
		Handler:           requestLogger(mux),
	}
	// Start server
	fmt.Println("Starting...")
	log.Fatal(srv.ListenAndServe())
}
