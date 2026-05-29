## Warning: This code is provided entirely for demonstration purposes and comes with zero warranty. Any use of this software is done so at the user's own risk

# Asset Processor Demo Tool

This is a demonstration LTI (Learning Tools Interoperability) asset processor tool that handles processing of submitted assets (files) and generates reports based on their content.

## Prerequisites

- **Docker** (version 20.10 or higher)
- **Docker Compose** (version 2.0 or higher)
- **Ngrok** (optional, for external access)

### Ngrok Setup (Optional)
For the asset processor tool to function correctly with external LMS platforms, it needs to be externally accessible. Ngrok is used to proxy incoming requests. For ngrok to allow for the service to be exposed on a consistent URL, a paid account is required.

1. Sign up for a [Ngrok paid account](https://ngrok.com/pricing)
2. Get your authtoken from the [Ngrok dashboard](https://dashboard.ngrok.com/get-started/your-authtoken)
3. Create a `.env` file in the project root directory with the following content:

```env
NGROK_AUTH=your_ngrok_authtoken_here
NGROK_DOMAIN=your_custom_domain (optional)
```

### Database Registration Configuration

To add your registration and deployment configuration, you need to insert the relevant rows into the PostgreSQL database. Create a file called `zdata.sql` inside the `db` folder.

**Important**: The SQL queries in `zdata.sql` should match your LMS platform's configuration. Below is an example configuration:

```sql
-- Key Set
INSERT INTO key_set VALUES('d48a53de-021f-46f7-a0a4-7134812c2235');

-- Private Key for signing JWT tokens
INSERT INTO a_key VALUES(
    '1e3f0512-2066-4f8a-8916-2d278bf49524',
    'd48a53de-021f-46f7-a0a4-7134812c2235',
    '-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQD05ZgAT0Ysstqh
IJdzNcpl5N0R0jgDF6aY/4PZ7lWr/wHGxLvY/6Ta6QxEfAzHIGjrrqUwxmvdr8Rn
mG4FD0Wx7Qexe2HO9cLtVD+keLBOOz8C++dXvrugUZ9G4Ea2NE60nZgKDgPvNIJS
ki8Aq8oKHmvlZKFTCTQU+PlfMxa+tZ+9TcDfgjtpisBf00USXCuoSbAfqa+cHEYq
7lcU5nKo+HfO0kJTytGVjxTYvgw9C1hxnhMzof3+tmVlLVJyr15FTvxtBr8C7uMU
zbEP8fXCO+Eg2pL8vXH5834ia1OXG2SEUvEvxGwrASjkmU6/oJVXtvPcc6Y4cD/B
O98wXUuvAgMBAAECggEAVvi0xyNgJh6skz16W8GWLCKfuiUAnGSJV1ujDUHdrhQF
ovwaREHh35aVMzsk5JDngg/Hfa9x/kxeQXY6WFSoqTwnF8pcHX5dKCDb60KrRlU3
Fw20Bo7nnlNub+LVaf7glrdDGAsLLaflwoJE7AWWXoqYQeK/gjhKBUq4cS05Hie7
fg9lKOSFB7WXk76j/C9K2Ab7ah/6NbzmrV6mCruX1gPk19tG0Yw+10e5OIlLtxKT
55NDAiSq1/getURpE9SGi3ZPZcJzE0w28AjS2d8pfesRYROb5c8IaZygszPFEOcZ
EG57rXZtul5aCSUz87DXYcfJ8pJG4bzJdyWtYSL3TQKBgQD/eparwlaVZ9BRkZ5O
uD4Hho+QmcEdxeNrP0+UOW9wwsSV8rAZkQNsRg8Tx2JgcxO2Is+8y/5pmDpXJ+0W
S3xAcmIlbh4Xzvh7Eg7z+blB3SJnRqmleJSqLsG5Wlx4QptzH4UVUecmrnwIr7uE
XEpdj6rDi+L1MtF7ns8bgwtgbQKBgQD1ZXqwbKjTRXLhUGbNb+M1Eotr3JEKR+AN
6j9HRaH/NAe7sg9xNXHHjJ28Ihk/9cjmGXggCuL5N5cOflTnc5DTQr/uiebB8dlP
mGhlfinWvrCVryEpKPzuCq5eh5Y3CgidB6VuI72oYdGVAEsSWemr/0tl/f31Ksds
HUlkCS5jCwKBgHgpx76HzMO/LXuAO26ZOAvAHbyMpQmE7z+daqe0EBeAdIh2up97
1plRpnvOFxZ4afgMDZumc0ZlZGNkEx6eaJXDdyhVz++w2KzCRKg6eAljom/jC54Z
xgr5rQKqXr3tzkHqvGTXvhoyjYJkbZWG9y9kiJQrMpfTzDYR7yXokCxNAoGAXNmJ
w4lJk67aWdBPJXopUOJ0aFprcqVhbEJusOvy8JniNy2XVDFxnJxi7lVEkoPQAOgw
IIed+8gB2tUIEQ8UBCtkbcA11LpKjChRj91dvUgnjmtWM7mzgen+sfvBZY/hVHEZ
MgRJ9ZUVdLhIr2ff11lgUPX6ijImhIzMQRKMP6MCgYBw03imwmHKM0jIrHL+z9Gz
DfUc0xP30WfDF49WI2mEfotkuj4m8kZXOuz2FGMV/3WGB8lMhygfW29Uz3qs8YMj
/H2pj6pW16/IHFAiq7t+ZfJEsgjCyI7jOqyhCdlY1Ouue+bR4bl4oqIKhqZByrUV
YoDcMv8kd4KFhjvgdQPl8A==
-----END PRIVATE KEY-----',
    'RS256'
);

-- Registration configuration (update URLs to match your setup)
INSERT INTO registration VALUES(
    '56f3d0ed-0e0a-4ba5-a5a2-59aa4bbe6b57',
    'https://issuer.example.com',           -- Your LMS platform issuer URL
    '457df601-695d-4ba6-8fbf-fef291ab3fb6', -- Your client ID
    'https://example.com/oidc/login',       -- LMS OIDC login endpoint
    'https://example.com/service/token',    -- LMS service token endpoint
    'https://example.com/.well-known/jwks', -- LMS JWKS endpoint
    null,
    'https://lti-asset-processor-<user>.ngrok.io/lti/launch', -- Your ngrok URL
    'd48a53de-021f-46f7-a0a4-7134812c2235' -- Key set ID (must match above)
);

-- Deployment configuration
INSERT INTO deployment VALUES(
    '2394b381-8012-4310-98c4-76ef1d252157',
    '56f3d0ed-0e0a-4ba5-a5a2-59aa4bbe6b57',
    'Example Customer'
);
```

**Note**: Replace all placeholder values ( URLs, client IDs, etc.) with your actual LMS platform configuration details.

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd asset-processor-demo-tool
```

### 2. Create Environment Configuration

Create a `.env` file in the project root directory:

```env
# Optional: Ngrok configuration (for external access)
NGROK_AUTH=your_ngrok_authtoken
NGROK_DOMAIN=your_custom_domain

# LTI Server configuration (optional, defaults are shown)
# LTI_SERVER_PORT=9002
```

**Important**: 
- If you need external access (for LMS integration), add your `NGROK_AUTH` token
- If you don't need external access, you can omit the `NGROK_AUTH` variable and the ngrok service won't start

### 3. Prepare Database Registration

Create the `db/zdata.sql` file with your registration configuration (see [Database Registration Configuration](#database-registration-configuration) above).

### 4. Build and Start the Application

Start all services (PostgreSQL, LLM server, and the asset processor):

```bash
docker-compose up --build
```

This command will:
- Build the Go application
- Start PostgreSQL database with the initial schema
- Start the LLM server (llama.cpp) for text processing
- Start the asset processor application
- Start ngrok (if `NGROK_AUTH` is configured)

### 5. Verify the Application

The application should now be running. You can verify this by:

1. **Checking logs**: Look for the message `Server started on port 8000` in the terminal output
2. **Testing the endpoint**: The application is available at `http://localhost:9002` (or your custom port)
3. **Ngrok URL**: If configured, check the ngrok dashboard at `http://localhost:4040` for the public URL

## Available Services

| Service | Port | Description |
|---------|------|-------------|
| Asset Processor | 9002 (host) / 8000 (container) | Main application server |
| PostgreSQL | 5422 (host) / 5432 (container) | Database server |
| LLM Server | 8080 | Llama.cpp server for text processing |
| Ngrok (if configured) | 4040 | Public URL tunnel |

## Stopping the Application

To stop all services:

```bash
docker-compose down
```

To stop and remove volumes (database data):

```bash
docker-compose down -v
```

## Running in Detached Mode

To run the services in the background:

```bash
docker-compose up -d
```

To view logs:

```bash
docker-compose logs -f
```

## Accessing Individual Containers

### Access PostgreSQL:

```bash
docker-compose exec postgres psql -U postgres -d postgres
```

### Access Asset Processor:

```bash
docker-compose exec ap-server bash
```

## Troubleshooting

### Database Connection Issues

If you encounter database connection issues:

1. Ensure PostgreSQL is running: `docker-compose ps`
2. Check PostgreSQL logs: `docker-compose logs postgres`
3. Verify database schema is loaded: `docker-compose exec postgres psql -U postgres -d postgres -c "\dt"`

### LLM Server Issues

If the LLM server is not responding:

1. Check if the model file exists in `./llm/` directory
2. Verify the model name in `docker-compose.yml` matches your file
3. Check LLM server logs: `docker-compose logs llamacpp-server`

### Port Already in Use

If you see "port already in use" errors:

1. Check what's using the port: `lsof -i :9002` (or your port)
2. Change the port mapping in `docker-compose.yml`
3. Or stop the conflicting service

### Ngrok Not Connecting

If ngrok is not connecting:

1. Verify your `NGROK_AUTH` token is correct
2. Check ngrok logs: `docker-compose logs ngrok`
3. Ensure your firewall allows outbound connections

## Development

### Running Unit Tests

```bash
docker-compose exec ap-server go test ./src/processors/...
```

### Rebuilding the Application

```bash
docker-compose build --no-cache ap-server
```

## Architecture Overview

The application consists of the following components:

- **LTI Message Handling**: Processes LTI launches through `/lti/launch` endpoint
- **OIDC Authentication**: Handles OpenID Connect authentication via `/oidc/login` and `/.well-known/jwks.json`
- **Asset Processing**: Uses processors in `/src/processors/` to analyze submitted assets
- **Database Integration**: Uses PostgreSQL for storing registration information and asset reports
- **External LLM Service**: Communicates with a local LLM server for text processing

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/lti/launch` | POST | LTI launch endpoint for asset processing |
| `/oidc/login` | GET | OIDC authentication endpoint |
| `/.well-known/jwks` | GET | JWKS endpoint for token validation |
| `/lti/notice` | POST | LTI platform notification service endpoint |
| `/lti/deeplink/return` | POST | Deep linking return endpoint |

## License

This code is provided entirely for demonstration purposes and comes with zero warranty.