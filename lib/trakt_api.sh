#!/bin/bash
#
# Trakt API interaction functions
#

# Refresh access token if expired
refresh_access_token() {
    local refresh_token="$1"
    local api_key="$2"
    local api_secret="$3"
    local redirect_uri="$4"
    local config_file="$5"
    local sed_inplace="$6"
    local log_file="$7"
    
    echo "🔄 Refreshing Trakt token..." | tee -a "${log_file}"
    echo "  - Using refresh token: ${refresh_token:0:5}...${refresh_token: -5}" | tee -a "${log_file}"
    echo "  - API key: ${api_key:0:5}...${api_key: -5}" | tee -a "${log_file}"
    
    local response=$(curl -s -X POST "https://api.trakt.tv/oauth/token" \
        -H "Content-Type: application/json" -v \
        -d "{
            \"refresh_token\": \"${refresh_token}\",
            \"client_id\": \"${api_key}\",
            \"client_secret\": \"${api_secret}\",
            \"redirect_uri\": \"${redirect_uri}\",
            \"grant_type\": \"refresh_token\"
        }")

    # Debug response (without exposing sensitive data)
    echo "  - Response received: $(if [ -n "$response" ]; then echo "✅"; else echo "❌ (empty)"; fi)" | tee -a "${log_file}"
    
    local new_access_token=$(echo "$response" | jq -r '.access_token')
    local new_refresh_token=$(echo "$response" | jq -r '.refresh_token')

    if [[ "$new_access_token" != "null" && "$new_refresh_token" != "null" ]]; then
        echo "✅ Token refreshed successfully." | tee -a "${log_file}"
        echo "  - New access token: ${new_access_token:0:5}...${new_access_token: -5}" | tee -a "${log_file}"
        echo "  - New refresh token: ${new_refresh_token:0:5}...${new_refresh_token: -5}" | tee -a "${log_file}"
        
        # Determine which config file to update
        if [ ! -f "$config_file" ]; then
            echo "  - Config file not found: $config_file" | tee -a "${log_file}"
            return 1
        fi
        
        echo "  - Updating config file: $config_file" | tee -a "${log_file}"
        
        # Check if config file exists and is writable
        if [ -f "$config_file" ]; then
            if [ -w "$config_file" ]; then
                echo "  - Config file is writable: ✅" | tee -a "${log_file}"
            else
                echo "  - Config file is writable: ❌ - Permissions: $(ls -la "$config_file" | awk '{print $1}')" | tee -a "${log_file}"
                return 1
            fi
        else
            echo "  - Config file exists: ❌ (not found)" | tee -a "${log_file}"
            return 1
        fi
        
        $sed_inplace "s|ACCESS_TOKEN=.*|ACCESS_TOKEN=\"$new_access_token\"|" "$config_file"
        $sed_inplace "s|REFRESH_TOKEN=.*|REFRESH_TOKEN=\"$new_refresh_token\"|" "$config_file"
        
        echo "  - Config file updated: $(if [ $? -eq 0 ]; then echo "✅"; else echo "❌"; fi)" | tee -a "${log_file}"
        
        # Return the new tokens as a string
        echo "${new_access_token}:${new_refresh_token}"
        return 0
    else
        echo "❌ Error refreshing token. Check your configuration!" | tee -a "${log_file}"
        echo "  - Response: $response" | tee -a "${log_file}"
        echo "  - Make sure your API credentials are correct and try again." | tee -a "${log_file}"
        return 1
    fi
}

# Check if the token is valid
check_token_validity() {
    local api_url="$1"
    local api_key="$2"
    local access_token="$3"
    local log_file="$4"
    
    echo "🔒 Checking token validity..." | tee -a "${log_file}"
    
    local response=$(curl -s -X GET "${api_url}/users/me/history/movies" \
        -H "Content-Type: application/json" \
        -H "trakt-api-key: ${api_key}" \
        -H "trakt-api-version: 2" \
        -H "Authorization: Bearer ${access_token}")
    
    if echo "$response" | grep -q "invalid_grant"; then
        echo "⚠️ Token expired or invalid" | tee -a "${log_file}"
        return 1
    else
        echo "✅ Token is valid" | tee -a "${log_file}"
        return 0
    fi
}

# Fetch data from Trakt API
fetch_trakt_data() {
    local api_url="$1"
    local api_key="$2"
    local access_token="$3"
    local endpoint="$4"
    local output_file="$5"
    local username="$6"
    local log_file="$7"
    
    # Check if tokens are defined
    if [ -z "$access_token" ] || [ "$access_token" = '""' ] || [ "$access_token" = "" ]; then
        echo -e "\e[31mERROR: ACCESS_TOKEN not defined. Run the setup_trakt.sh script first to get a token.\e[0m" | tee -a "${log_file}"
        echo -e "Command: ./setup_trakt.sh" | tee -a "${log_file}"
        return 1
    fi
    
    # Make the API request
    curl -s -X GET "${api_url}/users/me/${endpoint}" \
        -H "Content-Type: application/json" \
        -H "trakt-api-key: ${api_key}" \
        -H "trakt-api-version: 2" \
        -H "Authorization: Bearer ${access_token}" \
        -o "${output_file}" \
        && echo -e "\e[32m${username}/${endpoint}\e[0m Retrieved successfully" | tee -a "${log_file}" \
        || { echo -e "\e[31m${username}/${endpoint}\e[0m Request failed" | tee -a "${log_file}"; return 1; }
    
    # Validate the response
    if [ -f "${output_file}" ]; then
        if ! jq empty "${output_file}" 2>/dev/null; then
            echo -e "\e[31mERROR: The file $(basename "${output_file}") does not contain valid JSON.\e[0m" | tee -a "${log_file}"
            return 1
        elif [ "$(jq '. | length' "${output_file}")" = "0" ]; then
            echo -e "\e[33mWARNING: The file $(basename "${output_file}") does not contain any data.\e[0m" | tee -a "${log_file}"
            # This is a warning, not an error, so still return success
            return 0
        fi
    else
        echo -e "\e[31mERROR: The file $(basename "${output_file}") was not created.\e[0m" | tee -a "${log_file}"
        return 1
    fi
    
    return 0
}

# Get endpoints based on mode
get_endpoints_for_mode() {
    local mode="$1"
    local log_file="$2"
    
    case "$mode" in
        "complete")
            echo -e "Complete Mode activated" | tee -a "${log_file}"
            echo "watchlist/movies watchlist/shows watchlist/episodes watchlist/seasons ratings/movies ratings/shows ratings/episodes ratings/seasons collection/movies collection/shows watched/movies watched/shows history/movies history/shows"
            ;;
        "initial")
            echo -e "Initial Mode activated" | tee -a "${log_file}"
            echo "history/movies ratings/movies watched/movies"
            ;;
        *)
            echo -e "Normal Mode activated" | tee -a "${log_file}"
            echo "history/movies ratings/movies ratings/episodes history/movies history/shows history/episodes watchlist/movies watchlist/shows"
            ;;
    esac
} 