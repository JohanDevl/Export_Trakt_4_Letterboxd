#!/bin/bash
# health_check.sh - Health check script for Docker container

# Source config if available
CONFIG_FILE="${CONFIG_DIR:-/app/config}/.config.cfg"
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
fi

# Check status variables
STATUS="ok"
DETAILS=()

# Check for required directories
check_directory() {
    if [[ ! -d "$1" ]]; then
        STATUS="error"
        DETAILS+=("Directory $1 not found or not accessible")
        return 1
    elif [[ ! -w "$1" ]]; then
        STATUS="error"
        DETAILS+=("Directory $1 not writable")
        return 1
    fi
    return 0
}

# Check for required files
check_file() {
    if [[ ! -f "$1" ]]; then
        STATUS="error"
        DETAILS+=("File $1 not found or not accessible")
        return 1
    elif [[ ! -r "$1" ]]; then
        STATUS="error"
        DETAILS+=("File $1 not readable")
        return 1
    fi
    return 0
}

# Check for required commands
check_command() {
    if ! command -v "$1" &> /dev/null; then
        STATUS="error"
        DETAILS+=("Command $1 not found or not executable")
        return 1
    fi
    return 0
}

# Check API connectivity (if tokens are available)
check_api_connectivity() {
    if [[ -n "$ACCESS_TOKEN" && -n "$API_URL" ]]; then
        # Attempt a simple API call
        local api_response
        api_response=$(curl -s -f -H "Content-Type: application/json" \
                           -H "Authorization: Bearer $ACCESS_TOKEN" \
                           -H "trakt-api-version: 2" \
                           -H "trakt-api-key: $API_KEY" \
                           "${API_URL}/users/settings" 2>&1)
        
        if [[ $? -ne 0 ]]; then
            STATUS="warning"
            DETAILS+=("Cannot connect to Trakt API: $api_response")
            return 1
        fi
    else
        STATUS="warning"
        DETAILS+=("API credentials not configured")
        return 1
    fi
    return 0
}

# Run all checks
run_health_checks() {
    # Check essential directories
    check_directory "${DOSLOG:-/app/logs}"
    check_directory "${DOSCOPY:-/app/copy}"
    check_directory "${BACKUP_DIR:-/app/backup}"
    check_directory "${CONFIG_DIR:-/app/config}"
    
    # Check essential files
    check_file "/app/Export_Trakt_4_Letterboxd.sh"
    check_file "/app/docker-entrypoint.sh"
    
    # Check essential commands
    check_command "bash"
    check_command "curl"
    check_command "jq"
    check_command "sed"
    
    # Check API connectivity
    check_api_connectivity
    
    # Prepare response
    local health_response
    health_response=$(cat <<EOF
{
  "status": "$STATUS",
  "version": "${APP_VERSION:-unknown}",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "details": [
    $(printf '"%s",' "${DETAILS[@]}" | sed 's/,$//')
  ]
}
EOF
    )
    
    echo "$health_response"
    
    # Return success or failure based on status
    if [[ "$STATUS" == "ok" ]]; then
        return 0
    else
        return 1
    fi
}

# If script is called directly, run health checks and output results
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    run_health_checks
    exit $?
fi 