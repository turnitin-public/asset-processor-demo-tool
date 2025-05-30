CREATE TABLE key_set (
  id UUID NOT NULL,

  CONSTRAINT key_set_id PRIMARY KEY (id)
);

CREATE TABLE a_key
(
    id          UUID NOT NULL,
    key_set_id  UUID NOT NULL REFERENCES key_set(id),
    private_key TEXT NOT NULL,
    alg         TEXT NOT NULL,
    created     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_ea_key_id PRIMARY KEY (id)
);

CREATE TABLE registration (
    id                             UUID NOT NULL,
    issuer                         TEXT NOT NULL,
    client_id                      TEXT NOT NULL,
    platform_login_auth_endpoint   TEXT NOT NULL,
    platform_service_auth_endpoint TEXT NOT NULL,
    platform_jwks_endpoint         TEXT NOT NULL,
    platform_auth_provider         TEXT,
    tool_redirect_uri              TEXT NOT NULL,
    key_set_id                     UUID NOT NULL REFERENCES key_set(id),

    CONSTRAINT pk_registration_id PRIMARY KEY (id),
    UNIQUE (issuer, client_id)
);

CREATE TABLE deployment (
  deployment_id TEXT NOT NULL,
  registration_id UUID NOT NULL REFERENCES registration(id),
  customer_id TEXT NOT NULL,

  CONSTRAINT pk_deployment_id PRIMARY KEY (registration_id, deployment_id)
);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE asset_report (
    id UUID NOT NULL DEFAULT uuid_generate_v4(),
    registration_id UUID NOT NULL REFERENCES registration(id),
    deployment_id TEXT NOT NULL,
    asset_id TEXT NOT NULL,
    asset_type TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT pk_asset_report_id PRIMARY KEY (id, asset_type),
    UNIQUE (registration_id, deployment_id, asset_id, created_at)
);
