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

# Get the latest backup directory
get_latest_backup_dir() {
    local base_dir="$1"
    local log_file="$2"
    
    echo "🔍 Using provided backup directory: $base_dir" | tee -a "${log_file}"
    
    # Just return the base_dir since it's already a timestamped directory created in run_export
    echo "$base_dir"
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
    
    # Create directory for output file if it doesn't exist
    mkdir -p "$(dirname "$output_file")"
    
    echo "📥 Requesting data from: ${api_url}/users/me/${endpoint}" | tee -a "${log_file}"
    echo "🔑 Using access token: ${access_token:0:5}...${access_token: -5}" | tee -a "${log_file}"
    echo "💾 Saving to: ${output_file}" | tee -a "${log_file}"
    
    # Set up initial pagination parameters
    local page=1
    local limit=50000 # Maximum allowed by Trakt API
    # No safety limit (will use total_pages from API response)
    local total_items=0
    local total_pages=0
    local temp_file="${output_file}.temp"
    local success=false
    local max_retries=3
    local retry_count=0
    
    # Initialize an empty array in our temp file
    echo "[]" > "$temp_file"
    
    # Make initial request to get total page count from headers
    local initial_response_headers=$(curl -s -I -X GET "${api_url}/users/me/${endpoint}?page=1&limit=${limit}" \
        -H "Content-Type: application/json" \
        -H "trakt-api-key: ${api_key}" \
        -H "trakt-api-version: 2" \
        -H "Authorization: Bearer ${access_token}")
    
    # Try to extract total page count from X-Pagination-Page-Count header
    if [[ "$initial_response_headers" == *"X-Pagination-Page-Count:"* ]]; then
        total_pages=$(echo "$initial_response_headers" | grep -i "X-Pagination-Page-Count:" | awk '{print $2}' | tr -d '\r')
        local total_count=$(echo "$initial_response_headers" | grep -i "X-Pagination-Item-Count:" | awk '{print $2}' | tr -d '\r' || echo "?")
        
        echo "📚 PAGES ESTIMATION FOR $endpoint:" | tee -a "${log_file}"
        echo "   ┌─────────────────────────────────────────────┐" | tee -a "${log_file}"
        echo "   │ Total pages: $total_pages (limit: $limit items/page) │" | tee -a "${log_file}"
        echo "   │ Total items: $total_count                              │" | tee -a "${log_file}"
        
        # Calculate estimated time (2s per page + 1s retry buffer)
        if [[ "$total_pages" =~ ^[0-9]+$ ]]; then
            local estimated_seconds=$((total_pages * 3))
            local estimated_minutes=$((estimated_seconds / 60))
            local remaining_seconds=$((estimated_seconds % 60))
            echo "   │ Est. time: ~${estimated_minutes}m ${remaining_seconds}s                           │" | tee -a "${log_file}"
        fi
        
        echo "   └─────────────────────────────────────────────┘" | tee -a "${log_file}"
    else
        echo "⚠️ Could not determine total page count from API response" | tee -a "${log_file}"
        echo "⚠️ Will fetch pages until no more data is returned" | tee -a "${log_file}"
        # Set a large default value to effectively remove the limit
        total_pages=1000000
    fi
    
    # Make paginated API requests until we get all data
    while [ $page -le $total_pages ]; do
        retry_count=0
        local page_success=false
        
        # Calculate and display progress percentage
        if [[ "$total_pages" =~ ^[0-9]+$ ]] && [ $total_pages -gt 0 ]; then
            local progress_pct=$((page * 100 / total_pages))
            local progress_bar=""
            local bar_width=20
            local filled_width=$((progress_pct * bar_width / 100))
            local empty_width=$((bar_width - filled_width))
            
            # Create the progress bar
            progress_bar="["
            for ((i=0; i<filled_width; i++)); do progress_bar+="▓"; done
            for ((i=0; i<empty_width; i++)); do progress_bar+="░"; done
            progress_bar+="] ${progress_pct}%"
            
            echo "⏳ Progress: $progress_bar (Page $page of $total_pages)" | tee -a "${log_file}"
        fi
        
        while [ $retry_count -lt $max_retries ] && [ "$page_success" != "true" ]; do
            echo "📄 Fetching page $page of $total_pages for endpoint $endpoint..." | tee -a "${log_file}"
            
            # Make the paginated API request
            local response=$(curl -s -X GET "${api_url}/users/me/${endpoint}?page=${page}&limit=${limit}" \
                -H "Content-Type: application/json" \
                -H "trakt-api-key: ${api_key}" \
                -H "trakt-api-version: 2" \
                -H "Authorization: Bearer ${access_token}" \
                -D "${temp_file}.headers.${page}")
            
            # Extract pagination info from headers for this request
            if [ -f "${temp_file}.headers.${page}" ]; then
                local current_page=$(grep -i "X-Pagination-Page:" "${temp_file}.headers.${page}" | awk '{print $2}' | tr -d '\r' || echo "?")
                local page_count=$(grep -i "X-Pagination-Page-Count:" "${temp_file}.headers.${page}" | awk '{print $2}' | tr -d '\r' || echo "?")
                local total_count=$(grep -i "X-Pagination-Item-Count:" "${temp_file}.headers.${page}" | awk '{print $2}' | tr -d '\r' || echo "?")
                
                # Update total_pages if we have more accurate info
                if [[ "$page_count" != "?" ]]; then
                    total_pages=$page_count
                fi
                
                echo "📊 Pagination: Page $current_page of $page_count (Total items: $total_count)" | tee -a "${log_file}"
                rm -f "${temp_file}.headers.${page}"
            fi
            
            # Check if the response is valid JSON and not empty
            if echo "$response" | jq empty 2>/dev/null && [ "$(echo "$response" | jq 'length')" -gt 0 ]; then
                # Save the current items and merge with previous pages
                echo "$response" > "${temp_file}.page${page}"
                
                # Merge with existing data
                jq -s 'add' "$temp_file" "${temp_file}.page${page}" > "${temp_file}.new"
                mv "${temp_file}.new" "$temp_file"
                rm "${temp_file}.page${page}"
                
                # Get item count for this page
                local items_count=$(echo "$response" | jq 'length')
                total_items=$((total_items + items_count))
                echo "✅ Page $page: Retrieved $items_count items (running total: $total_items)" | tee -a "${log_file}"
                
                # If fewer items than the limit, we've reached the end
                if [ $items_count -lt $limit ]; then
                    echo "🏁 Reached end of data for endpoint $endpoint (fewer items than limit)" | tee -a "${log_file}"
                    success=true
                    break
                fi
                
                page_success=true
                page=$((page + 1))
            elif [ $retry_count -lt $((max_retries - 1)) ]; then
                retry_count=$((retry_count + 1))
                echo "⚠️ Retry $retry_count for page $page (endpoint: $endpoint)" | tee -a "${log_file}"
                sleep 2 # Wait before retrying
            else
                echo -e "\e[33mWARNING: Failed to retrieve page $page for endpoint $endpoint after $max_retries attempts.\e[0m" | tee -a "${log_file}"
                # If we got at least one page successfully, consider it a partial success
                if [ $total_items -gt 0 ]; then
                    success=true
                    echo "⚠️ Continuing with partial data ($total_items items)" | tee -a "${log_file}"
                    break
                else
                    echo -e "\e[31mERROR: Failed to retrieve any data for endpoint $endpoint.\e[0m" | tee -a "${log_file}"
                    return 1
                fi
            fi
        done
        
        # If we've exhausted all pages or reached the end, break
        if [ "$page_success" != "true" ] || [ "$success" = "true" ]; then
            break
        fi
    done
    
    # If we got here with data, move the temporary file to the final location
    if [ "$success" = "true" ] || [ $total_items -gt 0 ]; then
        mv "$temp_file" "$output_file"
        echo "📊 Successfully saved $total_items items for endpoint $endpoint" | tee -a "${log_file}"
        echo -e "\e[32m${username}/${endpoint}\e[0m Retrieved successfully ($total_items items in ${page} pages)" | tee -a "${log_file}"
        return 0
    else
        echo -e "\e[31mERROR: Failed to retrieve data for endpoint ${endpoint}.\e[0m" | tee -a "${log_file}"
        # If we have a temporary file but no data, clean up
        rm -f "$temp_file"
        return 1
    fi
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