package main

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltimessages"
	"1edtech/ap-demo/ltinotices"
	"1edtech/ap-demo/oidc"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
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
		Handler:           mux,
	}
	// Start server
	fmt.Println("Starting...")
	log.Fatal(srv.ListenAndServe())
}
