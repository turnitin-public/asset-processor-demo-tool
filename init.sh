#!/bin/sh
set -e

LOG_FILE="/fluentd/log/tunnel_0.log"
MAX_WAIT=30  # seconds
ELAPSED=0

echo "⏳ Waiting for tunnel URL..."

while [ $ELAPSED -lt $MAX_WAIT ]; do
  if [ -f "$LOG_FILE" ]; then
    URL=$(grep -o 'https://[^ ]*\.trycloudflare\.com' "$LOG_FILE" | tail -n1)
    if [ -n "$URL" ]; then
      echo "✅ Found tunnel URL: $URL"
      export TUNNEL_URL="$URL"
      exec go run *.go
    fi
  fi
  sleep 1
  ELAPSED=$((ELAPSED + 1))
done

echo "❌ Timeout waiting for tunnel URL"
exit 1
