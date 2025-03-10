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

# Detect OS for sed compatibility
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS uses BSD sed
    SED_INPLACE="sed -i ''"
else
    # Linux and others use GNU sed
    SED_INPLACE="sed -i"
fi

# Always use the config file from the config directory
CONFIG_DIR="${SCRIPT_DIR}/config"
if [ -f "/app/config/.config.cfg" ]; then
    # If running in Docker, use the absolute path
    source /app/config/.config.cfg
else
    # If running locally, use the relative path
    source ${CONFIG_DIR}/.config.cfg
fi

# Use the user's temporary directory
TEMP_DIR="/tmp/trakt_export_$USER"
rm -rf "$TEMP_DIR"
mkdir -p "$TEMP_DIR"

if [ ! -d $DOSLOG ]
	then
	mkdir -p $DOSLOG
fi


refresh_access_token() {
    echo "🔄 Refreshing Trakt token..." | tee -a "${LOG}"
    
    RESPONSE=$(curl -s -X POST "https://api.trakt.tv/oauth/token" \
        -H "Content-Type: application/json" -v \
        -d "{
            \"refresh_token\": \"${REFRESH_TOKEN}\",
            \"client_id\": \"${API_KEY}\",
            \"client_secret\": \"${API_SECRET}\",
            \"redirect_uri\": \"${REDIRECT_URI}\",
            \"grant_type\": \"refresh_token\"
        }")

    NEW_ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
    NEW_REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refresh_token')

    if [[ "$NEW_ACCESS_TOKEN" != "null" && "$NEW_REFRESH_TOKEN" != "null" ]]; then
        echo "✅ Token refreshed successfully." | tee -a "${LOG}"
        
        # Determine which config file to update
        CONFIG_FILE="/app/config/.config.cfg"
        if [ ! -f "$CONFIG_FILE" ]; then
            CONFIG_FILE="${CONFIG_DIR}/.config.cfg"
        fi
        
        $SED_INPLACE "s|ACCESS_TOKEN=.*|ACCESS_TOKEN=\"$NEW_ACCESS_TOKEN\"|" "$CONFIG_FILE"
        $SED_INPLACE "s|REFRESH_TOKEN=.*|REFRESH_TOKEN=\"$NEW_REFRESH_TOKEN\"|" "$CONFIG_FILE"
        
        # Re-source the config file to update variables
        source "$CONFIG_FILE"
    else
        echo "❌ Error refreshing token. Check your configuration!" | tee -a "${LOG}"
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
            chmod 755 "$DOSCOPY"
            echo "Title, Year, imdbID, tmdbID, WatchedDate, Rating10" > "${DOSCOPY}/letterboxd_import.csv"
            chmod 644 "${DOSCOPY}/letterboxd_import.csv"
        fi
    else
        COUNT=$(cat "$TEMP_DIR/temp.csv" | wc -l)
        for ((o=1; o<=$COUNT; o++))
        do
          LIGNE=$(cat "$TEMP_DIR/temp.csv" | head -$o | tail +$o)
          DEBUT=$(echo "$LIGNE" | cut -d "," -f1,2,3,4)
          SCENEIN=$(grep -e "^${DEBUT}" $TEMP_DIR/temp_rating.csv 2>/dev/null || echo "")
          
            if [[ -n $SCENEIN ]]
              then
              NOTE=$(echo "${SCENEIN}" | cut -d "," -f6 )
              
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
            COUNT=$(cat "$TEMP_DIR/temp_show.csv" | wc -l)
            for ((o=1; o<=$COUNT; o++))
            do
              LIGNE=$(cat "$TEMP_DIR/temp_show.csv" | head -$o | tail +$o)
              DEBUT=$(echo "$LIGNE" | cut -d "," -f1,2,3,4)
              SCENEIN=$(grep -e "^${DEBUT}" $TEMP_DIR/temp_rating_episodes.csv 2>/dev/null || echo "")
              
                if [[ -n $SCENEIN ]]
                  then
                  NOTE=$(echo "${SCENEIN}" | cut -d "," -f9 )
                  
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
        else
            echo -e "Generating letterboxd_import.csv file" | tee -a "${LOG}"
            # Create the DOSCOPY directory if it doesn't exist
            mkdir -p "$DOSCOPY"
            chmod 755 "$DOSCOPY"
            echo "Title, Year, imdbID, tmdbID, WatchedDate, Rating10" > "${DOSCOPY}/letterboxd_import.csv"
            chmod 644 "${DOSCOPY}/letterboxd_import.csv"
        fi
        echo -e "Adding the following data: " | tee -a "${LOG}"
        COUNTTEMP=$(cat "$TEMP_DIR/temp.csv" | wc -l)
        for ((p=1; p<=$COUNTTEMP; p++))
        do
          LIGNETEMP=$(cat "$TEMP_DIR/temp.csv" | head -$p | tail +$p)
          DEBUT=$(echo "$LIGNETEMP" | cut -d "," -f1,2,3,4)
          #echo "debut $DEBUT"
          DEBUTCOURT=$(echo "$LIGNETEMP" | cut -d "," -f1,2)
          MILIEU=$(echo "$LIGNETEMP" | cut -d "," -f5 | cut -d "T" -f1 | tr -d "\"")
          #echo "MILIEU $MILIEU"
          FIN=$(echo "$LIGNETEMP" | cut -d "," -f6)
         # echo "FIN $FIN"
          SCENEIN1=$(grep -e "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv 2>/dev/null || echo "")
          
          #echo "SCENEIN1 $SCENEIN1"
            if [[ -n $SCENEIN1 ]]
              then
              FIN1=$(echo "$SCENEIN1" | cut -d "," -f6)
              #echo "fin1 $FIN1"
              SCENEIN2=$(grep -n "^${DEBUT},${MILIEU}" ${DOSCOPY}/letterboxd_import.csv | cut -d ":" -f 1)
              #echo "scenein2 $SCENEIN2"
              if [[ "${DEBUT},${MILIEU},${FIN}" == "${DEBUT},${MILIEU},${FIN1}" ]]
                then
                echo "Movie: ${DEBUTCOURT} already present in import file" | tee -a "${LOG}"
              else
                #FIN2=$(echo "$SCENEIN2" | cut -d "," -f6)
                if [[ -n $FIN1 ]]
                  then
                  if [[ "$OSTYPE" == "darwin"* ]]; then
                      # macOS version
                      sed -i '' "${SCENEIN2}s/$FIN1/$FIN/" ${DOSCOPY}/letterboxd_import.csv
                  else
                      # Linux version
                      sed -i "${SCENEIN2}s/$FIN1/$FIN/" ${DOSCOPY}/letterboxd_import.csv
                  fi
                  else
                  if [[ "$OSTYPE" == "darwin"* ]]; then
                      # macOS version
                      sed -i '' "${SCENEIN2}s/$/$FIN/" ${DOSCOPY}/letterboxd_import.csv
                  else
                      # Linux version
                      sed -i "${SCENEIN2}s/$/$FIN/" ${DOSCOPY}/letterboxd_import.csv
                  fi
                  fi
                echo "Movie: ${DEBUTCOURT} already present but adding rating $FIN" | tee -a "${LOG}"
              fi
            else
              echo "${DEBUT},${MILIEU},${FIN}"
              echo "${DEBUT},${MILIEU},${FIN}" | tee -a "${LOG}" >> "${DOSCOPY}/letterboxd_import.csv"
            fi  
        done
        
        echo " " | tee -a "${LOG}"
        echo -e "File letterboxd_import.csv created in directory $DOSCOPY" | tee -a "${LOG}"
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
rm -rf "$TEMP_DIR"
