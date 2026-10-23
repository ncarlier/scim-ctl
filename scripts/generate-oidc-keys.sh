#!/bin/bash
set -e

# Configuration
KEY_NAME="keycloak-client"
PASSWORD="changeit"
SUBJECT="/CN=scim-ctl-client/O=ncarlier/C=FR"

echo "Generating private key and certificate for Keycloak Client Authentication..."

# 1. Generate a modern PKCS#8 encrypted private key
# We use genpkey which directly generates PKCS#8 format instead of legacy RSA format
echo "[1/2] Generating encrypted PKCS#8 private key ($KEY_NAME.key)..."
openssl genpkey -algorithm RSA \
    -out $KEY_NAME.key \
    -aes-256-cbc \
    -pass pass:$PASSWORD \
    -pkeyopt rsa_keygen_bits:2048

# 2. Generate a self-signed certificate using the private key
# Keycloak can import this certificate to verify the client_assertion JWTs 
# signed by scim-ctl.
echo "[2/2] Generating self-signed certificate ($KEY_NAME.crt)..."
openssl req -new -x509 \
    -key $KEY_NAME.key \
    -passin pass:$PASSWORD \
    -out $KEY_NAME.crt \
    -days 3650 \
    -subj "$SUBJECT"

echo ""
echo "============================================================"
echo "Success! Keys generated successfully."
echo ""
echo "Generated Files:"
echo "  - $KEY_NAME.key : The encrypted private key for scim-ctl."
echo "  - $KEY_NAME.crt : The certificate to import into Keycloak."
echo ""
echo "Private Key Password: $PASSWORD"
echo ""
echo "Next steps in Keycloak:"
echo "1. Go to your Client > 'Credentials' tab."
echo "2. Set 'Client Authenticator' to 'Signed Jwt'."
echo "3. Click 'Keys' tab, then 'Import'."
echo "4. Archive format: 'Certificate PEM'"
echo "5. Select the '$KEY_NAME.crt' file generated."
echo ""
echo "Next steps for scim-ctl:"
echo "Configure scim-ctl using the following flags (or ENV vars):"
echo "  --oidc-grant-type=client_credentials"
echo "  --oidc-private-key=$KEY_NAME.key"
echo "  --oidc-private-key-password=$PASSWORD"
echo "============================================================"
