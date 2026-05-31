package main

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/ltimessages"
	"1edtech/ap-demo/ltinotices"
	"1edtech/ap-demo/oidc"
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

	fmt.Printf("Tunnel URL: %s\n", os.Getenv("TUNNEL_URL"))

	datastore.DBInit()

	srv := &http.Server{
		Addr:              ":8000",
		ReadTimeout:       0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		ReadHeaderTimeout: 0,
		Handler:           mux,
	}
	// Start server with HTTPS
	// Replace with your certificate files
	// For development, you can generate self-signed certs with:
	// openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
	fmt.Println("Starting HTTPS server...")
	log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))
}
