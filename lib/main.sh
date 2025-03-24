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

# Process data from Trakt API
process_data() {
    local backup_dir="$1"
    local temp_dir="$2"
    local dosdir="$3"
    local log="$4"
    local username="$5"
    local option="$6"
    
    # Create alternate paths for the output file, in case the script is not run from the root directory
    local final_output_file="${dosdir}/letterboxd_import.csv"
    local alt_paths=(
        "${dosdir}/letterboxd_import.csv"
        "./copy/letterboxd_import.csv"
        "/app/copy/letterboxd_import.csv"
    )
    
    # Remove any existing CSV files at alternate locations to avoid confusion
    for alt_path in "${alt_paths[@]}"; do
        if [ -f "$alt_path" ] && [ "$alt_path" != "$final_output_file" ]; then
            echo "🗑️ Removing CSV file at alternate location: $alt_path" | tee -a "${log}"
            rm -f "$alt_path"
        fi
    done
    
    # Create empty CSV file with header
    echo "Title,Year,imdbID,tmdbID,WatchedDate,Rating10,Rewatch" > "${temp_dir}/movies_export.csv"
    
    # Show size and content of backup directory
    echo "🔍 Diagnostic: Checking backup directory" | tee -a "${log}"
    echo "  - Backup directory path: $backup_dir" | tee -a "${log}"
    echo "  - Directory exists: $(if [ -d "$backup_dir" ]; then echo "Yes"; else echo "No"; fi)" | tee -a "${log}"
    echo "  - Directory content:" | tee -a "${log}"
    ls -la "$backup_dir" 2>&1 | tee -a "${log}"
    
    # Create ratings lookup
    local ratings_file="${backup_dir}/${username}-ratings_movies.json"
    echo "📄 Looking for ratings file: $ratings_file" | tee -a "${log}"
    echo "  - File exists: $(if [ -f "$ratings_file" ]; then echo "Yes"; else echo "No"; fi)" | tee -a "${log}"
    if [ -f "$ratings_file" ]; then
        echo "  - File size: $(wc -c < "$ratings_file") bytes" | tee -a "${log}"
        echo "  - Number of ratings: $(jq length "$ratings_file" 2>/dev/null || echo "Error parsing JSON")" | tee -a "${log}"
    fi
    
    local ratings_lookup="${temp_dir}/ratings_lookup.json"
    create_ratings_lookup "$ratings_file" "$ratings_lookup" "$log"
    
    # Create plays count lookup
    local watched_file="${backup_dir}/${username}-watched_movies.json"
    echo "📄 Looking for watched file: $watched_file" | tee -a "${log}"
    echo "  - File exists: $(if [ -f "$watched_file" ]; then echo "Yes"; else echo "No"; fi)" | tee -a "${log}"
    if [ -f "$watched_file" ]; then
        echo "  - File size: $(wc -c < "$watched_file") bytes" | tee -a "${log}"
        echo "  - Number of watched movies: $(jq length "$watched_file" 2>/dev/null || echo "Error parsing JSON")" | tee -a "${log}"
    fi
    
    local plays_lookup="${temp_dir}/plays_count_lookup.json"
    create_plays_count_lookup "$watched_file" "$plays_lookup" "$log"
    
    # Process history movies
    local history_file="${backup_dir}/${username}-history_movies.json"
    echo "📄 Looking for history file: $history_file" | tee -a "${log}"
    echo "  - File exists: $(if [ -f "$history_file" ]; then echo "Yes"; else echo "No"; fi)" | tee -a "${log}"
    if [ -f "$history_file" ]; then
        echo "  - File size: $(wc -c < "$history_file") bytes" | tee -a "${log}"
        echo "  - Number of history entries: $(jq length "$history_file" 2>/dev/null || echo "Error parsing JSON")" | tee -a "${log}"
        if [ "$(jq length "$history_file" 2>/dev/null || echo 0)" -gt 0 ]; then
            echo "  - First movie: $(jq -r '.[0].movie.title' "$history_file" 2>/dev/null || echo "Error parsing JSON")" | tee -a "${log}"
            echo "  - First watch date: $(jq -r '.[0].watched_at' "$history_file" 2>/dev/null || echo "Error parsing JSON")" | tee -a "${log}"
        fi
    fi
    
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
    
    # Check the initial CSV before deduplication
    echo "📄 CSV status before deduplication:" | tee -a "${log}"
    echo "  - File exists: $(if [ -f "$csv_output" ]; then echo "Yes"; else echo "No"; fi)" | tee -a "${log}"
    if [ -f "$csv_output" ]; then
        echo "  - File size: $(wc -c < "$csv_output") bytes" | tee -a "${log}"
        echo "  - Number of lines: $(wc -l < "$csv_output")" | tee -a "${log}"
        echo "  - First few lines:" | tee -a "${log}"
        head -n 5 "$csv_output" | tee -a "${log}"
    fi
    
    # Deduplicate the final CSV to ensure no duplicate entries
    local final_csv="${temp_dir}/final_export.csv"
    deduplicate_movies "$csv_output" "$final_csv" "$log"
    
    # Apply limit if LIMIT_FILMS is set or if mode is "normal"
    local limited_csv="${temp_dir}/limited_export.csv"
    
    # Set default limit for normal mode if LIMIT_FILMS is not explicitly set
    if [ "$option" = "normal" ] && [ -z "$LIMIT_FILMS" ]; then
        export LIMIT_FILMS=10
        echo "🎯 Normal mode: Automatically limiting to 10 most recent films" | tee -a "${log}"
    fi
    
    limit_movies_in_csv "$final_csv" "$limited_csv" "$log"
    final_csv="$limited_csv"
    
    # Make sure the final CSV file was created
    if [ ! -f "$final_csv" ]; then
        echo "❌ ERROR: Final CSV file was not created in temp directory" | tee -a "${log}"
        return 1
    fi
    
    # Check the final CSV before copying
    echo "📄 Final CSV status before copying:" | tee -a "${log}"
    echo "  - File exists: $(if [ -f "$final_csv" ]; then echo "Yes"; else echo "No"; fi)" | tee -a "${log}"
    if [ -f "$final_csv" ]; then
        echo "  - File size: $(wc -c < "$final_csv") bytes" | tee -a "${log}"
        echo "  - Number of lines: $(wc -l < "$final_csv")" | tee -a "${log}"
        echo "  - First few lines:" | tee -a "${log}"
        head -n 5 "$final_csv" | tee -a "${log}"
    fi
    
    # Ensure target directory exists with proper permissions
    if [ ! -d "$dosdir" ]; then
        echo "📁 Creating copy directory: $dosdir" | tee -a "${log}"
        mkdir -p "$dosdir"
        chmod 777 "$dosdir"
    else
        echo "📁 Ensuring copy directory has proper permissions" | tee -a "${log}"
        chmod 777 "$dosdir"
    fi
    
    # Try to copy file to final destination with verbose output
    echo "📋 Copying final CSV to: ${dosdir}/letterboxd_import.csv" | tee -a "${log}"
    cp -v "$final_csv" "${dosdir}/letterboxd_import.csv" 2>&1 | tee -a "${log}"
    
    if [ -f "${dosdir}/letterboxd_import.csv" ]; then
        echo "✅ CSV file created in ${dosdir}/letterboxd_import.csv" | tee -a "${log}"
        echo "✅ Export process completed. CSV file is ready for Letterboxd import." | tee -a "${log}"
        return 0
    else
        echo "❌ ERROR: Failed to create CSV file in ${dosdir}/letterboxd_import.csv" | tee -a "${log}"
        return 1
    fi
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
    process_data "$BACKUP_DIR_WITH_TIMESTAMP" "$TEMP_DIR" "$DOSCOPY" "${LOG}" "$USERNAME" "$mode"
    local process_result=$?
    
    if [ $process_result -eq 0 ]; then
        echo "✅ Export process completed. CSV file is ready for Letterboxd import." | tee -a "${LOG}"
        return 0
    else
        echo "❌ Export process completed with errors. CSV file may not be available." | tee -a "${LOG}"
        return 1
    fi
} 