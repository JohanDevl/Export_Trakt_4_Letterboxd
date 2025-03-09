#!/bin/bash
SCRIPT_DIR=$(dirname "$(realpath "$0")")
CONFIG_FILE="${SCRIPT_DIR}/.config.cfg"

echo "=== Trakt Authentication Configuration ==="
echo ""
echo "This script will help you configure authentication with the Trakt API."
echo ""

# Check if the configuration file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: Configuration file does not exist."
    exit 1
fi

# Request API information
echo "Step 1: Create an application at https://trakt.tv/oauth/applications"
echo "         - Name: Export Trakt 4 Letterboxd"
echo "         - Redirect URL: urn:ietf:wg:oauth:2.0:oob"
echo ""
read -p "Enter your Client ID (API Key): " API_KEY
read -p "Enter your Client Secret: " API_SECRET
echo ""

# Update the configuration file
sed -i '' "s|API_KEY=.*|API_KEY=\"$API_KEY\"|" "$CONFIG_FILE"
sed -i '' "s|API_SECRET=.*|API_SECRET=\"$API_SECRET\"|" "$CONFIG_FILE"
sed -i '' "s|REDIRECT_URI=.*|REDIRECT_URI=\"urn:ietf:wg:oauth:2.0:oob\"|" "$CONFIG_FILE"

echo "Step 2: Get an authorization code"
echo ""
echo "Open the following link in your browser:"
echo "https://trakt.tv/oauth/authorize?response_type=code&client_id=${API_KEY}&redirect_uri=urn:ietf:wg:oauth:2.0:oob"
echo ""
read -p "Enter the displayed authorization code: " AUTH_CODE
echo ""

# Get tokens
echo "Step 3: Getting access tokens..."
RESPONSE=$(curl -s -X POST "https://api.trakt.tv/oauth/token" \
    -H "Content-Type: application/json" \
    -d "{
        \"code\": \"${AUTH_CODE}\",
        \"client_id\": \"${API_KEY}\",
        \"client_secret\": \"${API_SECRET}\",
        \"redirect_uri\": \"urn:ietf:wg:oauth:2.0:oob\",
        \"grant_type\": \"authorization_code\"
    }")

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refresh_token')

if [[ "$ACCESS_TOKEN" != "null" && "$REFRESH_TOKEN" != "null" && -n "$ACCESS_TOKEN" && -n "$REFRESH_TOKEN" ]]; then
    # Update the configuration file
    sed -i '' "s|ACCESS_TOKEN=.*|ACCESS_TOKEN=\"$ACCESS_TOKEN\"|" "$CONFIG_FILE"
    sed -i '' "s|REFRESH_TOKEN=.*|REFRESH_TOKEN=\"$REFRESH_TOKEN\"|" "$CONFIG_FILE"
    
    echo "✅ Configuration completed successfully!"
    echo ""
    echo "You can now run the Export_Trakt_4_Letterboxd.sh script"
else
    echo "❌ Error obtaining tokens."
    echo "API response: $RESPONSE"
    exit 1
fi 