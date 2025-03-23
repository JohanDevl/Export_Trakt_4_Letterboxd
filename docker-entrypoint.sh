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

# Initial system information
log_message "INFO" "Starting Docker container for Export_Trakt_4_Letterboxd"
log_message "DEBUG" "Container environment:"
log_message "DEBUG" "User: $(id)"
log_message "DEBUG" "Working directory: $(pwd)"
log_message "DEBUG" "Environment variables:"
log_message "DEBUG" "- TZ: ${TZ:-Not set}"
log_message "DEBUG" "- CRON_SCHEDULE: ${CRON_SCHEDULE:-Not set}"
log_message "DEBUG" "- EXPORT_OPTION: ${EXPORT_OPTION:-Not set}"

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

# Install cron job
install_cron_job() {
    log_message "INFO" "Installing cron job with schedule: $CRON_SCHEDULE"
    
    # Create wrapper script for cron with improved logging
    cat > /app/cron_wrapper.sh << 'EOF'
#!/bin/bash
START_TIME=$(date +"%Y-%m-%d %H:%M:%S")
EXPORT_OPTION=${1:-normal}

# Log to container stdout with friendly messages
echo ""
echo "======================================================================"
echo "🎬 [CRON] Starting Trakt to Letterboxd Export at ${START_TIME} 🎬"
echo "📊 Exporting your Trakt data with option '${EXPORT_OPTION}'..."
echo "======================================================================"

# Run the export script and redirect output to log file
/app/Export_Trakt_4_Letterboxd.sh ${EXPORT_OPTION} > /app/logs/cron_export.log 2>&1
EXIT_CODE=$?

END_TIME=$(date +"%Y-%m-%d %H:%M:%S")

# Check result and show friendly message
if [ $EXIT_CODE -eq 0 ]; then
    echo "======================================================================"
    echo "✅ [CRON] Export completed successfully at ${END_TIME}"
    echo "🎉 Your Letterboxd import file is ready in the copy directory!"
    
    # Show CSV file info if it exists
    if [ -f /app/copy/letterboxd_import.csv ]; then
        MOVIES_COUNT=$(wc -l < /app/copy/letterboxd_import.csv)
        MOVIES_COUNT=$((MOVIES_COUNT - 1))  # Subtract header row
        echo "📋 Exported ${MOVIES_COUNT} movies to CSV file"
        echo "📂 File: /app/copy/letterboxd_import.csv ($(du -h /app/copy/letterboxd_import.csv | cut -f1) size)"
    else
        echo "⚠️ No CSV file was created. Check the logs for errors."
    fi
    echo "======================================================================"
    echo ""
else
    echo "======================================================================"
    echo "❌ [CRON] Export failed with exit code ${EXIT_CODE} at ${END_TIME}"
    echo "⚠️ Please check the logs for errors: /app/logs/cron_export.log"
    echo "======================================================================"
    echo ""
fi
EOF

    # Make wrapper script executable
    chmod +x /app/cron_wrapper.sh
    
    # Ensure cron.d directory exists
    mkdir -p /etc/cron.d
    
    # Create cron job file using the wrapper script
    cat > /etc/crontab << EOF
# Trakt Export Cron Job
$CRON_SCHEDULE /app/cron_wrapper.sh ${EXPORT_OPTION:-normal}
# Empty line required at the end

EOF
    
    # Install the cron job
    chmod 0644 /etc/crontab
    crontab /etc/crontab
    
    log_message "SUCCESS" "Cron job installed with friendly logging"
}

# Check for CRON_SCHEDULE environment variable
if [ -n "$CRON_SCHEDULE" ]; then
    log_message "INFO" "Setting up cron job with schedule: $CRON_SCHEDULE"
    
    # Install cron job
    install_cron_job
    
    # Start crond in the foreground
    log_message "INFO" "Starting crond in the foreground"
    exec crond -f
else
    log_message "INFO" "No cron schedule specified. Running export script once with option: ${EXPORT_OPTION:-normal}"
    
    # Run the export script once
    exec /app/Export_Trakt_4_Letterboxd.sh "${EXPORT_OPTION:-normal}"
fi 