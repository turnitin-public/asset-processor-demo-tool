package datastore_test

import (
	"testing"

	"1edtech/ap-demo/datastore"

	"github.com/golang-jwt/jwt"
)

func TestGenerateSigningKeyMaterial(t *testing.T) {
	material, err := datastore.GenerateSigningKeyMaterial()
	if err != nil {
		t.Fatalf("expected key generation to succeed: %v", err)
	}
	if material.KeySetID == "" {
		t.Fatal("expected key set id to be generated")
	}
	if material.KeyID == "" {
		t.Fatal("expected key id to be generated")
	}
	if material.Alg != "RS256" {
		t.Fatalf("expected RS256 algorithm, got %s", material.Alg)
	}
	if _, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(material.PrivateKey)); err != nil {
		t.Fatalf("expected generated private key to be parseable: %v", err)
	}
}