#!/bin/bash
# 
# Configuration management functions
#

# Load configuration from appropriate location
load_config() {
    local script_dir="$1"
    local config_dir="${script_dir}/config"
    local log_file="$2"
    
    # Determine config file path based on environment
    if [ -f "/app/config/.config.cfg" ]; then
        # If running in Docker, use the absolute path
        source /app/config/.config.cfg
        echo "Using Docker config file: /app/config/.config.cfg" | tee -a "${log_file}"
    else
        # If running locally, use the relative path
        source "${config_dir}/.config.cfg"
        echo "Using local config file: ${config_dir}/.config.cfg" | tee -a "${log_file}"
    fi
}

# Initialize temporary directory
init_temp_dir() {
    local temp_dir="$1"
    local log_file="$2"
    
    # Use the user's temporary directory
    rm -rf "$temp_dir"
    mkdir -p "$temp_dir"
    echo "Created temporary directory: $temp_dir" | tee -a "${log_file}"
}

# Ensure required directories exist
ensure_directories() {
    local log_dir="$1"
    local copy_dir="$2"
    local log_file="$3"
    
    # Create log directory if needed
    if [ ! -d "$log_dir" ]; then
        mkdir -p "$log_dir"
        echo "Created log directory: $log_dir" | tee -a "${log_file}"
    fi
    
    # Check and create copy directory if needed
    if [ -d "$copy_dir" ]; then
        echo "Copy directory exists: ✅" | tee -a "${log_file}"
        echo "Copy directory permissions: $(ls -la "$copy_dir" | head -n 1 | awk '{print $1}')" | tee -a "${log_file}"
    else
        echo "Copy directory exists: ❌ (will attempt to create)" | tee -a "${log_file}"
        mkdir -p "$copy_dir"
    fi
}

# Log environment information
log_environment() {
    local log_file="$1"
    local script_dir="$2"
    local copy_dir="$3"
    local log_dir="$4"
    local backup_dir="$5"
    
    echo "🌍 Environment information:" | tee -a "${log_file}"
    echo "  - User: $(whoami)" | tee -a "${log_file}"
    echo "  - Working directory: $(pwd)" | tee -a "${log_file}"
    echo "  - Script directory: $script_dir" | tee -a "${log_file}"
    echo "  - Copy directory: $copy_dir" | tee -a "${log_file}"
    echo "  - Log directory: $log_dir" | tee -a "${log_file}"
    echo "  - Backup directory: $backup_dir" | tee -a "${log_file}"
    echo "  - OS Type: $OSTYPE" | tee -a "${log_file}"
    echo "-----------------------------------" | tee -a "${log_file}"
}

# Detect OS for sed compatibility
detect_os_sed() {
    local log_file="$1"
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS uses BSD sed
        echo "sed -i ''"
        echo "Detected macOS: Using BSD sed with empty string backup parameter" | tee -a "${log_file}"
    else
        # Linux and others use GNU sed
        echo "sed -i"
        echo "Detected Linux/other: Using GNU sed" | tee -a "${log_file}"
    fi
}

# Initialize backup directory
init_backup_dir() {
    local backup_dir="$1"
    
    # Create backup folder
    mkdir -p "${backup_dir}"
} 