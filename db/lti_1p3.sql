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

/* Insert dummy data */

INSERT INTO key_set VALUES('d48a53de-021f-46f7-a0a4-7134812c2235');

INSERT INTO a_key VALUES(
    '1e3f0512-2066-4f8a-8916-2d278bf49524',
    'd48a53de-021f-46f7-a0a4-7134812c2235',
    '-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCiHuIw6bGd/UmN
3cSQKBQ2Ueo43gG2mAoHjlZWWkdngs8GNNGDkQyBXx2FQaj57UsAjwn6uoFz9Xol
20gw5ZezEXe3AcoBa1Af3MGVfbTBQAG5eKEVdJdWp1c7u4EFLXROoWqg4MaBrFEb
l0vZubaXz8AGNvLgWRATgjAKlCIKMNf6jcZ1vd1eNHjDWrBrgtVHlsHZKYpINTVf
qRO6ivHD2eJWAuJ4N6dc0giPdqyXaSlyMe0/eHwlPM8uERxOhnhloHrIyxhmUGzG
4ifTHKk4BTzP9IFup1ezXNDLMdmOCJ2iS61Xv6pc+q0dQPqwstaL9L+Bl8RZIafZ
zsvNCGBDAgMBAAECggEAMpj881MccjipCjPaszsvA70RIup3EmvlRXJxE8ZdXrr+
resyMKPGiWIMLpjNiiM7M1NxQ+WNnYlRtBr6LviZHfQnruBKEaNSgH8/k86F6YJ2
h1JUxAN9cDgOC8B7hggnsprCUq+UhMgpEDlqHOvPRxY50ja4GrjxQYVyRPrynMce
f3YZL0PTS5VRKrpbvsg5upWbQT02fq5whMd89aaPe+f8q1tOGsfFfHZkgozD082H
K5ovzIRJEhdm6W7OHEoD9SHBcqcajozXJX6KzYbu7Pf/u2iAwLI7trgjV8aG8gGd
QzrWctCySJkSf7aKSDmNkUE2R/M6+KGkZkyK/6kUEQKBgQDTnvBr5ELXBEDhz1u1
UI+nzLKWszWeVx//W6AUWY/PejS71ObtC7RKEOO9Gks3OORbN8d8Mx87SL3/Boj5
kSiKnxXAwo9anYvk793+hsDzUc2gCwOTGLul+mnQqLDVe2e/bfRQQBCEu3f4JrmX
TR6OBPHB/buH9JWAh9CUBeQBRwKBgQDEHnrG3R0sdQc674vakyC6pi322KYZ5sIQ
wNMh6I8h9NgVNzDRIHFwN7tRYODIsWHMiXL8JBpNuRHJgrN4r2XJI8OKolMbG2e3
cjeZEHXjW7Rv5Q3QzuMNmil4Cb/JXdeFgZz0vUezsfCB8Dud+pKvEo3Y7dnYqvqu
V6nYNDrHJQKBgC8LMTU67CTyfB32w9Nd0mGiHr1jn3LQuXtB+icr9c1QxHJRFPjz
ViP09zutobTn/9PLZZxVnQbH1/zejgq020ddsC9G0Sl6xoOhUz9m43Pz5ntCl4vW
vrhaH7XUGmOK6Hhk0CAa7dEj/7p5mV5qNXWq4beXWV4S4D1Pc+3EFXi3AoGAbUny
72j+veyFV/FvxSEiJwE+MgXvIhX25XEe9xFq2ehglgoIeTGUJY3ZI/NRsGUw89NQ
sXPI+LD+WYYtTz6nARyd9l6Y4001UgQjOXfzyfwrpANH3Km927GiFFOSfbt+w9ZD
yhrEnz20oiRmhJXDMi6rv0xkjppRUeBmNKZ+bsUCgYBrqsX1IFCz4X5A3+5caSjd
6cqbflHQ5PaIslkbzd4A3OtktzudY9mVnPy7Z8+bs6Wh5VRWDamaGr2VZ7RpU68s
06Jl9YGpzLKqYy701ICuQ7OolcNFG1PSMq5m9gtLedTS81lV50YhJ0V0aLLfXD78
SlqmDB4k31sMR6Zhx0icvQ==
-----END PRIVATE KEY-----',
    'RS256'
);

INSERT INTO registration VALUES(
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'https://canvas.instructure.com',
    '138950000000005153',
    'https://sso.canvaslms.com/api/lti/authorize_redirect',
    'https://sso.canvaslms.com/login/oauth2/token',
    'https://sso.canvaslms.com/api/lti/security/jwks',
    null,
    'https://localhost:9002/lti/launch',
    'd48a53de-021f-46f7-a0a4-7134812c2235'
);


INSERT INTO deployment VALUES(
    '361:c171795ac76afc4b9c7ed8dfc2f97f04741f715f',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '35993:1dcbe0e28e0b86ec9ef98134a0ad3e16806e62e5',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '35994:16256b8bef8f64c7a2f6d08614573150cbd55bd2',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '35995:7ee4927cfbe58059715f694e47e83dc15934cf30',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '35996:e29a7133b42614d47e6c1815bb4837645912059c',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '35997:7a9f1324c249c2e39bad4197c5f899823b2c964c',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '35998:0dae1b84dd3e14923caa85d535de984e561cdd51',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '35999:91ebd8b032072cbba5f69606e4da0fff113246b5',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '36000:212d4107262c41d4c543cab15c1e22bae1ff1b78',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '36001:7c8a1ecfcef2dd0fbab3491fafac8d54660a864b',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);

INSERT INTO deployment VALUES(
    '36002:d250a8fa86df716d1db1b3f9b62ab2e103715a1a',
    '18e7ce58-180f-4af0-91a4-e5707265b902',
    'Example Customer'
);
