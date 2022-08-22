#!/bin/bash

# init environment variables
source .env.client

# request an access token
TKN_RESP=$(curl -s \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -X POST https://$AUTH0_DOMAIN/oauth/token \
    --data @<(cat <<EOF
{
    "client_id": "$OAUTH_CLIENT_ID",
    "client_secret": "$OAUTH_CLIENT_SECRET",
    "audience": "$AUTH0_AUDIENCE",
    "grant_type": "client_credentials"
}
EOF
)) || { echo "Could not get access token" >&2 ; exit 1; }

# check for errors in the response
if echo $TKN_RESP | jq -e 'has("error")' > /dev/null; then
    echo "Got error in token response" >&2
    echo $TKN_RESP | jq >&2
    exit 1
fi

# parse json to get access token
TOKEN=$(echo $TKN_RESP | jq -r '.access_token')

echo $TKN_RESP | jq

curl -s \
    -D /dev/stderr \
    -w "\n" \
    --cacert ca.crt \
    -H "Authorization: Bearer ${TOKEN}" \
    -X GET https://localhost:$PORT/api/v1/products
