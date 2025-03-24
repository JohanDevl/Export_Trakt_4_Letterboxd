#!/bin/bash
#
# Main module - Orchestrates the Trakt to Letterboxd export process
#

# Import all required modules
import_modules() {
    local script_dir="$1"
    
    # List of modules to import
    modules=("config" "utils" "i18n" "trakt_api" "data_processing")
    
    for module in "${modules[@]}"; do
        # First try the local path
        if [ -f "${script_dir}/lib/${module}.sh" ]; then
            source "${script_dir}/lib/${module}.sh"
        # Then try the Docker path
        elif [ -f "/app/lib/${module}.sh" ]; then
            source "/app/lib/${module}.sh"
        else
            echo "ERROR: Required module not found: ${module}.sh"
            exit 1
        fi
    done
}

# Initialize script environment
initialize_environment() {
    local script_dir="$1"
    local option="$2"
    local log="$3"
    
    # Print debug information
    echo "=========== DEBUG INFORMATION ==========="
    echo "Script called with option: $option"
    echo "Number of arguments: $#"
    if [ -n "$option" ]; then
        echo "Option value: '$option'"
    else
        echo "No option provided, using default"
    fi
    echo "========================================="
    
    # Get sed command based on OS
    SED_INPLACE=$(detect_os_sed "$log")
    
    # Load configuration
    load_config "$script_dir" "$log"
    
    # Initialize internationalization
    init_i18n "$script_dir" "$log"
    
    # Initialize temporary directory
    init_temp_dir "$TEMP_DIR" "$log"
    
    # Ensure required directories exist
    ensure_directories "$DOSLOG" "$DOSCOPY" "$log"
    
    # Log environment information
    log_environment "$log" "$script_dir" "$DOSCOPY" "$DOSLOG" "$BACKUP_DIR"
    
    # Initialize backup directory
    init_backup_dir "$BACKUP_DIR" "$log"
    
    # Check for existing CSV file
    if [ -f "${DOSCOPY}/letterboxd_import.csv" ]; then
        debug_file_info "${DOSCOPY}/letterboxd_import.csv" "$(_  "existing_csv_check")" "$log"
    fi
    
    # Check for required dependencies
    check_dependencies "$log" || exit 1
    
    echo -e "$(_ "retrieving_info")" | tee -a "${log}"
}

# Process command line arguments
process_arguments() {
    local script_dir="$1"
    local log="$2"
    local option="$3"
    
    # Initialize environment with command line option
    initialize_environment "$script_dir" "$option" "$log"
    
    # Process based on option
    case "$option" in
        "help"|"-h"|"--help")
            show_help "$log"
            exit 0
            ;;
        "normal"|"")
            # Default option - normal export
            export_trakt_history "$log" "$DOSCOPY" "$script_dir"
            ;;
        "update")
            # Update export with new entries
            echo -e "$(_ "update_export")" | tee -a "${log}"
            update_export "$log" "$DOSCOPY" "$script_dir"
            ;;
        "backup")
            # Backup Trakt data
            echo -e "$(_ "backup_data")" | tee -a "${log}"
            backup_trakt_data "$log" "$BACKUP_DIR" "$script_dir"
            ;;
        "setup")
            # Setup Trakt API
            echo -e "$(_ "setup_api")" | tee -a "${log}"
            setup_trakt_api "$log" "$script_dir"
            ;;
        "clean")
            # Clean temporary files
            echo -e "$(_ "cleaning_files")" | tee -a "${log}"
            clean_temp_files "$log" "$TEMP_DIR" "$script_dir"
            ;;
        *)
            echo -e "$(_ "unknown_option") $option" | tee -a "${log}"
            show_help "$log"
            exit 1
            ;;
    esac
    
    echo -e "$(_ "script_complete")" | tee -a "${log}"
}

# Show help information
show_help() {
    local log="$1"
    
    echo "$(_ "usage"): ./Export_Trakt_4_Letterboxd.sh [option]" | tee -a "${log}"
    echo "" | tee -a "${log}"
    echo "$(_ "options"):" | tee -a "${log}"
    echo "  help        $(_ "show_help")" | tee -a "${log}"
    echo "  normal      $(_ "normal_export") ($(_ "default"))" | tee -a "${log}"
    echo "  update      $(_ "update_export")" | tee -a "${log}"
    echo "  backup      $(_ "backup_data")" | tee -a "${log}"
    echo "  setup       $(_ "setup_api")" | tee -a "${log}"
    echo "  clean       $(_ "cleaning_files")" | tee -a "${log}"
}

# Export Trakt history to CSV
export_trakt_history() {
    local log="$1"
    local output_dir="$2"
    local script_dir="$3"
    
    echo -e "$(_ "exporting_history")" | tee -a "${log}"
    
    # Fetch watch history from Trakt API
    fetch_watched_history "$log" "$TEMP_DIR" "$script_dir"
    
    # Process the data into CSV format
    process_data_to_csv "$log" "$TEMP_DIR" "$output_dir" "$script_dir"
    
    echo -e "$(_ "export_complete")" | tee -a "${log}"
    echo -e "$(_ "output_location"): ${output_dir}/letterboxd_import.csv" | tee -a "${log}"
}

# Update export with new entries
update_export() {
    local log="$1"
    local output_dir="$2"
    local script_dir="$3"
    
    echo -e "$(_ "updating_export")" | tee -a "${log}"
    
    # Check if previous export exists
    if [ ! -f "${output_dir}/letterboxd_import.csv" ]; then
        echo -e "$(_ "no_previous_export")" | tee -a "${log}"
        echo -e "$(_ "running_full_export")" | tee -a "${log}"
        export_trakt_history "$log" "$output_dir" "$script_dir"
        return
    fi
    
    # Backup existing export
    local backup_file="${output_dir}/letterboxd_import_${DATE}.csv.bak"
    cp "${output_dir}/letterboxd_import.csv" "$backup_file"
    echo -e "$(_ "previous_export_backed_up"): $backup_file" | tee -a "${log}"
    
    # Fetch new history since last export
    fetch_history_since_last_export "$log" "$TEMP_DIR" "$output_dir" "$script_dir"
    
    # Merge with existing export
    merge_with_existing_export "$log" "$TEMP_DIR" "$output_dir" "$script_dir"
    
    echo -e "$(_ "update_complete")" | tee -a "${log}"
    echo -e "$(_ "output_location"): ${output_dir}/letterboxd_import.csv" | tee -a "${log}"
}

# Backup Trakt data
backup_trakt_data() {
    local log="$1"
    local backup_dir="$2"
    local script_dir="$3"
    
    echo -e "$(_ "backup_started")" | tee -a "${log}"
    
    # Ensure backup directory exists
    mkdir -p "$backup_dir"
    
    # Backup different data types
    backup_watched_history "$log" "$backup_dir" "$script_dir"
    backup_ratings "$log" "$backup_dir" "$script_dir"
    backup_watchlist "$log" "$backup_dir" "$script_dir"
    backup_lists "$log" "$backup_dir" "$script_dir"
    
    echo -e "$(_ "backup_complete")" | tee -a "${log}"
    echo -e "$(_ "backup_location"): $backup_dir" | tee -a "${log}"
}

# Setup Trakt API
setup_trakt_api() {
    local log="$1"
    local script_dir="$2"
    
    echo -e "$(_ "setup_started")" | tee -a "${log}"
    
    # Run the setup script
    "$script_dir/setup_trakt.sh"
    
    echo -e "$(_ "setup_complete")" | tee -a "${log}"
}

# Clean temporary files
clean_temp_files() {
    local log="$1"
    local temp_dir="$2"
    local script_dir="$3"
    
    echo -e "$(_ "cleaning_started")" | tee -a "${log}"
    
    # Clean temporary directory
    if [ -d "$temp_dir" ]; then
        rm -rf "${temp_dir:?}/"* 2>/dev/null
        echo -e "$(_ "temp_dir_cleaned"): $temp_dir" | tee -a "${log}"
    else
        echo -e "$(_ "temp_dir_not_found"): $temp_dir" | tee -a "${log}"
    fi
    
    # Clean old log files (older than 30 days)
    if [ -d "$DOSLOG" ]; then
        find "$DOSLOG" -name "*.log" -type f -mtime +30 -delete 2>/dev/null
        echo -e "$(_ "old_logs_cleaned")" | tee -a "${log}"
    fi
    
    echo -e "$(_ "cleaning_complete")" | tee -a "${log}"
}