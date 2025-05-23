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
    
    # Create directory if it doesn't exist
    if [ ! -d "$temp_dir" ]; then
        mkdir -p "$temp_dir"
        echo "Created temporary directory: $temp_dir" | tee -a "${log_file}"
    else
        # Only remove the contents, not the directory itself
        echo "Cleaning temporary directory: $temp_dir" | tee -a "${log_file}"
        find "$temp_dir" -mindepth 1 -delete 2>/dev/null || {
            # If find fails, try a more aggressive approach
            chmod -R 777 "$temp_dir" 2>/dev/null
            find "$temp_dir" -mindepth 1 -delete 2>/dev/null || echo "$(_ "warning"): Could not clean temporary directory completely" | tee -a "${log_file}"
        }
    fi
    
    # Ensure directory has proper permissions
    chmod -R 777 "$temp_dir" 2>/dev/null || echo "$(_ "warning"): Could not set permissions on temporary directory" | tee -a "${log_file}"
    echo "Temporary directory ready: $temp_dir" | tee -a "${log_file}"
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
        echo "$(_ "directory_exists"): ✅" | tee -a "${log_file}"
        echo "$(_ "directory_permissions"): $(ls -la "$copy_dir" | head -n 1 | awk '{print $1}')" | tee -a "${log_file}"
    else
        echo "$(_ "directory_exists"): ❌ (will attempt to create)" | tee -a "${log_file}"
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
    
    echo "🌍 $(_ "environment_info"):" | tee -a "${log_file}"
    echo "  - $(_ "user"): $(whoami)" | tee -a "${log_file}"
    echo "  - $(_ "working_directory"): $(pwd)" | tee -a "${log_file}"
    echo "  - $(_ "script_directory"): $script_dir" | tee -a "${log_file}"
    echo "  - $(_ "copy_directory"): $copy_dir" | tee -a "${log_file}"
    echo "  - $(_ "log_directory"): $log_dir" | tee -a "${log_file}"
    echo "  - $(_ "backup_directory"): $backup_dir" | tee -a "${log_file}"
    echo "  - $(_ "os_type"): $OSTYPE" | tee -a "${log_file}"
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
    local log_file="$2"
    
    # Create backup folder if it doesn't exist
    if [ ! -d "${backup_dir}" ]; then
        mkdir -p "${backup_dir}"
        echo "$(_ "created_backup_directory"): ${backup_dir}" | tee -a "${log_file}"
    else
        echo "$(_ "backup_directory_exists"): ${backup_dir}" | tee -a "${log_file}"
    fi
    
    # Check permissions
    if [ -w "${backup_dir}" ]; then
        echo "$(_ "backup_directory_writable"): ✅" | tee -a "${log_file}"
    else
        echo "$(_ "backup_directory_not_writable")" | tee -a "${log_file}"
    fi
    
    # Return the backup directory for use
    echo "${backup_dir}"
} 