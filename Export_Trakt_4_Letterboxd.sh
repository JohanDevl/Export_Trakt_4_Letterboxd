#!/bin/bash
############################################################################## 
#                                                                            #
#	SHELL: !/bin/bash       version 2  	                                     #
#									                                                           #
#	NAME: u2pitchjami						                                               #
#									                                                           #
#							  					                                                   #
#									                                                           #
#	DATE: 18/09/2024          	           				                             #
#									                                                           #
#	PURPOSE: Export trakt to letterboxd format                             		 #
#									                                                           #
############################################################################## 
# Trakt backup script (note that user profile must be public)
# Trakt API documentation: http://docs.trakt.apiary.io
# Trakt client API key: http://docs.trakt.apiary.io/#introduction/create-an-app
SCRIPT_DIR=$(dirname "$(realpath "$0")")
echo "$SCRIPT_DIR"

# Debug options
echo "=========== DEBUG INFORMATION ==========="
echo "Script called with option: $1"
echo "Number of arguments: $#"
if [ -n "$1" ]; then
  echo "Option value: '$1'"
else
  echo "No option provided, using default"
fi
echo "========================================="

# Detect OS for sed compatibility
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS uses BSD sed
    SED_INPLACE="sed -i ''"
    echo "Detected macOS: Using BSD sed with empty string backup parameter" | tee -a "${LOG}"
else
    # Linux and others use GNU sed
    SED_INPLACE="sed -i"
    echo "Detected Linux/other: Using GNU sed" | tee -a "${LOG}"
fi

# Debug messaging function
debug_msg() {
    local message="$1"
    echo -e "DEBUG: $message" | tee -a "${LOG}"
}

# File manipulation debug function
debug_file_info() {
    local file="$1"
    local message="$2"
    
    echo "📄 $message:" | tee -a "${LOG}"
    if [ -f "$file" ]; then
        echo "  - File exists: ✅" | tee -a "${LOG}"
        echo "  - File size: $(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "unknown") bytes" | tee -a "${LOG}"
        echo "  - File permissions: $(ls -la "$file" | awk '{print $1}')" | tee -a "${LOG}"
        echo "  - Owner: $(ls -la "$file" | awk '{print $3":"$4}')" | tee -a "${LOG}"
        
        # Check if file is readable
        if [ -r "$file" ]; then
            echo "  - File is readable: ✅" | tee -a "${LOG}"
        else
            echo "  - File is readable: ❌" | tee -a "${LOG}"
        fi
        
        # Check if file is writable
        if [ -w "$file" ]; then
            echo "  - File is writable: ✅" | tee -a "${LOG}"
        else
            echo "  - File is writable: ❌" | tee -a "${LOG}"
        fi
        
        # Check if file has content
        if [ -s "$file" ]; then
            echo "  - File has content: ✅" | tee -a "${LOG}"
            echo "  - First line: $(head -n 1 "$file" 2>/dev/null || echo "Cannot read file")" | tee -a "${LOG}"
            echo "  - Line count: $(wc -l < "$file" 2>/dev/null || echo "Cannot count lines")" | tee -a "${LOG}"
        else
            echo "  - File has content: ❌ (empty file)" | tee -a "${LOG}"
        fi
    else
        echo "  - File exists: ❌ (not found)" | tee -a "${LOG}"
        echo "  - Directory exists: $(if [ -d "$(dirname "$file")" ]; then echo "✅"; else echo "❌"; fi)" | tee -a "${LOG}"
        echo "  - Directory permissions: $(ls -la "$(dirname "$file")" 2>/dev/null | head -n 1 | awk '{print $1}' || echo "Cannot access directory")" | tee -a "${LOG}"
    fi
    echo "-----------------------------------" | tee -a "${LOG}"
}

# Always use the config file from the config directory
CONFIG_DIR="${SCRIPT_DIR}/config"
if [ -f "/app/config/.config.cfg" ]; then
    # If running in Docker, use the absolute path
    source /app/config/.config.cfg
    echo "Using Docker config file: /app/config/.config.cfg" | tee -a "${LOG}"
else
    # If running locally, use the relative path
    source ${CONFIG_DIR}/.config.cfg
    echo "Using local config file: ${CONFIG_DIR}/.config.cfg" | tee -a "${LOG}"
fi

# Use the user's temporary directory
TEMP_DIR="/tmp/trakt_export_$USER"
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"
echo "Created temporary directory: $TEMP_DIR" | tee -a "${LOG}"

if [ ! -d $DOSLOG ]
	then
	mkdir -p $DOSLOG
	echo "Created log directory: $DOSLOG" | tee -a "${LOG}"
fi

# Log environment information
echo "🌍 Environment information:" | tee -a "${LOG}"
echo "  - User: $(whoami)" | tee -a "${LOG}"
echo "  - Working directory: $(pwd)" | tee -a "${LOG}"
echo "  - Script directory: $SCRIPT_DIR" | tee -a "${LOG}"
echo "  - Copy directory: $DOSCOPY" | tee -a "${LOG}"
echo "  - Log directory: $DOSLOG" | tee -a "${LOG}"
echo "  - Backup directory: $BACKUP_DIR" | tee -a "${LOG}"
echo "  - OS Type: $OSTYPE" | tee -a "${LOG}"
echo "-----------------------------------" | tee -a "${LOG}"

# Check key directories
if [ -d "$DOSCOPY" ]; then
    echo "Copy directory exists: ✅" | tee -a "${LOG}"
    echo "Copy directory permissions: $(ls -la "$DOSCOPY" | head -n 1 | awk '{print $1}')" | tee -a "${LOG}"
else
    echo "Copy directory exists: ❌ (will attempt to create)" | tee -a "${LOG}"
fi

# Check for existing CSV file
if [ -f "${DOSCOPY}/letterboxd_import.csv" ]; then
    debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Existing CSV file check"
fi

refresh_access_token() {
    echo "🔄 Refreshing Trakt token..." | tee -a "${LOG}"
    echo "  - Using refresh token: ${REFRESH_TOKEN:0:5}...${REFRESH_TOKEN: -5}" | tee -a "${LOG}"
    echo "  - API key: ${API_KEY:0:5}...${API_KEY: -5}" | tee -a "${LOG}"
    
    RESPONSE=$(curl -s -X POST "https://api.trakt.tv/oauth/token" \
        -H "Content-Type: application/json" -v \
        -d "{
            \"refresh_token\": \"${REFRESH_TOKEN}\",
            \"client_id\": \"${API_KEY}\",
            \"client_secret\": \"${API_SECRET}\",
            \"redirect_uri\": \"${REDIRECT_URI}\",
            \"grant_type\": \"refresh_token\"
        }")

    # Debug response (without exposing sensitive data)
    echo "  - Response received: $(if [ -n "$RESPONSE" ]; then echo "✅"; else echo "❌ (empty)"; fi)" | tee -a "${LOG}"
    
    NEW_ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
    NEW_REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refresh_token')

    if [[ "$NEW_ACCESS_TOKEN" != "null" && "$NEW_REFRESH_TOKEN" != "null" ]]; then
        echo "✅ Token refreshed successfully." | tee -a "${LOG}"
        echo "  - New access token: ${NEW_ACCESS_TOKEN:0:5}...${NEW_ACCESS_TOKEN: -5}" | tee -a "${LOG}"
        echo "  - New refresh token: ${NEW_REFRESH_TOKEN:0:5}...${NEW_REFRESH_TOKEN: -5}" | tee -a "${LOG}"
        
        # Determine which config file to update
        CONFIG_FILE="/app/config/.config.cfg"
        if [ ! -f "$CONFIG_FILE" ]; then
            CONFIG_FILE="${CONFIG_DIR}/.config.cfg"
        fi
        
        echo "  - Updating config file: $CONFIG_FILE" | tee -a "${LOG}"
        
        # Check if config file exists and is writable
        if [ -f "$CONFIG_FILE" ]; then
            if [ -w "$CONFIG_FILE" ]; then
                echo "  - Config file is writable: ✅" | tee -a "${LOG}"
            else
                echo "  - Config file is writable: ❌ - Permissions: $(ls -la "$CONFIG_FILE" | awk '{print $1}')" | tee -a "${LOG}"
            fi
        else
            echo "  - Config file exists: ❌ (not found)" | tee -a "${LOG}"
        fi
        
        $SED_INPLACE "s|ACCESS_TOKEN=.*|ACCESS_TOKEN=\"$NEW_ACCESS_TOKEN\"|" "$CONFIG_FILE"
        $SED_INPLACE "s|REFRESH_TOKEN=.*|REFRESH_TOKEN=\"$NEW_REFRESH_TOKEN\"|" "$CONFIG_FILE"
        
        echo "  - Config file updated: $(if [ $? -eq 0 ]; then echo "✅"; else echo "❌"; fi)" | tee -a "${LOG}"
        
        # Re-source the config file to update variables
        source "$CONFIG_FILE"
        echo "  - Config file re-sourced" | tee -a "${LOG}"
    else
        echo "❌ Error refreshing token. Check your configuration!" | tee -a "${LOG}"
        echo "  - Response: $RESPONSE" | tee -a "${LOG}"
        echo "  - Make sure your API credentials are correct and try again." | tee -a "${LOG}"
        exit 1
    fi
}

##########################CHECK IF "COMPLETE" OPTION IS ACTIVE################
if [ ! -z $1 ]
	then
	OPTION=$(echo $1 | tr '[:upper:]' '[:lower:]')
	if [ $OPTION == "complete" ]
		then
		echo -e "${SAISPAS}${BOLD}[`date`] - Complete Mode activated${NC}" | tee -a "${LOG}"
    endpoints=(
    watchlist/movies
    watchlist/shows
    watchlist/episodes
    watchlist/seasons
    ratings/movies
    ratings/shows
    ratings/episodes
    ratings/seasons
    collection/movies
    collection/shows
    watched/movies
    watched/shows
    history/movies
    history/shows
    ) 
  elif [ $OPTION == "initial" ]
		then
		echo -e "${SAISPAS}${BOLD}[`date`] - Initial Mode activated${NC}" | tee -a "${LOG}"
    endpoints=(
    history/movies
    ratings/movies
    watched/movies
    )     
	else
		echo -e "${SAISPAS}${BOLD}[`date`] - Unknown variable, normal mode activated${NC}" | tee -a "${LOG}"
		OPTION=$(echo "normal")
    endpoints=(
    history/movies
    ratings/movies
    ratings/episodes
    history/movies
    history/shows
    history/episodes
    watchlist/movies
    watchlist/shows
    )  
	fi
else
  OPTION=$(echo "normal")
  echo -e "${SAISPAS}${BOLD}[`date`] - Normal Mode activated${NC}" | tee -a "${LOG}"
  endpoints=(
    history/movies
    ratings/movies
    ratings/episodes
    history/movies
    history/shows
    history/episodes
    watchlist/movies
    watchlist/shows
    )     
fi

echo -e "Retrieving information..." | tee -a "${LOG}"

# create backup folder
mkdir -p ${BACKUP_DIR}

# Check if the token is still valid before each request
RESPONSE=$(curl -s -X GET "${API_URL}/users/me/history/movies" \
    -H "Content-Type: application/json" \
    -H "trakt-api-key: ${API_KEY}" \
    -H "trakt-api-version: 2" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}")

if echo "$RESPONSE" | grep -q "invalid_grant"; then
    echo "⚠️ Token expired, attempting to refresh..."
    refresh_access_token
fi

# Trakt requests
for endpoint in ${endpoints[*]}
do
  filename="${USERNAME}-${endpoint//\//_}.json"
 
  # Check if tokens are defined
  if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = '""' ] || [ "$ACCESS_TOKEN" = "" ]; then
    echo -e "\e[31mERROR: ACCESS_TOKEN not defined. Run the setup_trakt.sh script first to get a token.\e[0m" | tee -a "${LOG}"
    echo -e "Command: ./setup_trakt.sh" | tee -a "${LOG}"
    exit 1
  fi

  curl -X GET "${API_URL}/users/me/${endpoint}" \
    -H "Content-Type: application/json" \
    -H "trakt-api-key: ${API_KEY}" \
    -H "trakt-api-version: 2" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -o ${BACKUP_DIR}/${filename} \
    && echo -e "\e[32m${USERNAME}/${endpoint}\e[0m Retrieved successfully" \
    || echo -e "\e[31m${USERNAME}/${endpoint}\e[0m Request failed" | tee -a "${LOG}"
    
  # Check if the JSON file contains valid data
  if [ -f "${BACKUP_DIR}/${filename}" ]; then
    if ! jq empty "${BACKUP_DIR}/${filename}" 2>/dev/null; then
      echo -e "\e[31mERROR: The file ${filename} does not contain valid JSON.\e[0m" | tee -a "${LOG}"
    elif [ "$(jq '. | length' "${BACKUP_DIR}/${filename}")" = "0" ]; then
      echo -e "\e[33mWARNING: The file ${filename} does not contain any data.\e[0m" | tee -a "${LOG}"
    fi
  else
    echo -e "\e[31mERROR: The file ${filename} was not created.\e[0m" | tee -a "${LOG}"
  fi
done

echo -e "All files have been retrieved\n Starting processing" | tee -a "${LOG}"

# Create the output directory if it doesn't exist
mkdir -p "$DOSCOPY"

# Create empty CSV file with header
echo "Title,Year,imdbID,tmdbID,WatchedDate,Rating10,Rewatch" > "${TEMP_DIR}/movies_export.csv"

# Create ratings lookup
if [ -f "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" ]; then
  echo "DEBUG: Creating ratings lookup file..." | tee -a "${LOG}"
  jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt | tostring] = $item.rating)' "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" > "${TEMP_DIR}/ratings_lookup.json"
else
  echo "{}" > "${TEMP_DIR}/ratings_lookup.json"
fi

# Create a plays count lookup from watched_movies to determine rewatches
if [ -f "${BACKUP_DIR}/${USERNAME}-watched_movies.json" ]; then
  echo "DEBUG: Creating plays count lookup from watched_movies..." | tee -a "${LOG}"
  jq -c 'reduce .[] as $item ({}; if $item.movie.ids.imdb != null then .[$item.movie.ids.imdb] = $item.plays else . end)' "${BACKUP_DIR}/${USERNAME}-watched_movies.json" > "${TEMP_DIR}/plays_count_lookup.json"
else
  echo "{}" > "${TEMP_DIR}/plays_count_lookup.json"
fi

# Process watched movies history with ratings
if [ -f "${BACKUP_DIR}/${USERNAME}-history_movies.json" ]; then
  echo "DEBUG: Processing history_movies file with ratings..." | tee -a "${LOG}"
  
  # Extract watched movies with their date, rating, and rewatch status
  jq -r --slurpfile ratings "${TEMP_DIR}/ratings_lookup.json" --slurpfile plays "${TEMP_DIR}/plays_count_lookup.json" '.[] | 
    [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
     (if .watched_at then .watched_at | split("T")[0] else "" end),
     ($ratings[0][.movie.ids.trakt | tostring] // ""),
     (if ($plays[0][.movie.ids.imdb] // 1) > 1 then "true" else "false" end)] | 
    @csv' "${BACKUP_DIR}/${USERNAME}-history_movies.json" > "${TEMP_DIR}/raw_output.csv"
  
  # Process the CSV line by line to properly format
  cat "${TEMP_DIR}/raw_output.csv" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
    # Remove any existing "tt" prefix from IMDb ID to avoid duplication
    imdb=$(echo "$imdb" | sed 's/^"*tt//g' | sed 's/"*$//g')
    # Properly quote title and format IMDb ID with tt prefix
    echo "${title},${year},\"tt${imdb}\",${tmdb},${date},${rating},${rewatch}" >> "${TEMP_DIR}/movies_export.csv"
  done
  
  echo "Movies history: $(wc -l < "${TEMP_DIR}/movies_export.csv") movies processed" | tee -a "${LOG}"
fi

# In complete mode, also process watched_movies to get a more complete list
if [ "$OPTION" == "complete" ] && [ -f "${BACKUP_DIR}/${USERNAME}-watched_movies.json" ]; then
  echo "DEBUG: Processing watched_movies file with ratings (complete mode)..." | tee -a "${LOG}"
  
  # Extract all movie IDs from movies_export.csv to avoid duplicates
  awk -F, '{print $3}' "${TEMP_DIR}/movies_export.csv" | sed 's/"//g' > "${TEMP_DIR}/existing_imdb_ids.txt"
  
  # Use watched count from the API for rewatch status
  jq -r --slurpfile ratings "${TEMP_DIR}/ratings_lookup.json" '.[] | 
    [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
     (if .last_watched_at then .last_watched_at | split("T")[0] else "" end),
     ($ratings[0][.movie.ids.trakt | tostring] // ""),
     (if .plays > 1 then "true" else "false" end)] | 
    @csv' "${BACKUP_DIR}/${USERNAME}-watched_movies.json" > "${TEMP_DIR}/watched_raw_output.csv"
  
  # Process the CSV line by line, only adding movies not already in the history
  cat "${TEMP_DIR}/watched_raw_output.csv" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
    # Format IMDb ID consistently
    clean_imdb=$(echo "$imdb" | sed 's/^"*tt//g' | sed 's/"*$//g')
    formatted_imdb="tt${clean_imdb}"
    
    # Check if this movie is already in our list
    if ! grep -q "$formatted_imdb" "${TEMP_DIR}/existing_imdb_ids.txt"; then
      # Add this movie to our output
      echo "${title},${year},\"${formatted_imdb}\",${tmdb},${date},${rating},${rewatch}" >> "${TEMP_DIR}/movies_export.csv"
      # Add to our tracking of existing IDs
      echo "$formatted_imdb" >> "${TEMP_DIR}/existing_imdb_ids.txt"
    fi
  done
  
  echo "Total movies after combining history and watched list: $(wc -l < "${TEMP_DIR}/movies_export.csv") movies processed" | tee -a "${LOG}"
elif [ ! -f "${BACKUP_DIR}/${USERNAME}-history_movies.json" ] && [ -f "${BACKUP_DIR}/${USERNAME}-watched_movies.json" ]; then
  echo "DEBUG: No history found. Processing watched_movies file with ratings..." | tee -a "${LOG}"
  
  # Extract watched movies with their date and rating
  jq -r --slurpfile ratings "${TEMP_DIR}/ratings_lookup.json" '.[] | 
    [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
     (if .last_watched_at then .last_watched_at | split("T")[0] else "" end),
     ($ratings[0][.movie.ids.trakt | tostring] // ""),
     (if .plays > 1 then "true" else "false" end)] | 
    @csv' "${BACKUP_DIR}/${USERNAME}-watched_movies.json" > "${TEMP_DIR}/raw_output.csv"
  
  # Process the CSV line by line to properly format
  cat "${TEMP_DIR}/raw_output.csv" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
    # Remove any existing "tt" prefix from IMDb ID to avoid duplication
    imdb=$(echo "$imdb" | sed 's/^"*tt//g' | sed 's/"*$//g')
    # Properly quote title and format IMDb ID with tt prefix
    echo "${title},${year},\"tt${imdb}\",${tmdb},${date},${rating},${rewatch}" >> "${TEMP_DIR}/movies_export.csv"
  done
  
  echo "Movies watched: $(wc -l < "${TEMP_DIR}/movies_export.csv") movies processed" | tee -a "${LOG}"
else
  echo -e "Movies history: No movies found in history or watched" | tee -a "${LOG}"
fi

# Copy file to final destination
cp "${TEMP_DIR}/movies_export.csv" "${DOSCOPY}/letterboxd_import.csv"
debug_msg "CSV file created in ${DOSCOPY}/letterboxd_import.csv"

# Create backup if in complete mode
if [ "$OPTION" == "complete" ]; then
  debug_msg "Creating backup archive"
  tar -czvf "${BACKUP_DIR}/backup-$(date '+%Y%m%d%H%M%S').tar.gz" -C "$(dirname "${BACKUP_DIR}")" "$(basename "${BACKUP_DIR}")"
  echo -e "Backup completed"
fi

echo "Export process completed. CSV file is ready for Letterboxd import."
