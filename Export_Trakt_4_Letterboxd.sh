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

if [ $OPTION == "complete" ]
	then
# compress backup folder
echo -e "Compressing backup..." | tee -a "${LOG}"
tar -czvf ${BACKUP_DIR}.tar.gz ${BACKUP_DIR}
echo -e "Backup compressed: \e[32m${BACKUP_DIR}.tar.gz\e[0m\n" | tee -a "${LOG}"
echo -e "That's it, backup completed" | tee -a "${LOG}"
else
    if [ $OPTION == "initial" ]
      then
      cat ${BACKUP_DIR}/${USERNAME}-ratings_movies.json | jq -r '.[]|[.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, .last_watched_at, .rating]|@csv' >> "$TEMP_DIR/temp_rating.csv"
      cat ${BACKUP_DIR}/${USERNAME}-watched_movies.json | jq -r '.[]|[.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, .last_watched_at, .rating]|@csv' >> "$TEMP_DIR/temp.csv"
    else
      # Create empty files to avoid "No such file or directory" errors
      touch "$TEMP_DIR/temp_rating.csv"
      touch "$TEMP_DIR/temp_rating_episodes.csv"
      touch "$TEMP_DIR/temp.csv"
      touch "$TEMP_DIR/temp_show.csv"
      touch "$TEMP_DIR/temp_watchlist.csv"
      
      # Make sure files have the correct permissions
      chmod 644 "$TEMP_DIR/"*.csv
      
      # Process JSON files if they exist and contain data
      if [ -f "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" ] && [ -s "${BACKUP_DIR}/${USERNAME}-ratings_movies.json" ]; then
        cat ${BACKUP_DIR}/${USERNAME}-ratings_movies.json | jq -r '.[]|[.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, .watched_at, .rating]|@csv' >> "$TEMP_DIR/temp_rating.csv"
      fi
      
      if [ -f "${BACKUP_DIR}/${USERNAME}-ratings_episodes.json" ] && [ -s "${BACKUP_DIR}/${USERNAME}-ratings_episodes.json" ]; then
        cat ${BACKUP_DIR}/${USERNAME}-ratings_episodes.json | jq -r '.[]|[.show.title, .show.year, .episode.title, .episode.season, .episode.number, .show.ids.imdb, .show.ids.tmdb, .watched_at, .rating]|@csv' >> "$TEMP_DIR/temp_rating_episodes.csv"
      fi
      
      if [ -f "${BACKUP_DIR}/${USERNAME}-history_movies.json" ] && [ -s "${BACKUP_DIR}/${USERNAME}-history_movies.json" ]; then
        cat ${BACKUP_DIR}/${USERNAME}-history_movies.json | jq -r '.[]|[.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, .watched_at, .rating]|@csv' >> "$TEMP_DIR/temp.csv"
      fi
      
      if [ -f "${BACKUP_DIR}/${USERNAME}-history_shows.json" ] && [ -s "${BACKUP_DIR}/${USERNAME}-history_shows.json" ]; then
        cat ${BACKUP_DIR}/${USERNAME}-history_shows.json | jq -r '.[]|[.show.title, .show.year, .episode.title, .episode.season, .episode.number, .show.ids.imdb, .show.ids.tmdb, .watched_at, .rating]|@csv' >> "$TEMP_DIR/temp_show.csv"
      fi
      
      if [ -f "${BACKUP_DIR}/${USERNAME}-watchlist_movies.json" ] && [ -s "${BACKUP_DIR}/${USERNAME}-watchlist_movies.json" ]; then
        cat ${BACKUP_DIR}/${USERNAME}-watchlist_movies.json | jq -r '.[]|[.type, .movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, .listed_at]|@csv' >> "$TEMP_DIR/temp_watchlist.csv"
      fi
      
      if [ -f "${BACKUP_DIR}/${USERNAME}-watchlist_shows.json" ] && [ -s "${BACKUP_DIR}/${USERNAME}-watchlist_shows.json" ]; then
        cat ${BACKUP_DIR}/${USERNAME}-watchlist_shows.json | jq -r '.[]|[.type, .show.title, .show.year, .show.ids.imdb, .show.ids.tmdb, .listed_at]|@csv' >> "$TEMP_DIR/temp_watchlist.csv"
      fi
    fi   
    
    # Check if temporary files contain data
    if [ ! -s "$TEMP_DIR/temp.csv" ]; then
        echo -e "\e[33mWARNING: No movie data was retrieved from the Trakt API.\e[0m" | tee -a "${LOG}"
        echo -e "The letterboxd_import.csv file will not be updated." | tee -a "${LOG}"
        # Create an empty letterboxd_import.csv file if it doesn't exist
        if [ ! -f "${DOSCOPY}/letterboxd_import.csv" ]; then
            echo -e "Generating an empty letterboxd_import.csv file" | tee -a "${LOG}"
            # Create the DOSCOPY directory if it doesn't exist
            mkdir -p "$DOSCOPY"
            echo "Created DOSCOPY directory: $DOSCOPY" | tee -a "${LOG}"
            echo "Directory permissions: $(ls -la "$DOSCOPY" | head -n 1 | awk '{print $1}')" | tee -a "${LOG}"
            
            chmod 755 "$DOSCOPY"
            echo "Set directory permissions to 755" | tee -a "${LOG}"
            echo "New directory permissions: $(ls -la "$DOSCOPY" | head -n 1 | awk '{print $1}')" | tee -a "${LOG}"
            
            echo "Title, Year, imdbID, tmdbID, WatchedDate, Rating10" > "${DOSCOPY}/letterboxd_import.csv"
            echo "Created empty CSV file with header" | tee -a "${LOG}"
            
            chmod 644 "${DOSCOPY}/letterboxd_import.csv"
            echo "Set CSV file permissions to 644" | tee -a "${LOG}"
            
            debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Empty CSV file creation"
        fi
    else
        echo "Processing movie data for CSV file" | tee -a "${LOG}"
        echo "Found $(cat "$TEMP_DIR/temp.csv" | wc -l) movies to process" | tee -a "${LOG}"
        
        COUNT=$(cat "$TEMP_DIR/temp.csv" | wc -l)
        for ((o=1; o<=$COUNT; o++))
        do
          LIGNE=$(cat "$TEMP_DIR/temp.csv" | head -$o | tail +$o)
          DEBUT=$(echo "$LIGNE" | cut -d "," -f1,2,3,4)
          SCENEIN=$(grep -e "^${DEBUT}" $TEMP_DIR/temp_rating.csv 2>/dev/null || echo "")
          
            if [[ -n $SCENEIN ]]
              then
              NOTE=$(echo "${SCENEIN}" | cut -d "," -f6 )
              echo "Found rating $NOTE for movie entry: ${DEBUT:0:30}..." | tee -a "${LOG}"
              
              if [[ "$OSTYPE" == "darwin"* ]]; then
                  # macOS version
                  sed -i '' "${o}s|$|$NOTE|" $TEMP_DIR/temp.csv
              else
                  # Linux version
                  sed -i "${o}s|$|$NOTE|" $TEMP_DIR/temp.csv
              fi
            fi
         
        done

        if [ -s "$TEMP_DIR/temp_show.csv" ]; then
            echo "Processing show data for CSV file" | tee -a "${LOG}"
            echo "Found $(cat "$TEMP_DIR/temp_show.csv" | wc -l) shows to process" | tee -a "${LOG}"
            
            COUNT=$(cat "$TEMP_DIR/temp_show.csv" | wc -l)
            for ((o=1; o<=$COUNT; o++))
            do
              LIGNE=$(cat "$TEMP_DIR/temp_show.csv" | head -$o | tail +$o)
              DEBUT=$(echo "$LIGNE" | cut -d "," -f1,2,3,4)
              SCENEIN=$(grep -e "^${DEBUT}" $TEMP_DIR/temp_rating_episodes.csv 2>/dev/null || echo "")
              
                if [[ -n $SCENEIN ]]
                  then
                  NOTE=$(echo "${SCENEIN}" | cut -d "," -f9 )
                  echo "Found rating $NOTE for show entry: ${DEBUT:0:30}..." | tee -a "${LOG}"
                  
                  if [[ "$OSTYPE" == "darwin"* ]]; then
                      # macOS version
                      sed -i '' "${o}s|$|$NOTE|" $TEMP_DIR/temp_show.csv
                  else
                      # Linux version
                      sed -i "${o}s|$|$NOTE|" $TEMP_DIR/temp_show.csv
                  fi
                fi
             
            done    
        fi
        
        if [[ -f "${DOSCOPY}/letterboxd_import.csv" ]]
            then
            echo -e "File exists, new movies will be appended" | tee -a "${LOG}"
            debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Existing CSV file check before appending"
        else
            echo -e "Generating letterboxd_import.csv file" | tee -a "${LOG}"
            # Create the DOSCOPY directory if it doesn't exist
            mkdir -p "$DOSCOPY"
            echo "Created DOSCOPY directory: $DOSCOPY" | tee -a "${LOG}"
            echo "Directory permissions: $(ls -la "$DOSCOPY" | head -n 1 | awk '{print $1}')" | tee -a "${LOG}"
            
            chmod 755 "$DOSCOPY"
            echo "Set directory permissions to 755" | tee -a "${LOG}"
            echo "New directory permissions: $(ls -la "$DOSCOPY" | head -n 1 | awk '{print $1}')" | tee -a "${LOG}"
            
            echo "Title, Year, imdbID, tmdbID, WatchedDate, Rating10" > "${DOSCOPY}/letterboxd_import.csv"
            echo "Created CSV file with header" | tee -a "${LOG}"
            
            chmod 644 "${DOSCOPY}/letterboxd_import.csv"
            echo "Set CSV file permissions to 644" | tee -a "${LOG}"
            
            debug_file_info "${DOSCOPY}/letterboxd_import.csv" "New CSV file creation"
        fi
        
        echo -e "Adding the following data: " | tee -a "${LOG}"
        COUNTTEMP=$(cat "$TEMP_DIR/temp.csv" | wc -l)
        echo "Processing $COUNTTEMP movies to add to CSV" | tee -a "${LOG}"
        
        for ((p=1; p<=$COUNTTEMP; p++))
        do
          LIGNETEMP=$(cat "$TEMP_DIR/temp.csv" | head -$p | tail +$p)
          DEBUT=$(echo "$LIGNETEMP" | cut -d "," -f1,2,3,4)
          DEBUTCOURT=$(echo "$LIGNETEMP" | cut -d "," -f1,2)
          MILIEU=$(echo "$LIGNETEMP" | cut -d "," -f5 | cut -d "T" -f1 | tr -d "\"")
          FIN=$(echo "$LIGNETEMP" | cut -d "," -f6)
          
          echo "Processing movie $p/$COUNTTEMP: ${DEBUTCOURT:0:30}..." | tee -a "${LOG}"
          
          # Check if the CSV file exists before trying to grep from it
          if [ ! -f "${DOSCOPY}/letterboxd_import.csv" ]; then
              echo "ERROR: CSV file doesn't exist at expected location: ${DOSCOPY}/letterboxd_import.csv" | tee -a "${LOG}"
              debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Missing CSV file"
              echo "Attempting to create CSV file" | tee -a "${LOG}"
              
              mkdir -p "$DOSCOPY"
              chmod 755 "$DOSCOPY"
              echo "Title, Year, imdbID, tmdbID, WatchedDate, Rating10" > "${DOSCOPY}/letterboxd_import.csv"
              chmod 644 "${DOSCOPY}/letterboxd_import.csv"
              
              debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Created CSV file"
          fi
          
          # Check if file is searchable before grepping
          if [ ! -r "${DOSCOPY}/letterboxd_import.csv" ]; then
              echo "ERROR: CSV file is not readable: ${DOSCOPY}/letterboxd_import.csv" | tee -a "${LOG}"
              debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Non-readable CSV file"
              
              # Try to fix permissions
              chmod 644 "${DOSCOPY}/letterboxd_import.csv"
              echo "Attempted to fix file permissions" | tee -a "${LOG}"
              debug_file_info "${DOSCOPY}/letterboxd_import.csv" "After fixing CSV file permissions"
          fi
          
          SCENEIN1=$(grep -e "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv 2>/dev/null || echo "")
          
            if [[ -n $SCENEIN1 ]]
              then
              FIN1=$(echo "$SCENEIN1" | cut -d "," -f6)
              SCENEIN2=$(grep -n "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv | cut -d ":" -f 1)
              
              if [[ "${DEBUT},${MILIEU},${FIN}" == "${DEBUT},${MILIEU},${FIN1}" ]]
                then
                echo "Movie: ${DEBUTCOURT} already present in import file" | tee -a "${LOG}"
              else
                if [[ -n $FIN1 ]]
                  then
                  echo "Movie: ${DEBUTCOURT} found with rating $FIN1, updating to $FIN" | tee -a "${LOG}"
                  
                  if [[ "$OSTYPE" == "darwin"* ]]; then
                      # macOS version
                      sed -i '' "${SCENEIN2}s/$FIN1/$FIN/" ${DOSCOPY}/letterboxd_import.csv
                      echo "Used macOS sed to update rating" | tee -a "${LOG}"
                  else
                      # Linux version
                      sed -i "${SCENEIN2}s/$FIN1/$FIN/" ${DOSCOPY}/letterboxd_import.csv
                      echo "Used Linux sed to update rating" | tee -a "${LOG}"
                  fi
                  
                  # Verify the update was successful
                  VERIFY=$(grep -e "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv | grep -e "$FIN" || echo "")
                  if [[ -n $VERIFY ]]; then
                      echo "Rating update verified successfully" | tee -a "${LOG}"
                  else
                      echo "WARNING: Rating update could not be verified" | tee -a "${LOG}"
                      echo "Current line in CSV: $(grep -e "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv || echo "Not found")" | tee -a "${LOG}"
                  fi
                  
                else
                  echo "Movie: ${DEBUTCOURT} found without rating, adding rating $FIN" | tee -a "${LOG}"
                  
                  if [[ "$OSTYPE" == "darwin"* ]]; then
                      # macOS version
                      sed -i '' "${SCENEIN2}s/$/$FIN/" ${DOSCOPY}/letterboxd_import.csv
                      echo "Used macOS sed to add rating" | tee -a "${LOG}"
                  else
                      # Linux version
                      sed -i "${SCENEIN2}s/$/$FIN/" ${DOSCOPY}/letterboxd_import.csv
                      echo "Used Linux sed to add rating" | tee -a "${LOG}"
                  fi
                  
                  # Verify the update was successful
                  VERIFY=$(grep -e "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv | grep -e "$FIN" || echo "")
                  if [[ -n $VERIFY ]]; then
                      echo "Rating addition verified successfully" | tee -a "${LOG}"
                  else
                      echo "WARNING: Rating addition could not be verified" | tee -a "${LOG}"
                      echo "Current line in CSV: $(grep -e "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv || echo "Not found")" | tee -a "${LOG}"
                  fi
                fi
                echo "Updated: ${DEBUTCOURT} rating to $FIN" | tee -a "${LOG}"
              fi
            else
              echo "New movie: ${DEBUTCOURT} with rating $FIN" | tee -a "${LOG}"
              echo "${DEBUT},${MILIEU},${FIN}" | tee -a "${LOG}" >> "${DOSCOPY}/letterboxd_import.csv"
              
              # Verify the addition was successful
              VERIFY=$(grep -e "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv || echo "")
              if [[ -n $VERIFY ]]; then
                  echo "Addition verified successfully" | tee -a "${LOG}"
              else
                  echo "WARNING: Addition could not be verified" | tee -a "${LOG}"
                  debug_file_info "${DOSCOPY}/letterboxd_import.csv" "After attempting to add new entry"
                  
                  # Try direct file write with cat
                  echo "Attempting alternative write method..." | tee -a "${LOG}"
                  echo "${DEBUT},${MILIEU},${FIN}" > /tmp/new_entry.tmp
                  cat /tmp/new_entry.tmp >> "${DOSCOPY}/letterboxd_import.csv"
                  rm /tmp/new_entry.tmp
                  
                  VERIFY=$(grep -e "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv || echo "")
                  if [[ -n $VERIFY ]]; then
                      echo "Addition succeeded with alternative method" | tee -a "${LOG}"
                  else
                      echo "ERROR: Failed to add entry with alternative method" | tee -a "${LOG}"
                  fi
              fi
            fi  
        done
        
        echo " " | tee -a "${LOG}"
        echo -e "File letterboxd_import.csv created in directory $DOSCOPY" | tee -a "${LOG}"
        debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Final CSV file check"
        echo -e "${BOLD}To be imported at: https://letterboxd.com/import/ ${NC}" | tee -a "${LOG}"
        echo " " | tee -a "${LOG}"
        echo -e "${BOLD}Don't forget to delete the csv file!!! ${NC}" | tee -a "${LOG}"
    fi

  # Process temporary files only if they exist and contain data
  if [ -s "$TEMP_DIR/temp.csv" ]; then
      awk -F, 'BEGIN {OFS=","} {gsub(/"/, "", $1); $2=$2",NULL,NULL,NULL"}1' $TEMP_DIR/temp.csv > $TEMP_DIR/temp2.csv
      if [[ "$OSTYPE" == "darwin"* ]]; then
          # macOS version
          sed -i '' 's/^/Movie,/; s/"//g' $TEMP_DIR/temp2.csv
      else
          # Linux version
          sed -i 's/^/Movie,/; s/"//g' $TEMP_DIR/temp2.csv
      fi
  else
      # Create an empty file
      touch $TEMP_DIR/temp2.csv
  fi
  
  if [ -s "$TEMP_DIR/temp_show.csv" ]; then
      if [[ "$OSTYPE" == "darwin"* ]]; then
          # macOS version
          sed -i '' 's/^/Show,/; s/"//g' $TEMP_DIR/temp_show.csv
      else
          # Linux version
          sed -i 's/^/Show,/; s/"//g' $TEMP_DIR/temp_show.csv
      fi
  fi
  
  if [ -s "$TEMP_DIR/temp_watchlist.csv" ]; then
      if [[ "$OSTYPE" == "darwin"* ]]; then
          # macOS version
          sed -i '' 's/"//g' $TEMP_DIR/temp_watchlist.csv
      else
          # Linux version
          sed -i 's/"//g' $TEMP_DIR/temp_watchlist.csv
      fi
  fi

  # Define BRAIN_OPS if it's not defined
  if [ -z "${BRAIN_OPS}" ]; then
    BRAIN_OPS="${SCRIPT_DIR}"
    echo "BRAIN_OPS variable not defined, using current directory: ${BRAIN_OPS}" | tee -a "${LOG}"
  fi

  # Check if the directory is writable
  if [ ! -w "${BRAIN_OPS}" ]; then
    echo "The directory ${BRAIN_OPS} is not writable. Using current directory." | tee -a "${LOG}"
    BRAIN_OPS="${SCRIPT_DIR}"
  fi

  # Create the BRAIN_OPS directory if it doesn't exist
  mkdir -p "${BRAIN_OPS}"

  # Write to output files only if temporary files exist and contain data
  if [ -s "$TEMP_DIR/temp2.csv" ]; then
      cat $TEMP_DIR/temp2.csv >> ${BRAIN_OPS}/watched_${DATE}.csv
  fi
  
  if [ -s "$TEMP_DIR/temp_show.csv" ]; then
      cat $TEMP_DIR/temp_show.csv >> ${BRAIN_OPS}/watched_${DATE}.csv
  fi
  
  if [ -s "$TEMP_DIR/temp_watchlist.csv" ]; then
      cat $TEMP_DIR/temp_watchlist.csv >> ${BRAIN_OPS}/watchlist_${DATE}.csv
  fi
fi
#rm -r ${BACKUP_DIR}/
#rm -r ${SCRIPT_DIR}/TEMP/

# Clean up temporary files
echo "Cleaning up temporary files in $TEMP_DIR" | tee -a "${LOG}"
rm -rf "$TEMP_DIR"
echo "Temporary files cleaned up: $(if [ ! -d "$TEMP_DIR" ]; then echo "✅"; else echo "❌"; fi)" | tee -a "${LOG}"

# Final verification of CSV file
echo "🏁 Final verification:" | tee -a "${LOG}"
debug_file_info "${DOSCOPY}/letterboxd_import.csv" "Final CSV file verification"

echo "Script execution completed at $(date)" | tee -a "${LOG}"
echo "=====================================================" | tee -a "${LOG}"
