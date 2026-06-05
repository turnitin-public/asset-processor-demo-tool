## Warning: This code is provided entirely for demonstration purposes and comes with zero warranty. Any use of this software is done so at the user's own risk

## Prerequisites
### Ngrok
For the asset processor tool to function correctly it is required that the service be externally accessible. To do this, ngrok is used to proxy incoming requests. For ngrok to allow for the service to be exposed on a consistent url, a paid account is required. If you have a paid account you will need an auth token and to add it as an "AUTH_TOKEN" environment variable in a `.env` file.

Troubleshooting note: if ngrok logs `failed to fetch CRL` or DNS errors for `*.ngrok-agent.com` (for example while running behind endpoint security tools such as Zscaler), DNS resolution may be blocked intermittently. As a temporary workaround, add `extra_hosts` entries for `crl.ngrok-agent.com`, `update.ngrok-agent.com`, and `connect.us.ngrok-agent.com` in the `ngrok` service in `docker-compose.yml`.

### Dynamic Registration Bootstrap
The application now bootstraps signing key material automatically on startup. If the `key_set` and `a_key` tables are empty, the tool generates a new RSA signing key and stores it in the database.

For new installs, you do not need a `zdata.sql` file to seed key material, registrations, or deployments. Registrations and deployments are created through the dynamic registration flow at runtime.

You may still use a seed file if you want deterministic local test data, but it is no longer required for the tool to start successfully.