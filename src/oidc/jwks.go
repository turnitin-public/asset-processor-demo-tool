package oidc

import (
	"1edtech/ap-demo/datastore"
	"1edtech/ap-demo/utils"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
)

// well known keyset
func PrintJwks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	errs := utils.JsonErrors{Errors: make([]utils.JsonError, 0), Code: 400}
	keys, err := datastore.RegistrationQueries.GetAllKeys()
	if err != nil {
		utils.AddError(&errs, "Unable to find keys", err)
		utils.WriteJsonError(w, r, errs)
		return
	}
	jwks := jwk.NewSet()
	for _, key := range keys {

		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key.PrivateKey))
		if err != nil {
			utils.AddError(&errs, "Failed to parse private key", err)
			utils.WriteJsonError(w, r, errs)
			return
		}
		j, _ := jwk.New(privateKey.PublicKey)
		j.Set("kid", key.Kid)
		jwks.Add(j)
	}

	json.NewEncoder(w).Encode(jwks)

}
