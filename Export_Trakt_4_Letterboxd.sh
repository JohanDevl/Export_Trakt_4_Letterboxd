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
    ratings/movies
    watched/movies
    )     
	else
		echo -e "${SAISPAS}${BOLD}[`date`] - Unknown variable, normal mode activated${NC}" | tee -a "${LOG}"
		OPTION=$(echo "normal")
    endpoints=(
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

# Check the passed option
debug_msg "Checking option: $OPTION"

if [ "$OPTION" == "complete" ]; then
  debug_msg "Complete Mode activated - will process data and create backup"
  
  # Create empty CSV files to prevent errors
  touch "${TEMP_DIR}/ratings_movies.csv"
  
  # Add header to CSV file
  echo "Title, Year, imdbID, tmdbID, WatchedDate, Rating10" > "${TEMP_DIR}/ratings_movies.csv"
  
  # Process movie ratings
  if [ -f "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" ] && [ $(stat -c%s "${BACKUP_DIR}/${USERNAME}-ratings_movies.json") -gt 10 ]; then
    echo "DEBUG: Processing ratings file - checking for date fields" | tee -a "${LOG}"
    jq -r 'first | keys' "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" | tee -a "${LOG}"
    
    # First get the watched history to create a lookup for watch dates
    echo "DEBUG: Creating watched date lookup" | tee -a "${LOG}"
    if [ -f "${BACKUP_DIR}/${USERNAME}-history_movies.json" ]; then
      jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt] = $item.watched_at)' "${BACKUP_DIR}/${USERNAME}-history_movies.json" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: Watched date lookup created" | tee -a "${LOG}"
    elif [ -f "${BACKUP_DIR}/${USERNAME}-watched_movies.json" ]; then
      jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt] = $item.last_watched_at)' "${BACKUP_DIR}/${USERNAME}-watched_movies.json" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: Watched date lookup created from watched_movies" | tee -a "${LOG}"
    else
      echo "{}" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: No watch history found, using empty lookup" | tee -a "${LOG}"
    fi
    
    # Now process ratings with watched dates
    jq -r --slurpfile dates "${TEMP_DIR}/watched_dates.json" '.[] | select(.type == "movie") | 
      [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
      ($dates[0][.movie.ids.trakt | tostring] // .rated_at | split("T")[0]), 
      .rating] | map(. | tostring) | join(",")' "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" | 
    sed 's/\(.*\),\(.*\),\(.*\),\(.*\),\(.*\),\(.*\)/"\1",\2,"tt\3",\4,\5,\6/' | 
    sed 's/"tttt/"tt/g' >> "${TEMP_DIR}/ratings_movies.csv"
    echo -e "Movies ratings: Ratings Retrieved"
  else
    echo -e "Movies ratings: No Ratings Retrieved"
  fi

  # Process watched history
  if [ -f "${BACKUP_DIR}/${USERNAME}-watched_movies.json" ] && [ $(stat -c%s "${BACKUP_DIR}/${USERNAME}-watched_movies.json") -gt 10 ]; then
    jq -r '.[] | [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, (.last_watched_at // .rated_at | split("T")[0]), ""] | map(. | tostring) | join(",")' "${BACKUP_DIR}/${USERNAME}-watched_movies.json" |
    sed 's/\(.*\),\(.*\),\(.*\),\(.*\),\(.*\),\(.*\)/"\1",\2,"tt\3",\4,\5,\6/' |
    sed 's/"tttt/"tt/g' >> "${TEMP_DIR}/ratings_movies.csv"
    echo -e "Movies history: History Retrieved"
  else
    echo -e "Movies history: No History Retrieved"
  fi
  
  # Copy files to final destination
  cp "${TEMP_DIR}/ratings_movies.csv" "${DOSCOPY}/letterboxd_import.csv"
  debug_msg "CSV file created in ${DOSCOPY}/letterboxd_import.csv"
  
  # Now create the backup
  debug_msg "Creating backup archive"
  tar -czvf "${BACKUP_DIR}/backup-$(date '+%Y%m%d%H%M%S').tar.gz" -C "$(dirname "${BACKUP_DIR}")" "$(basename "${BACKUP_DIR}")"
  echo -e "Backup completed"
elif [ "$OPTION" == "initial" ]; then
  debug_msg "Initial Mode activated - will process data"
  
  # Create empty CSV files to prevent errors
  touch "${TEMP_DIR}/ratings_movies.csv"
  
  # Add header to CSV file
  echo "Title, Year, imdbID, tmdbID, WatchedDate, Rating10" > "${TEMP_DIR}/ratings_movies.csv"
  
  # Process movie ratings
  if [ -f "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" ] && [ $(stat -c%s "${BACKUP_DIR}/${USERNAME}-ratings_movies.json") -gt 10 ]; then
    echo "DEBUG: Processing ratings file - checking for date fields" | tee -a "${LOG}"
    jq -r 'first | keys' "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" | tee -a "${LOG}"
    
    # First get the watched history to create a lookup for watch dates
    echo "DEBUG: Creating watched date lookup" | tee -a "${LOG}"
    if [ -f "${BACKUP_DIR}/${USERNAME}-history_movies.json" ]; then
      jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt] = $item.watched_at)' "${BACKUP_DIR}/${USERNAME}-history_movies.json" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: Watched date lookup created" | tee -a "${LOG}"
    elif [ -f "${BACKUP_DIR}/${USERNAME}-watched_movies.json" ]; then
      jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt] = $item.last_watched_at)' "${BACKUP_DIR}/${USERNAME}-watched_movies.json" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: Watched date lookup created from watched_movies" | tee -a "${LOG}"
    else
      echo "{}" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: No watch history found, using empty lookup" | tee -a "${LOG}"
    fi
    
    # Now process ratings with watched dates
    jq -r --slurpfile dates "${TEMP_DIR}/watched_dates.json" '.[] | select(.type == "movie") | 
      [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
      ($dates[0][.movie.ids.trakt | tostring] // .rated_at | split("T")[0]), 
      .rating] | map(. | tostring) | join(",")' "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" | 
    sed 's/\(.*\),\(.*\),\(.*\),\(.*\),\(.*\),\(.*\)/"\1",\2,"tt\3",\4,\5,\6/' | 
    sed 's/"tttt/"tt/g' >> "${TEMP_DIR}/ratings_movies.csv"
    echo -e "Movies ratings: Ratings Retrieved"
  else
    echo -e "Movies ratings: No Ratings Retrieved"
  fi

  # Process watched history
  if [ -f "${BACKUP_DIR}/${USERNAME}-watched_movies.json" ] && [ $(stat -c%s "${BACKUP_DIR}/${USERNAME}-watched_movies.json") -gt 10 ]; then
    jq -r '.[] | [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, (.last_watched_at // .rated_at | split("T")[0]), ""] | map(. | tostring) | join(",")' "${BACKUP_DIR}/${USERNAME}-watched_movies.json" |
    sed 's/\(.*\),\(.*\),\(.*\),\(.*\),\(.*\),\(.*\)/"\1",\2,"tt\3",\4,\5,\6/' |
    sed 's/"tttt/"tt/g' >> "${TEMP_DIR}/ratings_movies.csv"
    echo -e "Movies history: History Retrieved"
  else
    echo -e "Movies history: No History Retrieved"
  fi
  
  # Copy files to final destination
  cp "${TEMP_DIR}/ratings_movies.csv" "${DOSCOPY}/letterboxd_import.csv"
  debug_msg "CSV file created in ${DOSCOPY}/letterboxd_import.csv"
else
  debug_msg "Normal Mode activated - will process data"
  
  # Create empty CSV files to prevent errors
  touch "${TEMP_DIR}/ratings_movies.csv"
  
  # Add header to CSV file
  echo "Title, Year, imdbID, tmdbID, WatchedDate, Rating10" > "${TEMP_DIR}/ratings_movies.csv"
  
  # Process movie ratings
  if [ -f "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" ] && [ $(stat -c%s "${BACKUP_DIR}/${USERNAME}-ratings_movies.json") -gt 10 ]; then
    echo "DEBUG: Processing ratings file - checking for date fields" | tee -a "${LOG}"
    jq -r 'first | keys' "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" | tee -a "${LOG}"
    
    # First get the watched history to create a lookup for watch dates
    echo "DEBUG: Creating watched date lookup" | tee -a "${LOG}"
    if [ -f "${BACKUP_DIR}/${USERNAME}-history_movies.json" ]; then
      jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt] = $item.watched_at)' "${BACKUP_DIR}/${USERNAME}-history_movies.json" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: Watched date lookup created" | tee -a "${LOG}"
    elif [ -f "${BACKUP_DIR}/${USERNAME}-watched_movies.json" ]; then
      jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt] = $item.last_watched_at)' "${BACKUP_DIR}/${USERNAME}-watched_movies.json" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: Watched date lookup created from watched_movies" | tee -a "${LOG}"
    else
      echo "{}" > "${TEMP_DIR}/watched_dates.json"
      echo "DEBUG: No watch history found, using empty lookup" | tee -a "${LOG}"
    fi
    
    # Now process ratings with watched dates
    jq -r --slurpfile dates "${TEMP_DIR}/watched_dates.json" '.[] | select(.type == "movie") | 
      [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
      ($dates[0][.movie.ids.trakt | tostring] // .rated_at | split("T")[0]), 
      .rating] | map(. | tostring) | join(",")' "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" | 
    sed 's/\(.*\),\(.*\),\(.*\),\(.*\),\(.*\),\(.*\)/"\1",\2,"tt\3",\4,\5,\6/' | 
    sed 's/"tttt/"tt/g' >> "${TEMP_DIR}/ratings_movies.csv"
    echo -e "Movies ratings: Ratings Retrieved"
  else
    echo -e "Movies ratings: No Ratings Retrieved"
  fi

  # Process watched history
  if [ -f "${BACKUP_DIR}/${USERNAME}-history_movies.json" ] && [ $(stat -c%s "${BACKUP_DIR}/${USERNAME}-history_movies.json") -gt 10 ]; then
    jq -r '.[] | [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, (.last_watched_at // .rated_at | split("T")[0]), ""] | map(. | tostring) | join(",")' "${BACKUP_DIR}/${USERNAME}-history_movies.json" |
    sed 's/\(.*\),\(.*\),\(.*\),\(.*\),\(.*\),\(.*\)/"\1",\2,"tt\3",\4,\5,\6/' |
    sed 's/"tttt/"tt/g' >> "${TEMP_DIR}/ratings_movies.csv"
    echo -e "Movies history: History Retrieved"
  else
    echo -e "Movies history: No History Retrieved"
  fi
  
  # Copy files to final destination
  cp "${TEMP_DIR}/ratings_movies.csv" "${DOSCOPY}/letterboxd_import.csv"
  debug_msg "CSV file created in ${DOSCOPY}/letterboxd_import.csv"
fi
