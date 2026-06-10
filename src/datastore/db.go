package datastore

import (
	"1edtech/ap-demo/utils"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"crypto/x509"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var db *sql.DB

// Ping checks the database connection and returns an error if it is unavailable.
func Ping() error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.Ping()
}

type IRegistrationQueries interface {
	GetRegistration(i string, r string) (*ToolRegistration, error)
	GetRegistrationByClient(i string, c string) (*ToolRegistration, error)
	GetPrivateKeyAndRegForClient(i string, c string, errs *utils.JsonErrors) (*rsa.PrivateKey, *RegistrationWithKey, bool)
	GetAllKeys() ([]Key, error)
	GetDefaultKeySetID() (string, error)
	CreateRegistrationAndDeployment(reg DynamicRegistrationInsert) error
	ListRegistrations() ([]RegistrationSummary, error)
	DeleteRegistration(registrationID string) error
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

type DynamicRegistrationInsert struct {
	RegistrationID              string
	DeploymentID                string
	Issuer                      string
	ClientID                    string
	PlatformLoginAuthEndpoint   string
	PlatformServiceAuthEndpoint string
	PlatformJwksEndpoint        string
	PlatformAuthProvider        *string
	ToolRedirectURI             string
	KeySetID                    string
	CustomerID                  string
}

type RegistrationSummary struct {
	RegistrationID string
	Issuer         string
	ClientID       string
	DeploymentID   string
	CustomerID     string
}

type generatedKeyMaterial struct {
	KeySetID   string
	KeyID      string
	PrivateKey string
	Alg        string
}

type GeneratedKeyMaterial = generatedKeyMaterial

func DBInit() {
	fmt.Println("Connecting to db...")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		port = 5432
	}
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_HOST"), port, os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	// Open the database handle once, then wait for Postgres to accept connections.
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := waitForDatabase(60 * time.Second); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if err := ensureSigningKeyMaterial(); err != nil {
		log.Fatalf("Failed to initialize signing keys: %v", err)
	}
	fmt.Println("Connected to db...")
	RegistrationQueries = defaultRegistrationQueries{}
	AssetReportQueries = defaultAssetReportQueries{}
}

func waitForDatabase(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if err := db.Ping(); err == nil {
			return nil
		} else if time.Now().After(deadline) {
			return err
		}

		time.Sleep(2 * time.Second)
	}
}

func ensureSigningKeyMaterial() error {
	row := db.QueryRow(`SELECT COUNT(*) FROM a_key`)
	var keyCount int
	if err := row.Scan(&keyCount); err != nil {
		return err
	}
	if keyCount > 0 {
		return nil
	}

	material, err := generateSigningKeyMaterial()
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`INSERT INTO key_set (id) VALUES ($1)`, material.KeySetID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO a_key (
		id,
		key_set_id,
		private_key,
		alg
	) VALUES ($1, $2, $3, $4)`, material.KeyID, material.KeySetID, material.PrivateKey, material.Alg)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Print("Generated default signing key material")
	return nil
}

func generateSigningKeyMaterial() (*generatedKeyMaterial, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBytes})
	if len(privateKeyPEM) == 0 {
		return nil, fmt.Errorf("failed to encode private key")
	}

	return &generatedKeyMaterial{
		KeySetID:   uuid.New().String(),
		KeyID:      uuid.New().String(),
		PrivateKey: string(privateKeyPEM),
		Alg:        "RS256",
	}, nil
}

func GenerateSigningKeyMaterial() (*GeneratedKeyMaterial, error) {
	return generateSigningKeyMaterial()
}

func (defaultRegistrationQueries) GetRegistration(i string, r string) (*ToolRegistration, error) {
	log.Printf("GetRegistration called with issuer=%s, id=%s", i, r)
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

func (defaultRegistrationQueries) GetDefaultKeySetID() (string, error) {
	row := db.QueryRow(`SELECT key_set_id
	FROM a_key
	ORDER BY created DESC
	LIMIT 1`)
	var keySetID string
	if err := row.Scan(&keySetID); err != nil {
		return "", err
	}
	return keySetID, nil
}

func (defaultRegistrationQueries) CreateRegistrationAndDeployment(reg DynamicRegistrationInsert) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`INSERT INTO registration (
		id,
		issuer,
		client_id,
		platform_login_auth_endpoint,
		platform_service_auth_endpoint,
		platform_jwks_endpoint,
		platform_auth_provider,
		tool_redirect_uri,
		key_set_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		reg.RegistrationID,
		reg.Issuer,
		reg.ClientID,
		reg.PlatformLoginAuthEndpoint,
		reg.PlatformServiceAuthEndpoint,
		reg.PlatformJwksEndpoint,
		reg.PlatformAuthProvider,
		reg.ToolRedirectURI,
		reg.KeySetID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO deployment (
		deployment_id,
		registration_id,
		customer_id
	) VALUES ($1, $2, $3)`, reg.DeploymentID, reg.RegistrationID, reg.CustomerID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (defaultRegistrationQueries) ListRegistrations() ([]RegistrationSummary, error) {
	rows, err := db.Query(`SELECT
		r.id,
		r.issuer,
		r.client_id,
		d.deployment_id,
		d.customer_id
	FROM registration r
	LEFT JOIN deployment d ON d.registration_id = r.id
	ORDER BY r.issuer ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	registrations := make([]RegistrationSummary, 0)
	for rows.Next() {
		var summary RegistrationSummary
		if err := rows.Scan(&summary.RegistrationID, &summary.Issuer, &summary.ClientID, &summary.DeploymentID, &summary.CustomerID); err != nil {
			return nil, err
		}
		registrations = append(registrations, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return registrations, nil
}

func (defaultRegistrationQueries) DeleteRegistration(registrationID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`DELETE FROM deployment WHERE registration_id = $1`, registrationID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM registration WHERE id = $1`, registrationID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
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
