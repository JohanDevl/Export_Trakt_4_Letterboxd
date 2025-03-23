#!/bin/bash
#
# Main module - Orchestrates the Trakt to Letterboxd export process
#

# Import all required modules
import_modules() {
    local script_dir="$1"
    
    # List of modules to import
    modules=("config" "utils" "trakt_api" "data_processing")
    
    for module in "${modules[@]}"; do
        if [ -f "${script_dir}/lib/${module}.sh" ]; then
            source "${script_dir}/lib/${module}.sh"
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
    
    # Initialize temporary directory
    init_temp_dir "$TEMP_DIR" "$log"
    
    # Ensure required directories exist
    ensure_directories "$DOSLOG" "$DOSCOPY" "$log"
    
    # Log environment information
    log_environment "$log" "$script_dir" "$DOSCOPY" "$DOSLOG" "$BACKUP_DIR"
    
    # Initialize backup directory
    init_backup_dir "$BACKUP_DIR"
    
    # Check for existing CSV file
    if [ -f "${DOSCOPY}/letterboxd_import.csv" ]; then
        debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Existing CSV file check" "$log"
    fi
    
    # Check for required dependencies
    check_dependencies "$log" || exit 1
    
    echo -e "Retrieving information..." | tee -a "${log}"
}

# Process command line arguments
process_arguments() {
    local arg="$1"
    local log="$2"
    
    if [ ! -z "$arg" ]; then
        local option=$(echo "$arg" | tr '[:upper:]' '[:lower:]')
        echo -e "${SAISPAS}${BOLD}[$(date)] - Processing option: $option${NC}" | tee -a "${log}"
        echo "$option"
    else
        echo -e "${SAISPAS}${BOLD}[$(date)] - No option provided, using default${NC}" | tee -a "${log}"
        echo "normal"
    fi
}

# Fetch data from Trakt
fetch_all_data() {
    local api_url="$1"
    local api_key="$2"
    local api_secret="$3"
    local access_token="$4"
    local refresh_token="$5"
    local redirect_uri="$6"
    local username="$7"
    local option="$8"
    local backup_dir="$9"
    local config_file="${10}"
    local sed_inplace="${11}"
    local log="${12}"
    
    # Get endpoints based on mode
    local endpoints_string=$(get_endpoints_for_mode "$option" "$log")
    IFS=' ' read -r -a endpoints <<< "$endpoints_string"
    
    # Check token validity before proceeding
    if ! check_token_validity "$api_url" "$api_key" "$access_token" "$log"; then
        echo "⚠️ Token expired, attempting to refresh..." | tee -a "${log}"
        local new_tokens=$(refresh_access_token "$refresh_token" "$api_key" "$api_secret" "$redirect_uri" "$config_file" "$sed_inplace" "$log")
        if [ $? -eq 0 ]; then
            IFS=':' read -r new_access_token new_refresh_token <<< "$new_tokens"
            access_token="$new_access_token"
            refresh_token="$new_refresh_token"
        else
            echo "❌ Failed to refresh token. Exiting." | tee -a "${log}"
            exit 1
        fi
    fi
    
    # Fetch data for each endpoint
    local success=0
    local total=${#endpoints[@]}
    local current=0
    
    for endpoint in "${endpoints[@]}"; do
        current=$((current + 1))
        progress_bar $current $total "Fetching Trakt data" "$log"
        
        local filename="${username}-${endpoint//\//_}.json"
        local output_file="${backup_dir}/${filename}"
        
        fetch_trakt_data "$api_url" "$api_key" "$access_token" "$endpoint" "$output_file" "$username" "$log"
        success=$((success + $?))
    done
    
    echo -e "All files have been retrieved\n Starting processing" | tee -a "${log}"
    
    if [ $success -gt 0 ]; then
        echo "⚠️ Some fetches failed. Check the log for details." | tee -a "${log}"
    fi
}

# Process data and create Letterboxd CSV
process_data() {
    local option="$1"
    local username="$2"
    local backup_dir="$3"
    local temp_dir="$4"
    local doscopy="$5"
    local log="$6"
    
    # Create output directory if it doesn't exist
    mkdir -p "$doscopy"
    
    # Create empty CSV file with header
    echo "Title,Year,imdbID,tmdbID,WatchedDate,Rating10,Rewatch" > "${temp_dir}/movies_export.csv"
    
    # Create ratings lookup
    local ratings_file="${backup_dir}/${username}-ratings_movies.json"
    local ratings_lookup="${temp_dir}/ratings_lookup.json"
    create_ratings_lookup "$ratings_file" "$ratings_lookup" "$log"
    
    # Create plays count lookup
    local watched_file="${backup_dir}/${username}-watched_movies.json"
    local plays_lookup="${temp_dir}/plays_count_lookup.json"
    create_plays_count_lookup "$watched_file" "$plays_lookup" "$log"
    
    # Process history movies
    local history_file="${backup_dir}/${username}-history_movies.json"
    local raw_output="${temp_dir}/raw_output.csv"
    local csv_output="${temp_dir}/movies_export.csv"
    local history_processed=false
    
    if process_history_movies "$history_file" "$ratings_lookup" "$plays_lookup" "$csv_output" "$raw_output" "$log"; then
        history_processed=true
    fi
    
    # Process watched movies if in complete mode or if history processing failed
    if [ "$option" = "complete" ] && [ -f "$watched_file" ]; then
        local existing_ids="${temp_dir}/existing_imdb_ids.txt"
        local watched_raw="${temp_dir}/watched_raw_output.csv"
        process_watched_movies "$watched_file" "$ratings_lookup" "$csv_output" "$existing_ids" "$watched_raw" "false" "$log"
    elif [ "$history_processed" = "false" ] && [ -f "$watched_file" ]; then
        local existing_ids="${temp_dir}/existing_imdb_ids.txt"
        local watched_raw="${temp_dir}/watched_raw_output.csv"
        process_watched_movies "$watched_file" "$ratings_lookup" "$csv_output" "$existing_ids" "$watched_raw" "true" "$log"
    else
        if [ "$history_processed" = "false" ]; then
            echo -e "Movies history: No movies found in history or watched" | tee -a "${log}"
        fi
    fi
    
    # Copy file to final destination
    cp "${temp_dir}/movies_export.csv" "${doscopy}/letterboxd_import.csv"
    debug_msg "CSV file created in ${doscopy}/letterboxd_import.csv" "$log"
    
    # Create backup if in complete mode
    if [ "$option" = "complete" ]; then
        create_backup_archive "$backup_dir" "$log"
    fi
}

# Main function - entry point for the script
run_export() {
    local script_dir="$1"
    local option="$2"
    
    # Import all modules
    import_modules "$script_dir"
    
    # Parse mode from option
    mode=$(process_arguments "$option" "${LOG}")
    
    # Initialize script environment
    initialize_environment "$script_dir" "$option" "${LOG}"
    
    # Fetch all data from Trakt
    fetch_all_data "$API_URL" "$API_KEY" "$API_SECRET" "$ACCESS_TOKEN" "$REFRESH_TOKEN" "$REDIRECT_URI" "$USERNAME" "$mode" "$BACKUP_DIR" "${CONFIG_DIR}/.config.cfg" "$SED_INPLACE" "${LOG}"
    
    # Process data and create Letterboxd CSV
    process_data "$mode" "$USERNAME" "$BACKUP_DIR" "$TEMP_DIR" "$DOSCOPY" "${LOG}"
    
    echo "Export process completed. CSV file is ready for Letterboxd import." | tee -a "${LOG}"
} 