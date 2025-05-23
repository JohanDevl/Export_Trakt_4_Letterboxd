#!/bin/bash
set -e

# Improved logging function
log_message() {
    local level="$1"
    local message="$2"
    local timestamp=$(date "+%Y-%m-%d %H:%M:%S")
    
    case "$level" in
        "INFO")  echo -e "ℹ️ [INFO] $timestamp - $message" ;;
        "WARN")  echo -e "⚠️ [WARNING] $timestamp - $message" ;;
        "ERROR") echo -e "❌ [ERROR] $timestamp - $message" ;;
        "DEBUG") echo -e "🔍 [DEBUG] $timestamp - $message" ;;
        "SUCCESS") echo -e "✅ [SUCCESS] $timestamp - $message" ;;
    esac
}

# Show version information
show_version() {
    log_message "INFO" "Starting Export Trakt 4 Letterboxd container - Version: ${APP_VERSION:-unknown}"
}

# Health check HTTP server
start_health_server() {
    # Check if netcat-openbsd is installed
    if ! command -v nc &> /dev/null; then
        log_message "WARN" "Netcat not installed. Health server not available. Installing..."
        if command -v apk &> /dev/null; then
            apk add --no-cache netcat-openbsd
        else
            log_message "ERROR" "Package manager not found. Cannot install netcat."
            return 1
        fi
    fi
    
    # Source the health check script
    source /app/lib/health_check.sh

    # Start health check server
    log_message "INFO" "Starting health check server on port 8000"
    
    # Run in background with BusyBox compatible options
    (
        while true; do
            # For BusyBox nc, we need to use different syntax
            # Write the HTTP response to a temporary file first
            echo -e "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n$(run_health_checks)" > /tmp/health_response
            # Start netcat in listen mode
            nc -l -p 8000 < /tmp/health_response
            # Small delay to avoid CPU spinning
            sleep 0.1
        done
    ) &
    
    # Store PID to kill server when container stops
    HEALTH_SERVER_PID=$!
    log_message "INFO" "Health check server started with PID: $HEALTH_SERVER_PID"
    
    # Register trap to kill server on exit
    trap "log_message 'INFO' 'Stopping health check server'; kill $HEALTH_SERVER_PID 2>/dev/null || true" EXIT INT TERM
}

# Debug function for file and directory information
debug_file_info() {
    local path="$1"
    local type="$2"
    
    if [ -e "$path" ]; then
        log_message "DEBUG" "$type exists: $path"
        if [ -d "$path" ]; then
            log_message "DEBUG" "Directory permissions: $(stat -c '%a %n' "$path" 2>/dev/null || ls -la "$path" | head -n 1)"
            log_message "DEBUG" "Owner/Group: $(stat -c '%U:%G' "$path" 2>/dev/null || ls -la "$path" | head -n 1 | awk '{print $3":"$4}')"
            log_message "DEBUG" "Content count: $(ls -la "$path" | wc -l) items"
        elif [ -f "$path" ]; then
            log_message "DEBUG" "File permissions: $(stat -c '%a %n' "$path" 2>/dev/null || ls -la "$path" | head -n 1)"
            log_message "DEBUG" "Owner/Group: $(stat -c '%U:%G' "$path" 2>/dev/null || ls -la "$path" | head -n 1 | awk '{print $3":"$4}')"
            log_message "DEBUG" "File size: $(stat -c '%s' "$path" 2>/dev/null || ls -la "$path" | awk '{print $5}') bytes"
            
            if [ -s "$path" ]; then
                log_message "DEBUG" "File has content"
            else
                log_message "WARN" "File is empty"
            fi
            
            if [ -r "$path" ]; then
                log_message "DEBUG" "File is readable"
            else
                log_message "ERROR" "File is not readable"
            fi
            
            if [ -w "$path" ]; then
                log_message "DEBUG" "File is writable"
            else
                log_message "ERROR" "File is not writable"
            fi
        fi
    else
        log_message "ERROR" "$type does not exist: $path"
        log_message "DEBUG" "Parent directory exists: $(if [ -d "$(dirname "$path")" ]; then echo "Yes"; else echo "No"; fi)"
        if [ -d "$(dirname "$path")" ]; then
            log_message "DEBUG" "Parent directory permissions: $(stat -c '%a %n' "$(dirname "$path")" 2>/dev/null || ls -la "$(dirname "$path")" | head -n 1)"
        fi
    fi
}

# Function to clean temporary directories safely
clean_temp_directories() {
    log_message "INFO" "Cleaning temporary directories..."
    
    # Define temp directories to clean
    TEMP_DIRS=("/app/TEMP")
    
    for dir in "${TEMP_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            log_message "DEBUG" "Cleaning directory: $dir"
            
            # First try with current permissions
            if find "$dir" -mindepth 1 -delete 2>/dev/null; then
                log_message "SUCCESS" "Cleaned $dir successfully"
            else
                log_message "WARN" "Permission issues cleaning $dir, attempting with elevated permissions"
                
                # Try to make the directory writable if needed
                chmod -R 777 "$dir" 2>/dev/null || true
                find "$dir" -mindepth 1 -delete 2>/dev/null || log_message "ERROR" "Failed to clean $dir completely"
                
                # Make sure the directory exists and has correct permissions 
                mkdir -p "$dir" 2>/dev/null || true
                chmod -R 777 "$dir" 2>/dev/null || true
            fi
        else
            log_message "WARN" "Directory $dir does not exist, creating it"
            mkdir -p "$dir" 2>/dev/null || log_message "ERROR" "Failed to create $dir"
            chmod -R 777 "$dir" 2>/dev/null || log_message "ERROR" "Failed to set permissions on $dir"
        fi
    done
}

# Function to sync environment variables with config file
sync_env_to_config() {
    local config_file="/app/config/.config.cfg"
    
    log_message "INFO" "Checking for environment variables to sync to config..."
    
    # Check if config file is writable
    if [ ! -w "$config_file" ]; then
        log_message "WARN" "Config file is not writable: $config_file"
        log_message "INFO" "Attempting to make config file writable"
        chmod 666 "$config_file" 2>/dev/null || log_message "ERROR" "Failed to make config file writable"
    fi
    
    # Re-check if it's writable
    if [ ! -w "$config_file" ]; then
        log_message "ERROR" "Cannot write to config file: $config_file"
        log_message "INFO" "Will use environment variables directly without updating config file"
        return 1
    fi
    
    # Create a temp file for safer editing
    local temp_config="/tmp/config.tmp"
    cp "$config_file" "$temp_config"
    
    # List of environment variables to check and sync
    declare -A env_vars
    env_vars[TRAKT_API_KEY]="API_KEY"
    env_vars[TRAKT_API_SECRET]="API_SECRET"
    env_vars[TRAKT_USERNAME]="USERNAME"
    
    # Special handling for tokens - only update if they are empty in the config
    # Get current values from config using grep with awk which is more compatible
    current_access_token=$(grep "^ACCESS_TOKEN=" "$config_file" | awk -F '"' '{print $2}' || echo "")
    current_refresh_token=$(grep "^REFRESH_TOKEN=" "$config_file" | awk -F '"' '{print $2}' || echo "")
    
    # Only update tokens if they are empty in the config
    if [ -z "$current_access_token" ] && [ -n "$TRAKT_ACCESS_TOKEN" ]; then
        log_message "INFO" "Setting ACCESS_TOKEN from environment variable"
        sed -i 's|^ACCESS_TOKEN=.*|ACCESS_TOKEN="'"$TRAKT_ACCESS_TOKEN"'"|' "$temp_config"
    fi
    
    if [ -z "$current_refresh_token" ] && [ -n "$TRAKT_REFRESH_TOKEN" ]; then
        log_message "INFO" "Setting REFRESH_TOKEN from environment variable"
        sed -i 's|^REFRESH_TOKEN=.*|REFRESH_TOKEN="'"$TRAKT_REFRESH_TOKEN"'"|' "$temp_config"
    fi
    
    # Check each environment variable (except tokens which are handled above)
    for env_var in "${!env_vars[@]}"; do
        config_var="${env_vars[$env_var]}"
        
        # If environment variable is set, update config
        if [ -n "${!env_var}" ]; then
            log_message "INFO" "Setting $config_var from environment variable $env_var"
            
            if grep -q "^$config_var=" "$temp_config"; then
                # Update existing variable - preserve format, just update value
                sed -i "s|^$config_var=.*|$config_var=\"${!env_var}\"|" "$temp_config"
            else
                # Add new variable (should rarely happen)
                echo "$config_var=\"${!env_var}\"" >> "$temp_config"
            fi
        fi
    done
    
    # Also check for environment variables with _FILE suffix for Docker secrets
    # Special handling for token secrets
    if [ -n "$TRAKT_ACCESS_TOKEN_FILE" ] && [ -f "$TRAKT_ACCESS_TOKEN_FILE" ] && [ -z "$current_access_token" ]; then
        secret_value=$(cat "$TRAKT_ACCESS_TOKEN_FILE" 2>/dev/null | tr -d '\n')
        if [ -n "$secret_value" ]; then
            log_message "INFO" "Setting ACCESS_TOKEN from secret file"
            sed -i 's|^ACCESS_TOKEN=.*|ACCESS_TOKEN="'"$secret_value"'"|' "$temp_config"
        fi
    fi
    
    if [ -n "$TRAKT_REFRESH_TOKEN_FILE" ] && [ -f "$TRAKT_REFRESH_TOKEN_FILE" ] && [ -z "$current_refresh_token" ]; then
        secret_value=$(cat "$TRAKT_REFRESH_TOKEN_FILE" 2>/dev/null | tr -d '\n')
        if [ -n "$secret_value" ]; then
            log_message "INFO" "Setting REFRESH_TOKEN from secret file"
            sed -i 's|^REFRESH_TOKEN=.*|REFRESH_TOKEN="'"$secret_value"'"|' "$temp_config"
        fi
    fi
    
    # For other secrets
    for env_var in "${!env_vars[@]}"; do
        secret_env_var="${env_var}_FILE"
        config_var="${env_vars[$env_var]}"
        
        # If secret file environment variable is set
        if [ -n "${!secret_env_var}" ] && [ -f "${!secret_env_var}" ]; then
            # Read the secret from file
            secret_value=$(cat "${!secret_env_var}" 2>/dev/null | tr -d '\n')
            
            if [ -n "$secret_value" ]; then
                log_message "INFO" "Setting $config_var from secret file $secret_env_var"
                
                if grep -q "^$config_var=" "$temp_config"; then
                    # Update existing variable
                    sed -i "s|^$config_var=.*|$config_var=\"$secret_value\"|" "$temp_config"
                else
                    # Add new variable
                    echo "$config_var=\"$secret_value\"" >> "$temp_config"
                fi
            else
                log_message "WARN" "Secret file for $env_var is empty, skipping"
            fi
        fi
    done
    
    # Copy the temp file back to the actual config
    if ! cp "$temp_config" "$config_file"; then
        log_message "ERROR" "Failed to update config file from temp file"
        log_message "DEBUG" "Temp file: $(cat "$temp_config")"
        return 1
    fi
    
    log_message "SUCCESS" "Config file updated with environment variables"
    rm -f "$temp_config"
    return 0
}

# Initial system information
log_message "INFO" "Starting Docker container for Export_Trakt_4_Letterboxd"
show_version
log_message "DEBUG" "Container environment:"
log_message "DEBUG" "User: $(id)"
log_message "DEBUG" "Working directory: $(pwd)"
log_message "DEBUG" "Environment variables:"
log_message "DEBUG" "- TZ: ${TZ:-Not set}"
log_message "DEBUG" "- CRON_SCHEDULE: ${CRON_SCHEDULE:-Not set}"
log_message "DEBUG" "- EXPORT_OPTION: ${EXPORT_OPTION:-Not set}"
log_message "DEBUG" "- LIMIT_FILMS: ${LIMIT_FILMS:-Not set}"

# Create config directory if it doesn't exist
mkdir -p /app/config

# Create example config file if it doesn't exist
if [ ! -f /app/config/.config.cfg.example ]; then
    echo "Creating example config file in config directory..."
    cat > /app/config/.config.cfg.example << 'EOF'
############################################################################
# TRAKT API CONFIGURATION
############################################################################
# API credentials - Get these from https://trakt.tv/oauth/applications
API_KEY="YOUR_API_KEY_HERE"
API_SECRET="YOUR_API_SECRET_HERE"
API_URL="https://api.trakt.tv"

# Authentication tokens - Generated by setup_trakt.sh
ACCESS_TOKEN=""
REFRESH_TOKEN=""
REDIRECT_URI="urn:ietf:wg:oauth:2.0:oob"

# User information
USERNAME="YOUR_TRAKT_USERNAME"

############################################################################
# DIRECTORY PATHS
############################################################################
# Backup and output directories
BACKUP_DIR="./backup"
DOSLOG="./logs"
DOSCOPY="./copy"
CONFIG_DIR="./config"

# Date format for filenames
DATE=$(date +%Y%m%d_%H%M)
LOG="${DOSLOG}/Export_Trakt_4_Letterboxd_$(date '+%Y-%m-%d_%H-%M-%S').log"

############################################################################
# DISPLAY SETTINGS
############################################################################
# Terminal colors
RED='\033[0;31m'     # Color code for error messages
GREEN='\033[0;32m'   # Color code for success messages
NC='\033[0m'         # No Color 
BOLD='\033[1m'       # Code for bold text
SAISPAS='\e[1;33;41m' # Background color code: 1;33 for yellow, 44 for red
EOF
    echo "Example config file created at /app/config/.config.cfg.example"
fi

# Check if config file exists
if [ ! -f /app/config/.config.cfg ]; then
    echo "Config file not found. Creating from template..."
    cp /app/config/.config.cfg.example /app/config/.config.cfg
    echo "Please edit /app/config/.config.cfg with your Trakt API credentials."
fi

# Function to verify and add missing variables to the config file
verify_config_variables() {
    local config_file="/app/config/.config.cfg"
    local example_file="/app/config/.config.cfg.example"
    local missing_vars=0
    local added_vars=0
    
    log_message "INFO" "Verifying configuration variables..."
    
    # Create a temporary file to store the list of required variables
    cat > /tmp/required_vars.txt << 'EOF'
API_KEY
API_SECRET
API_URL
ACCESS_TOKEN
REFRESH_TOKEN
REDIRECT_URI
USERNAME
BACKUP_DIR
DOSLOG
DOSCOPY
CONFIG_DIR
DATE
LOG
RED
GREEN
NC
BOLD
SAISPAS
EOF
    
    # Check each required variable
    while IFS= read -r var; do
        if ! grep -q "^${var}=" "$config_file"; then
            log_message "WARN" "Missing variable: ${var}"
            missing_vars=$((missing_vars + 1))
            
            # Extract the variable definition from the example file
            var_line=$(grep "^${var}=" "$example_file")
            
            if [ -n "$var_line" ]; then
                # Add the variable to the config file
                echo "$var_line" >> "$config_file"
                added_vars=$((added_vars + 1))
                log_message "INFO" "Added ${var} to config file"
            else
                log_message "ERROR" "Could not find ${var} in example file"
            fi
        fi
    done < /tmp/required_vars.txt
    
    # Clean up temporary files
    rm -f /tmp/required_vars.txt
    
    # Report results
    if [ $missing_vars -eq 0 ]; then
        log_message "SUCCESS" "All required variables are present in the config file."
    else
        if [ $added_vars -eq $missing_vars ]; then
            log_message "SUCCESS" "Added $added_vars missing variables to the config file."
        else
            log_message "WARN" "Found $missing_vars missing variables, but could only add $added_vars."
            log_message "WARN" "Please check your config file manually."
        fi
    fi
}

# Remove any existing symlink or config file in the root directory
if [ -L /app/.config.cfg ] || [ -f /app/.config.cfg ]; then
    log_message "INFO" "Removing old config file from root directory"
    rm -f /app/.config.cfg
    log_message "SUCCESS" "Removed old config file from root directory"
fi

# Create necessary directories with proper permissions
log_message "INFO" "Creating necessary directories with proper permissions"
mkdir -p /app/logs /app/copy /app/backup /app/TEMP
chmod -R 777 /app/logs /app/copy /app/backup /app/TEMP /app/config
log_message "SUCCESS" "Directories created with permissions 777"

# Debug directory information
debug_file_info "/app/logs" "Logs directory"
debug_file_info "/app/copy" "Copy directory"
debug_file_info "/app/backup" "Backup directory"
debug_file_info "/app/TEMP" "Temp directory"
debug_file_info "/app/config" "Config directory"
debug_file_info "/app/lib" "Library directory"

# Verify config file variables
verify_config_variables

# Check if Trakt API credentials are set
if grep -q '^API_KEY="YOUR_API_KEY_HERE"' /app/config/.config.cfg || \
   grep -q '^API_SECRET="YOUR_API_SECRET_HERE"' /app/config/.config.cfg; then
    log_message "WARN" "API credentials not configured in .config.cfg"
    log_message "INFO" "Please edit /app/config/.config.cfg with your Trakt API credentials"
    log_message "INFO" "You can get API credentials at https://trakt.tv/oauth/applications"
fi

# Main entry point - Add this at the end of the file
if [ "$1" = "healthcheck" ]; then
    # Just run the health check and exit
    source /app/lib/health_check.sh
    run_health_checks
    exit $?
elif [ "$1" = "setup" ]; then
    # Run the setup script
    exec /app/setup_trakt.sh
else
    # Clean temporary directories before starting
    clean_temp_directories
    
    # Sync environment variables to config file
    sync_env_to_config
    
    # Start health check server in background
    start_health_server
    
    # Run the export script based on cron schedule or directly
    if [ -n "$CRON_SCHEDULE" ]; then
        log_message "INFO" "Setting up scheduled task with cron schedule: $CRON_SCHEDULE"
        
        # Function to calculate seconds until next cron execution
        calculate_next_run() {
            local cron_schedule="$1"
            local cron_min cron_hour cron_day cron_month cron_dow
            
            # Parse the cron schedule
            read -r cron_min cron_hour cron_day cron_month cron_dow <<< "$cron_schedule"
            
            # Redirect debug logs to a file instead of stdout to avoid interference with return value
            log_message "DEBUG" "Parsed cron: minute=$cron_min, hour=$cron_hour, day=$cron_day, month=$cron_month, dow=$cron_dow" > /app/logs/cron_parser.log 2>&1
            
            # Special case for * * * * * (every minute)
            if [ "$cron_min" = "*" ] && [ "$cron_hour" = "*" ] && [ "$cron_day" = "*" ] && [ "$cron_month" = "*" ]; then
                log_message "INFO" "Schedule format detected: every minute" >> /app/logs/cron_parser.log 2>&1
                echo "60"
                return
            fi
            
            # Special case for */X format (every X minutes)
            if [[ "$cron_min" =~ ^\*/([0-9]+)$ ]]; then
                local minutes="${BASH_REMATCH[1]}"
                log_message "INFO" "Schedule format detected: every $minutes minutes" >> /app/logs/cron_parser.log 2>&1
                echo "$((minutes * 60))"
                return
            fi
            
            # Handle daily schedule at specific time (e.g., "0 3 * * *" = every day at 3am)
            if [[ "$cron_min" =~ ^[0-9]+$ ]] && [[ "$cron_hour" =~ ^[0-9]+$ ]] && [ "$cron_day" = "*" ] && [ "$cron_month" = "*" ]; then
                log_message "INFO" "Schedule format detected: daily at $cron_hour:$cron_min" >> /app/logs/cron_parser.log 2>&1
                
                # Get current hour and minute
                local current_hour=$(date +%H)
                local current_min=$(date +%M)
                
                # Convert everything to minutes since midnight
                local schedule_minutes=$((cron_hour * 60 + cron_min))
                local current_minutes=$((current_hour * 60 + current_min))
                
                local wait_minutes=0
                
                # If scheduled time is in the future today
                if [ $schedule_minutes -gt $current_minutes ]; then
                    wait_minutes=$((schedule_minutes - current_minutes))
                else
                    # Schedule is for tomorrow
                    wait_minutes=$((schedule_minutes + 1440 - current_minutes))
                fi
                
                log_message "INFO" "Next run in $wait_minutes minutes" >> /app/logs/cron_parser.log 2>&1
                echo "$((wait_minutes * 60))"
                return
            fi
            
            # For unsupported formats, default to hourly
            log_message "WARN" "Complex cron format not fully supported, using hourly schedule" >> /app/logs/cron_parser.log 2>&1
            echo "3600"
        }
        
        # Run in a loop with the cron schedule
        while true; do
            # Get current timestamp for log filename
            TIMESTAMP=$(date '+%Y-%m-%d_%H-%M-%S')
            LOG_FILE="/app/logs/cron_export_${TIMESTAMP}.log"
            
            # Create a more visual log header
            log_message "INFO" "🔄 ==================== SCHEDULED EXPORT START ===================="
            log_message "INFO" "📅 Date: $(date '+%Y-%m-%d') ⏰ Time: $(date '+%H:%M:%S')"
            log_message "INFO" "📋 Export option: ${EXPORT_OPTION}"
            log_message "INFO" "📝 Log file: ${LOG_FILE}"
            log_message "INFO" "🔄 =========================================================="
            
            # Run the export script and log output
            log_message "INFO" "🚀 Launching export script..."
            START_TIME=$(date +%s)
            
            # Pass the LIMIT_FILMS parameter if set
            if [ -n "$LIMIT_FILMS" ]; then
                log_message "INFO" "🎯 Limiting to ${LIMIT_FILMS} films"
                export LIMIT_FILMS
            fi
            
            /app/Export_Trakt_4_Letterboxd.sh $EXPORT_OPTION > "$LOG_FILE" 2>&1
            EXIT_CODE=$?
            END_TIME=$(date +%s)
            DURATION=$((END_TIME - START_TIME))
            
            # Count films in generated CSV files
            count_films_in_csv() {
                # Check each type of CSV file that might be generated
                local total_films=0
                local csv_found=false
                local report=""
                local raw_count=0
                
                # Define all possible CSV files to check
                local csv_files=(
                    "/app/copy/letterboxd_import.csv"
                    "/app/copy/trakt_movies_history.csv"
                    "/app/copy/trakt_movies_watched.csv"
                    "/app/copy/trakt_movies_watchlist.csv"
                    "/app/copy/trakt_ratings.csv"
                )
                
                # Check for raw output file
                if [ -f "/app/TEMP/watched_raw_output.csv" ]; then
                    raw_count=$(($(wc -l < "/app/TEMP/watched_raw_output.csv") - 1))
                    log_message "INFO" "📊 Raw Trakt data: ${raw_count} films"
                fi
                
                # Check other temp files to diagnose potential issues
                local temp_files=(
                    "/app/TEMP/watched_filtered_output.csv"
                    "/app/TEMP/ratings_output.csv"
                    "/app/TEMP/watchlist_output.csv"
                    "/app/TEMP/shows_watched.csv"
                )
                
                for temp_file in "${temp_files[@]}"; do
                    if [ -f "$temp_file" ]; then
                        local temp_count=$(($(wc -l < "$temp_file") - 1))
                        local temp_name=$(basename "$temp_file")
                        log_message "INFO" "📁 Temp file ${temp_name}: ${temp_count} entries"
                    fi
                done
                
                # Check each file and count lines (minus header)
                for csv_file in "${csv_files[@]}"; do
                    if [ -f "$csv_file" ]; then
                        csv_found=true
                        # Get line count minus header line
                        local count=$(($(wc -l < "$csv_file") - 1))
                        local filename=$(basename "$csv_file")
                        
                        # Add to total and report
                        total_films=$((total_films + count))
                        
                        # Check if this is letterboxd_import and raw_count exists
                        if [[ "$filename" == "letterboxd_import.csv" ]] && [[ $raw_count -gt 0 ]]; then
                            local diff=$((raw_count - count))
                            if [ $diff -gt 0 ]; then
                                report="${report}📊 ${filename}: ${count} films (${diff} films missing from raw data)\n"
                            else
                                report="${report}📊 ${filename}: ${count} films\n"
                            fi
                        else
                            report="${report}📊 ${filename}: ${count} films\n"
                        fi
                    fi
                done
                
                if [ "$csv_found" = true ]; then
                    log_message "INFO" "📋 Export summary:"
                    # Print report without trailing newline
                    echo -e "$report" | sed '/^$/d' | while IFS= read -r line; do
                        log_message "INFO" "$line"
                    done
                    log_message "INFO" "📊 Total: $total_films films in export files"
                else
                    log_message "WARN" "⚠️ No CSV files found to count films"
                fi
            }
            
            # Log completion with duration and status
            if [ $EXIT_CODE -eq 0 ]; then
                log_message "SUCCESS" "✅ Export completed successfully in ${DURATION}s"
                # Count and log film numbers
                count_films_in_csv
            else
                log_message "ERROR" "❌ Export failed with exit code ${EXIT_CODE} after ${DURATION}s"
            fi
            
            # Calculate time until next run - capture the output in a variable
            SLEEP_SECONDS=$(calculate_next_run "$CRON_SCHEDULE")
            
            # Calculer la date/heure de la prochaine exécution de manière compatible avec Alpine Linux (BusyBox)
            NEXT_RUN_TIME=$(date -d "@$(($(date +%s) + SLEEP_SECONDS))" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || date "+%Y-%m-%d %H:%M:%S" -d "@$(($(date +%s) + SLEEP_SECONDS))" 2>/dev/null || echo "dans ${SLEEP_SECONDS}s")
            
            # Log details about timing with visual elements
            if [ "$SLEEP_SECONDS" -ge 3600 ]; then
                HOURS=$((SLEEP_SECONDS / 3600))
                MINUTES=$(((SLEEP_SECONDS % 3600) / 60))
                log_message "INFO" "⏱️ Next run in ${HOURS}h ${MINUTES}m (at ${NEXT_RUN_TIME})"
            else
                MINUTES=$((SLEEP_SECONDS / 60))
                log_message "INFO" "⏱️ Next run in ${MINUTES} minutes (at ${NEXT_RUN_TIME})"
            fi
            log_message "INFO" "🔄 ==================== SCHEDULED EXPORT END ====================="
            
            # Wait for the next interval
            sleep $SLEEP_SECONDS
        done
    else
        # Run script once
        log_message "INFO" "Running export script once with option: $EXPORT_OPTION"
        exec /app/Export_Trakt_4_Letterboxd.sh "$EXPORT_OPTION"
    fi
fi 