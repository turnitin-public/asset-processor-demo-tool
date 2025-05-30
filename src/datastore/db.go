package datastore

import (
	"1edtech/ap-demo/utils"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	_ "github.com/lib/pq"
)

var db *sql.DB

type IRegistrationQueries interface {
	GetRegistration(i string, r string) (*ToolRegistration, error)
	GetRegistrationByClient(i string, c string) (*ToolRegistration, error)
	GetPrivateKeyAndRegForClient(i string, c string, errs *utils.JsonErrors) (*rsa.PrivateKey, *RegistrationWithKey, bool)
	GetAllKeys() ([]Key, error)
}

var RegistrationQueries IRegistrationQueries

type defaultRegistrationQueries struct{}

type IAssetReportQueries interface {
	SaveAssetReport(id string, registrationId string, deploymentId string, assetId string, assetType string, content string) bool
	GetAssetReport(issuer string, clientId string, deploymentId string, assetId string, assetType string) (string, string, bool)
}

var AssetReportQueries IAssetReportQueries

type defaultAssetReportQueries struct{}

type LlmResponse struct {
	Content string `json:"content"`
}

type ToolRegistration struct {
	Id                          string
	Issuer                      string
	ClientId                    string
	PlatformJwksEndpoint        string
	PlatformLoginAuthEndpoint   string
	ToolRedirectUri             string
	PlatformAuthProvider        *string
	PlatformServiceAuthEndpoint string
}

type Key struct {
	Kid        string
	PrivateKey string
	Alg        string
}

type RegistrationWithKey struct {
	ToolRegistration
	Key
}

func DBInit() {
	fmt.Println("Connecting to db...")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		port = 5432
	}
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_HOST"), port, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	// open database
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		// Sleep for 10 seconds and try again
		log.Printf("Failed to connect to database: %v", err)
		time.Sleep(10 * time.Second)
		DBInit() // Retry initialization
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping to database")
	}
	fmt.Println("Connected to db...")
	RegistrationQueries = defaultRegistrationQueries{}
	AssetReportQueries = defaultAssetReportQueries{}
}

func (defaultRegistrationQueries) GetRegistration(i string, r string) (*ToolRegistration, error) {
	row := db.QueryRow(`SELECT
		id,
		issuer,
		client_id,
		platform_login_auth_endpoint,
		tool_redirect_uri,
		platform_service_auth_endpoint,
		platform_jwks_endpoint,
		platform_auth_provider
	FROM registration
	WHERE issuer = $1 AND id = $2`, i, r)
	var reg ToolRegistration
	if err := row.Scan(
		&reg.Id,
		&reg.Issuer,
		&reg.ClientId,
		&reg.PlatformLoginAuthEndpoint,
		&reg.ToolRedirectUri,
		&reg.PlatformServiceAuthEndpoint,
		&reg.PlatformJwksEndpoint,
		&reg.PlatformAuthProvider); err != nil {
		return nil, err
	}
	return &reg, nil
}

func (defaultRegistrationQueries) GetRegistrationByClient(i string, c string) (*ToolRegistration, error) {
	row := db.QueryRow(`SELECT
		id,
		issuer,
		client_id,
		platform_login_auth_endpoint,
		tool_redirect_uri,
		platform_service_auth_endpoint,
		platform_jwks_endpoint,
		platform_auth_provider
	FROM registration
	WHERE issuer = $1 AND client_id = $2`, i, c)
	var reg ToolRegistration
	if err := row.Scan(
		&reg.Id,
		&reg.Issuer,
		&reg.ClientId,
		&reg.PlatformLoginAuthEndpoint,
		&reg.ToolRedirectUri,
		&reg.PlatformServiceAuthEndpoint,
		&reg.PlatformJwksEndpoint,
		&reg.PlatformAuthProvider); err != nil {
		return nil, err
	}
	return &reg, nil
}

func (defaultRegistrationQueries) GetPrivateKeyAndRegForClient(i string, c string, errs *utils.JsonErrors) (*rsa.PrivateKey, *RegistrationWithKey, bool) {
	// Find registration
	row := db.QueryRow(`SELECT
		issuer,
		client_id,
		platform_login_auth_endpoint,
		tool_redirect_uri,
		platform_service_auth_endpoint,
		platform_auth_provider,
		k.id,
		k.private_key,
		k.alg
	FROM registration r
	JOIN key_set ks on r.key_set_id = ks.id
	JOIN a_key k on ks.id = k.key_set_id
	WHERE issuer = $1 AND client_id = $2
	ORDER BY k.created DESC
	LIMIT 1`, i, c)
	var registration RegistrationWithKey
	if err := row.Scan(&registration.Issuer,
		&registration.ClientId,
		&registration.PlatformLoginAuthEndpoint,
		&registration.ToolRedirectUri,
		&registration.PlatformServiceAuthEndpoint,
		&registration.PlatformAuthProvider,
		&registration.Kid,
		&registration.PrivateKey,
		&registration.Alg); err != nil {
		utils.AddError(errs, "Unable to find registration key", err)
		return nil, nil, false
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(registration.PrivateKey))
	if err != nil {
		utils.AddError(errs, "Unable to parse private key", err)
		errs.Code = 500
		return nil, nil, false
	}
	return privateKey, &registration, true
}

func (defaultRegistrationQueries) GetAllKeys() ([]Key, error) {
	rows, err := db.Query(`SELECT
		k.id,
		k.private_key,
		k.alg
	FROM a_key k
	ORDER BY k.created DESC`)
	if err != nil {
		return nil, err
	}
	var keys []Key
	for rows.Next() {
		var key Key
		if err := rows.Scan(&key.Kid, &key.PrivateKey, &key.Alg); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (defaultAssetReportQueries) SaveAssetReport(id string, registrationId string, deploymentId string, assetId string, assetType string, content string) bool {
	_, err := db.Exec(`INSERT INTO asset_report (
		id,
		registration_id,
		deployment_id,
		asset_id,
		asset_type,
		content) VALUES ($1, $2, $3, $4, $5, $6)`,
		id, registrationId, deploymentId, assetId, assetType, content)
	if err != nil {
		log.Printf("Failed to save asset report: %v", err)
		return false
	}
	return true
}

func (defaultAssetReportQueries) GetAssetReport(issuer string, clientId string, deploymentId string, assetId string, assetType string) (string, string, bool) {
	row := db.QueryRow(`SELECT ar.id, content
	FROM asset_report ar
	JOIN registration r ON r.id = ar.registration_id
	WHERE r.issuer = $1
		AND r.client_id = $2
		AND ar.deployment_id = $3
		AND ar.asset_id = $4
		AND ar.asset_type = $5
	ORDER BY created_at DESC limit 1`, issuer, clientId, deploymentId, assetId, assetType)
	var id string
	var content string
	if err := row.Scan(&id, &content); err != nil {
		log.Printf("Failed to get asset report: %v", err)
		return "", "", false
	}
	return id, content, true
}
