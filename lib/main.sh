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
    init_backup_dir "$BACKUP_DIR" "$log"
    
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
        
        # Normalize mode to one of our supported values
        local mode
        case "$option" in
            "complete"|"full")
                mode="complete"
                ;;
            "initial"|"init")
                mode="initial"
                ;;
            *)
                mode="normal"
                ;;
        esac
        
        echo -e "Option '$option' translated to mode: $mode" | tee -a "${log}"
        echo "$mode"
    else
        echo -e "${SAISPAS}${BOLD}[$(date)] - No option provided, using default${NC}" | tee -a "${log}"
        echo "normal"
    fi
}

# Find the most recent backup directory
find_latest_backup_dir() {
    local base_dir="$1"
    local log_file="$2"
    
    # Find the most recent backup directory (based on modification time)
    local latest_dir=$(find "${base_dir}" -maxdepth 1 -type d -name "*_trakt-backup" -print0 | xargs -0 ls -td 2>/dev/null | head -n 1)
    
    if [ -z "$latest_dir" ]; then
        echo "⚠️ No backup directory found. Will create a new one." | tee -a "${log_file}"
        return 1
    else
        echo "📁 Found latest backup directory: $latest_dir" | tee -a "${log_file}"
        echo "$latest_dir"
        return 0
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
    
    # Debug the received mode
    echo "🔍 fetch_all_data received mode: '$option'" | tee -a "${log}"
    
    # Set the backup directory
    export CURRENT_BACKUP_DIR="$backup_dir"  # Export for use in other functions
    
    # Debug the backup directory
    echo "📁 Using backup directory: $backup_dir" | tee -a "${log}"
    
    # Ensure backup directory exists
    if [ ! -d "$backup_dir" ]; then
        mkdir -p "$backup_dir"
        echo "📁 Created backup directory: $backup_dir" | tee -a "${log}"
    fi
    
    # Get endpoints based on mode
    local endpoints=()
    
    echo -e "🔄 Processing with mode: '$option'" | tee -a "${log}"
    
    case "$option" in
        "complete")
            echo -e "📚 Complete Mode activated - Will fetch all data from Trakt" | tee -a "${log}"
            endpoints=(
                "watchlist/movies"
                "watchlist/shows"
                "watchlist/episodes"
                "watchlist/seasons"
                "ratings/movies"
                "ratings/shows"
                "ratings/episodes"
                "ratings/seasons"
                "collection/movies"
                "collection/shows"
                "watched/movies"
                "watched/shows"
                "history/movies"
                "history/shows"
                "history/episodes"
            )
            ;;
        "initial")
            echo -e "🚀 Initial Mode activated - Will fetch only essential movie data" | tee -a "${log}"
            endpoints=(
                "history/movies"
                "ratings/movies"
                "watched/movies"
            )
            ;;
        *)
            echo -e "🔄 Normal Mode activated - Will fetch standard movie and show data" | tee -a "${log}"
            endpoints=(
                "history/movies"
                "ratings/movies"
                "watched/movies"
                "watchlist/movies"
            )
            ;;
    esac
    
    echo "🔄 Endpoints to fetch: ${endpoints[*]}" | tee -a "${log}"
    
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
        
        echo "📥 Fetching endpoint: $endpoint" | tee -a "${log}"
        fetch_trakt_data "$api_url" "$api_key" "$access_token" "$endpoint" "$output_file" "$username" "$log"
        success=$((success + $?))
    done
    
    echo -e "🎉 Data fetching completed with $((total-success))/$total successful requests" | tee -a "${log}"
    echo -e "🔄 Starting processing phase" | tee -a "${log}"
    
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
    
    # Use the current backup directory (should already be set in run_export)
    if [ -n "$CURRENT_BACKUP_DIR" ]; then
        backup_dir="$CURRENT_BACKUP_DIR"
    fi
    
    echo "📊 Processing data from: $backup_dir" | tee -a "${log}"
    
    # Debug: List files in backup directory
    if [ -d "$backup_dir" ]; then
        echo "📄 Files in backup directory:" | tee -a "${log}"
        ls -la "$backup_dir" >> "$log" 2>&1 || echo "Cannot list directory" | tee -a "${log}"
    else
        echo "❌ Backup directory does not exist: $backup_dir" | tee -a "${log}"
        # Create backup directory as a fallback
        mkdir -p "$backup_dir"
        echo "Created backup directory as fallback" | tee -a "${log}"
    fi
    
    # Create output directory if it doesn't exist
    mkdir -p "$doscopy"
    
    # Check if the output file exists and remove it
    local final_output_file="${doscopy}/letterboxd_import.csv"
    if [ -f "$final_output_file" ]; then
        echo "🗑️ Removing existing Letterboxd import file: $final_output_file" | tee -a "${log}"
        rm -f "$final_output_file"
        
        # Verify removal was successful
        if [ -f "$final_output_file" ]; then
            echo "❌ ERROR: Failed to remove existing CSV file. Trying with force option." | tee -a "${log}"
            rm -f "$final_output_file"
            sleep 1
        fi
    fi
    
    # Also check alternative paths and remove any existing files
    local alt_paths=(
        "${SCRIPT_DIR}/copy/letterboxd_import.csv"
        "/app/copy/letterboxd_import.csv"
        "./copy/letterboxd_import.csv"
    )
    
    for alt_path in "${alt_paths[@]}"; do
        if [ -f "$alt_path" ] && [ "$alt_path" != "$final_output_file" ]; then
            echo "🗑️ Removing CSV file at alternate location: $alt_path" | tee -a "${log}"
            rm -f "$alt_path"
        fi
    done
    
    # Create empty CSV file with header
    echo "Title,Year,imdbID,tmdbID,WatchedDate,Rating10,Rewatch" > "${temp_dir}/movies_export.csv"
    
    # Create ratings lookup
    local ratings_file="${backup_dir}/${username}-ratings_movies.json"
    echo "📄 Looking for ratings file: $ratings_file" | tee -a "${log}"
    local ratings_lookup="${temp_dir}/ratings_lookup.json"
    create_ratings_lookup "$ratings_file" "$ratings_lookup" "$log"
    
    # Create plays count lookup
    local watched_file="${backup_dir}/${username}-watched_movies.json"
    echo "📄 Looking for watched file: $watched_file" | tee -a "${log}"
    local plays_lookup="${temp_dir}/plays_count_lookup.json"
    create_plays_count_lookup "$watched_file" "$plays_lookup" "$log"
    
    # Process history movies
    local history_file="${backup_dir}/${username}-history_movies.json"
    echo "📄 Looking for history file: $history_file" | tee -a "${log}"
    local raw_output="${temp_dir}/raw_output.csv"
    local csv_output="${temp_dir}/movies_export.csv"
    local history_processed=false
    
    if process_history_movies "$history_file" "$ratings_lookup" "$plays_lookup" "$csv_output" "$raw_output" "$log"; then
        history_processed=true
    fi
    
    # Process watched movies in all cases to ensure maximum data recovery
    if [ -f "$watched_file" ]; then
        local existing_ids="${temp_dir}/existing_imdb_ids.txt"
        local watched_raw="${temp_dir}/watched_raw_output.csv"
        
        # If history processing failed, use watched as primary source
        # Otherwise, supplement history with watched data
        if [ "$history_processed" = "false" ]; then
            echo "⚠️ No history data found, using watched data as primary source" | tee -a "${log}"
            process_watched_movies "$watched_file" "$ratings_lookup" "$csv_output" "$existing_ids" "$watched_raw" "true" "$log"
        else
            echo "📊 Supplementing history data with watched data" | tee -a "${log}"
            process_watched_movies "$watched_file" "$ratings_lookup" "$csv_output" "$existing_ids" "$watched_raw" "false" "$log"
        fi
    else
        if [ "$history_processed" = "false" ]; then
            echo -e "⚠️ Movies history: No movies found in history or watched" | tee -a "${log}"
        fi
    fi
    
    # Deduplicate the final CSV to ensure no duplicate entries
    local final_csv="${temp_dir}/final_export.csv"
    deduplicate_movies "$csv_output" "$final_csv" "$log"
    
    # Make sure the final CSV file was created
    if [ ! -f "$final_csv" ]; then
        echo "❌ ERROR: Final CSV file was not created in temp directory" | tee -a "${log}"
        return 1
    fi
    
    # Ensure target directory exists with proper permissions
    if [ ! -d "$doscopy" ]; then
        echo "📁 Creating copy directory: $doscopy" | tee -a "${log}"
        mkdir -p "$doscopy"
        chmod 777 "$doscopy"
    else
        echo "📁 Ensuring copy directory has proper permissions" | tee -a "${log}"
        chmod 777 "$doscopy"
    fi
    
    # Try to copy file to final destination with verbose output
    echo "📋 Copying final CSV to: ${doscopy}/letterboxd_import.csv" | tee -a "${log}"
    
    # Debug information for paths
    echo "🔍 DEBUG: Current directory: $(pwd)" | tee -a "${log}"
    echo "🔍 DEBUG: Source file: ${final_csv}" | tee -a "${log}"
    echo "🔍 DEBUG: Target directory: ${doscopy}" | tee -a "${log}"
    
    # Try to convert relative paths to absolute if needed
    if [[ "$doscopy" == "./"* ]]; then
        echo "🔍 Converting relative path to absolute path" | tee -a "${log}"
        local abs_doscopy="$(cd "$(dirname "$doscopy")" || exit; pwd)/$(basename "$doscopy")"
        echo "🔍 Absolute target directory: $abs_doscopy" | tee -a "${log}"
        
        # Create directory with full permissions if it doesn't exist
        if [ ! -d "$abs_doscopy" ]; then
            echo "📁 Creating absolute copy directory: $abs_doscopy" | tee -a "${log}"
            mkdir -p "$abs_doscopy"
            chmod 777 "$abs_doscopy"
        fi
        
        # Copy to absolute path instead
        cp -v "${final_csv}" "${abs_doscopy}/letterboxd_import.csv" 2>&1 | tee -a "${log}"
        
        # Also copy to the original path as fallback
        cp -v "${final_csv}" "${doscopy}/letterboxd_import.csv" 2>&1 | tee -a "${log}"
    else
        # Use original path
        cp -v "${final_csv}" "${doscopy}/letterboxd_import.csv" 2>&1 | tee -a "${log}"
    fi
    
    # Check both possible locations
    if [ -f "${doscopy}/letterboxd_import.csv" ]; then
        echo "✅ CSV file created in ${doscopy}/letterboxd_import.csv" | tee -a "${log}"
        local output_file="${doscopy}/letterboxd_import.csv"
    elif [[ "$doscopy" == "./"* ]] && [ -f "${abs_doscopy}/letterboxd_import.csv" ]; then
        echo "✅ CSV file created in ${abs_doscopy}/letterboxd_import.csv" | tee -a "${log}"
        local output_file="${abs_doscopy}/letterboxd_import.csv"
    else
        echo "❌ ERROR: Failed to create final CSV file" | tee -a "${log}"
        echo "🔍 DEBUG: Checking target directory permissions" | tee -a "${log}"
        ls -la "$doscopy" >> "${log}" 2>&1
        if [[ "$doscopy" == "./"* ]]; then
            ls -la "$abs_doscopy" >> "${log}" 2>&1
        fi
        return 1
    fi
    
    # Ensure the final file has the right permissions
    chmod 666 "${output_file}"
    
    # Get file stats for logging
    local file_size=$(wc -c < "${output_file}")
    local movie_count=$(($(wc -l < "${output_file}") - 1)) # Subtract header line
    
    echo "📊 Exported $movie_count movies (file size: $file_size bytes)" | tee -a "${log}"
    
    # Create backup if in complete mode
    if [ "$option" = "complete" ]; then
        create_backup_archive "$backup_dir" "$log"
    fi
    
    # Return success
    return 0
}

# Main function - entry point for the script
run_export() {
    local script_dir="$1"
    local option="$2"
    
    # Import all modules
    import_modules "$script_dir"
    
    # Parse mode from option - capture only the last line of output
    mode=$(process_arguments "$option" "${LOG}" | tail -n 1)
    echo "Debug: Mode after processing is: '$mode'" | tee -a "${LOG}"
    
    # Initialize script environment
    initialize_environment "$script_dir" "$option" "${LOG}"
    
    # Create timestamped backup directory
    TIMESTAMP=$(date '+%Y-%m-%d_%H-%M-%S')
    BACKUP_DIR_WITH_TIMESTAMP="${BACKUP_DIR}/${TIMESTAMP}_trakt-backup"
    mkdir -p "$BACKUP_DIR_WITH_TIMESTAMP"
    echo "📂 Created backup directory: $BACKUP_DIR_WITH_TIMESTAMP" | tee -a "${LOG}"
    
    # Set a global variable for the current backup directory to avoid nested paths
    export CURRENT_BACKUP_DIR="$BACKUP_DIR_WITH_TIMESTAMP"
    
    echo "🚀 Starting fetch_all_data with mode: '$mode'" | tee -a "${LOG}"
    # Fetch all data from Trakt
    fetch_all_data "$API_URL" "$API_KEY" "$API_SECRET" "$ACCESS_TOKEN" "$REFRESH_TOKEN" "$REDIRECT_URI" "$USERNAME" "$mode" "$BACKUP_DIR_WITH_TIMESTAMP" "${CONFIG_DIR}/.config.cfg" "$SED_INPLACE" "${LOG}"
    
    # Process data and create Letterboxd CSV
    process_data "$mode" "$USERNAME" "$BACKUP_DIR_WITH_TIMESTAMP" "$TEMP_DIR" "$DOSCOPY" "${LOG}"
    local process_result=$?
    
    if [ $process_result -eq 0 ]; then
        echo "✅ Export process completed. CSV file is ready for Letterboxd import." | tee -a "${LOG}"
        return 0
    else
        echo "❌ Export process completed with errors. CSV file may not be available." | tee -a "${LOG}"
        return 1
    fi
} 