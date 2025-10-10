#!/bin/sh

KEYCLOAK_URL="http://localhost:8081"
ADMIN_USER="admin"
ADMIN_PASSWORD="admin"

until curl -s http://localhost:8081/health/ready > /dev/null; do
  echo "Waiting for Keycloak..."
  sleep 5
done
echo "Keycloak is ready!"


echo "Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=$ADMIN_USER" \
  -d "password=$ADMIN_PASSWORD" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" | \
  sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

if [ -z "$ADMIN_TOKEN" ]; then
  echo "Failed to get admin token"
  exit 1
fi

echo "Admin token obtained successfully"

echo "Creating aispace realm..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "realm": "aispace",
    "displayName": "AI Space",
    "enabled": true,
    "sslRequired": "external",
    "registrationAllowed": false,
    "loginWithEmailAllowed": true,
    "duplicateEmailsAllowed": false,
    "resetPasswordAllowed": false,
    "editUsernameAllowed": false,
    "bruteForceProtected": true
  }'

echo "Realm created"

echo "Creating AISPACE client..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms/aispace/clients" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "AISPACE",
    "name": "AI Space Application",
    "description": "Main application client",
    "enabled": true,
    "clientAuthenticatorType": "client-secret",
    "secret": "your-secret-key-here",
    "redirectUris": [
      "http://localhost:3000/auth",
      "http://localhost:3000/*"
    ],
    "webOrigins": [
      "http://localhost:3000",
      "*"
    ],
    "standardFlowEnabled": true,
    "implicitFlowEnabled": false,
    "directAccessGrantsEnabled": false,
    "serviceAccountsEnabled": false,
    "publicClient": false,
    "frontchannelLogout": true,
    "protocol": "openid-connect",
    "attributes": {
      "saml.assertion.signature": "false",
      "saml.force.post.binding": "false",
      "saml.multivalued.roles": "false",
      "saml.encrypt": "false",
      "saml.server.signature": "false",
      "saml.server.signature.keyinfo.ext": "false",
      "exclude.session.state.from.auth.response": "false",
      "saml_force_name_id_format": "false",
      "saml.client.signature": "false",
      "tls.client.certificate.bound.access.tokens": "false",
      "saml.authnstatement": "false",
      "display.on.consent.screen": "false",
      "saml.onetimeuse.condition": "false"
    },
    "authenticationFlowBindingOverrides": {},
    "fullScopeAllowed": true,
    "nodeReRegistrationTimeout": -1,
    "defaultClientScopes": [
      "web-origins",
      "roles",
      "profile",
      "email"
    ],
    "optionalClientScopes": [
      "address",
      "phone",
      "offline_access",
      "microprofile-jwt"
    ]
  }'

echo "Client created"

echo "Creating test user..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms/aispace/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "firstName": "Test",
    "lastName": "User",
    "email": "test@example.com",
    "enabled": true,
    "emailVerified": true,
    "credentials": [
      {
        "type": "password",
        "value": "password123",
        "temporary": false
      }
    ]
  }'

echo "User created"
echo "Setup completed!"
echo ""
echo "Realm: aispace"
echo "Client ID: AISPACE"
echo "Client Secret: your-secret-key-here"
echo "Test User: testuser / password123"
echo "Provider URL: http://localhost:8081/realms/aispace"
