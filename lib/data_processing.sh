#!/bin/bash
#
# Data processing functions for Trakt to Letterboxd export
#

# Create ratings lookup file from ratings_movies.json
create_ratings_lookup() {
    local ratings_file="$1"
    local output_file="$2"
    local log_file="$3"
    
    if [ -f "$ratings_file" ] && [ -s "$ratings_file" ]; then
        echo "DEBUG: Creating ratings lookup file..." | tee -a "${log_file}"
        
        # Dump the first few items for debugging
        echo "DEBUG: First few items in ratings file:" | tee -a "${log_file}"
        jq -r '.[:3] | map({id: .movie.ids.trakt, title: .movie.title, rating: .rating})' "$ratings_file" | tee -a "${log_file}"
        
        # Count ratings for debugging
        local rating_count=$(jq '. | length' "$ratings_file" 2>/dev/null)
        echo "📊 Found $rating_count ratings in file" | tee -a "${log_file}"
        
        # Verify JSON is valid before processing
        if ! jq empty "$ratings_file" 2>/dev/null; then
            echo "⚠️ WARNING: Invalid JSON in ratings file, creating empty lookup" | tee -a "${log_file}"
            echo "{}" > "$output_file"
            return 1
        fi
        
        # Ensure output directory exists
        local output_dir=$(dirname "$output_file")
        mkdir -p "$output_dir" 2>/dev/null
        
        # Extract a sample rating to verify structure
        jq -r 'first | {title: .movie.title, id: .movie.ids.trakt, rating: .rating} | tostring' "$ratings_file" > /dev/null 2>&1
        if [ $? -ne 0 ]; then
            echo "⚠️ WARNING: Unexpected JSON structure in ratings file" | tee -a "${log_file}"
            echo "⚠️ Using alternative approach to extract ratings" | tee -a "${log_file}"
            
            # Alternative approach - more direct, less prone to structure issues
            jq -c '{} as $result | reduce .[] as $item ($result; 
                if $item.movie and $item.movie.ids and $item.movie.ids.trakt and $item.rating then 
                    .[$item.movie.ids.trakt | tostring] = $item.rating 
                else . end)' "$ratings_file" > "$output_file" 2>/dev/null
            
            if [ $? -ne 0 ]; then
                echo "⚠️ WARNING: Alternative approach failed, creating basic lookup" | tee -a "${log_file}"
                # Print a sample of the JSON for debugging
                echo "DEBUG: Sample of ratings file:" | tee -a "${log_file}"
                jq -r 'first | tostring' "$ratings_file" | tee -a "${log_file}"
                echo "{}" > "$output_file"
                return 1
            fi
        else
            # Original approach if structure is as expected
            jq -c 'reduce .[] as $item ({}; 
                if $item.movie and $item.movie.ids and $item.movie.ids.trakt != null then 
                    .[$item.movie.ids.trakt | tostring] = $item.rating 
                else . end)' "$ratings_file" > "$output_file" 2>/dev/null
            
            if [ $? -ne 0 ]; then
                echo "⚠️ WARNING: Failed to process ratings file, creating empty lookup" | tee -a "${log_file}"
                echo "{}" > "$output_file"
                return 1
            fi
        fi
        
        # Verify the lookup file was created successfully
        if [ ! -s "$output_file" ]; then
            echo "⚠️ WARNING: Ratings lookup file is empty, creating basic JSON" | tee -a "${log_file}"
            echo "{}" > "$output_file"
            return 1
        fi
        
        # Check content of created lookup file
        local lookup_entries=$(jq 'length' "$output_file" 2>/dev/null || echo "0")
        echo "📊 Created ratings lookup with $lookup_entries entries" | tee -a "${log_file}"
        
        # Show a sample for debugging
        echo "📊 Sample ratings lookup:" | tee -a "${log_file}"
        jq -r 'to_entries | .[0:3] | map("\(.key):\(.value)") | join(", ")' "$output_file" 2>/dev/null | tee -a "${log_file}"
        
        # Save a map of Trakt IDs to movie titles for easier debugging
        local title_map_file="${output_file}.titles.json"
        jq -c 'reduce .[] as $item ({}; 
            if $item.movie and $item.movie.ids and $item.movie.ids.trakt != null then 
                .[$item.movie.ids.trakt | tostring] = $item.movie.title 
            else . end)' "$ratings_file" > "$title_map_file" 2>/dev/null
        
        # Create a lookup for recent films we're specifically interested in
        local recent_films_file="${output_file}.recent.json"
        jq -c '[.[] | select(.movie.title | test("Jumanji: The Next Level|The Alto Knights|Paddington in Peru|The Gorge|Mickey 17|God Save The Tuches")) | {id: .movie.ids.trakt | tostring, title: .movie.title, rating: .rating}]' "$ratings_file" > "$recent_films_file" 2>/dev/null
        
        if [ -s "$recent_films_file" ]; then
            echo "📊 Recent films ratings:" | tee -a "${log_file}"
            cat "$recent_films_file" | tee -a "${log_file}"
        fi
        
        return 0
    else
        echo "WARNING: Ratings file not found or empty, creating empty lookup" | tee -a "${log_file}"
        echo "{}" > "$output_file"
        return 1
    fi
}

# Create plays count lookup from watched_movies.json
create_plays_count_lookup() {
    local watched_file="$1"
    local output_file="$2"
    local log_file="$3"
    
    if [ -f "$watched_file" ] && [ -s "$watched_file" ]; then
        echo "DEBUG: Creating plays count lookup from watched_movies..." | tee -a "${log_file}"
        
        # Verify JSON is valid before processing
        if ! jq empty "$watched_file" 2>/dev/null; then
            echo "⚠️ WARNING: Invalid JSON in watched file, creating empty lookup" | tee -a "${log_file}"
            echo "{}" > "$output_file"
            return 1
        fi
        
        # Ensure output directory exists
        local output_dir=$(dirname "$output_file")
        mkdir -p "$output_dir" 2>/dev/null
        
        # Create the lookup with proper error handling - using IMDB ID as key
        if ! jq -c 'reduce .[] as $item ({}; if $item.movie.ids.imdb != null then .[$item.movie.ids.imdb] = $item.plays else . end)' "$watched_file" > "$output_file" 2>/dev/null; then
            echo "⚠️ WARNING: Failed to process watched file, creating empty lookup" | tee -a "${log_file}"
            echo "{}" > "$output_file"
            return 1
        fi
        
        # Verify the lookup file was created successfully
        if [ ! -s "$output_file" ]; then
            echo "⚠️ WARNING: Plays count lookup file is empty, creating basic JSON" | tee -a "${log_file}"
            echo "{}" > "$output_file"
            return 1
        fi
        
        return 0
    else
        echo "WARNING: Watched movies file not found or empty, creating empty lookup" | tee -a "${log_file}"
        echo "{}" > "$output_file"
        return 1
    fi
}

# Process history movies from history_movies JSON file
process_history_movies() {
    local history_file="$1"
    local ratings_lookup="$2"
    local plays_lookup="$3"
    local csv_output="$4"
    local raw_output="$5"
    local log="$6"
    
    # Check if history file exists
    if [ ! -f "$history_file" ]; then
        echo -e "⚠️ Movies history: No history_movies.json file found" | tee -a "${log}"
        return 1
    fi
    
    if [ ! -s "$history_file" ]; then
        echo -e "⚠️ Movies history: history_movies.json file is empty" | tee -a "${log}"
        return 1
    fi
    
    # Verify if the JSON is valid
    if ! jq empty "$history_file" 2>/dev/null; then
        echo -e "⚠️ Movies history: Invalid JSON in history_movies.json" | tee -a "${log}"
        return 1
    fi
    
    # Check ratings lookup file
    if [ -f "$ratings_lookup" ] && [ -s "$ratings_lookup" ]; then
        echo -e "📊 Ratings lookup file found: $(wc -c < "$ratings_lookup") bytes" | tee -a "${log}"
        echo -e "📊 Sample ratings: $(jq -r 'to_entries | .[0:3] | map("\(.key):\(.value)") | join(", ")' "$ratings_lookup" 2>/dev/null || echo "Error reading ratings")" | tee -a "${log}"
        
        # Debug: Verify if specific Trakt IDs exist in the lookup
        local debug_ids=("360095" "814646" "915974")  # IDs from previous verification
        echo -e "DEBUG: Verifying presence of known ratings IDs:" | tee -a "${log}"
        for id in "${debug_ids[@]}"; do
            local found_rating=$(jq -r --arg id "$id" '.[$id] // "not found"' "$ratings_lookup" 2>/dev/null)
            echo -e "DEBUG: ID $id rating: $found_rating" | tee -a "${log}"
        done
    else
        echo -e "⚠️ Ratings lookup file not found or empty: $ratings_lookup" | tee -a "${log}"
        # Try to find ratings file in same directory as history file
        local ratings_file="${history_file%/*}/johandev-ratings_movies.json"
        if [ -f "$ratings_file" ] && [ -s "$ratings_file" ]; then
            echo -e "🔍 Found ratings file: $ratings_file" | tee -a "${log}"
            # Create a temporary ratings lookup file directly
            local temp_ratings_lookup=$(mktemp)
            jq -c '{} as $result | reduce .[] as $item ($result; 
                if $item.movie and $item.movie.ids and $item.movie.ids.trakt and $item.rating then 
                    .[$item.movie.ids.trakt | tostring] = $item.rating 
                else . end)' "$ratings_file" > "$temp_ratings_lookup" 2>/dev/null
            if [ -s "$temp_ratings_lookup" ]; then
                echo -e "✅ Created temporary ratings lookup from ratings file" | tee -a "${log}"
                ratings_lookup="$temp_ratings_lookup"
            fi
        fi
    fi
    
    # Get number of movies in history
    local movie_count=$(jq length "$history_file")
    if [ "$movie_count" -eq 0 ]; then
        echo -e "⚠️ Movies history: No movies found in history" | tee -a "${log}"
        return 1
    fi
    
    echo "DEBUG: Processing history_movies file with ratings..." >> "${log}"
    
    # Create temporary files for processed data
    local tmp_file=$(mktemp)
    local ratings_direct_file=$(mktemp)
    
    # Extract ratings directly from history file if possible
    jq -r '.[] | 
        select(.movie and .movie.ids and .movie.ids.trakt) | 
        [.movie.ids.trakt, (.rating // "")]| 
        @tsv' "$history_file" > "$ratings_direct_file" 2>/dev/null
    
    # Extract all raw data with ratings included
    jq -r '.[] | 
        # Handle null values safely with defaults
        . as $item | 
        {
            title: (try ($item.movie.title) catch null // "Unknown Title"),
            year: (try ($item.movie.year | tostring) catch null // ""),
            imdb_id: (try ($item.movie.ids.imdb) catch null // ""),
            tmdb_id: (try ($item.movie.ids.tmdb | tostring) catch null // ""),
            watched_at: (try ($item.watched_at) catch null // ""),
            trakt_id: (try ($item.movie.ids.trakt | tostring) catch null // ""),
            rating: (try ($item.rating | tostring) catch null // "")
        } | 
        # Build CSV line with safe values
        [.title, .year, .imdb_id, .tmdb_id, .watched_at, .trakt_id, .rating] | 
        @csv' "$history_file" > "$tmp_file"
    
    # Handle errors in jq processing
    if [ $? -ne 0 ]; then
        echo -e "⚠️ Error processing history_movies.json with jq" | tee -a "${log}"
        echo -e "Attempting alternative processing method..." | tee -a "${log}"
        
        # Alternative processing using a simpler jq command
        jq -r '.[] | 
            [
                (.movie.title // "Unknown Title"),
                (.movie.year // ""),
                (.movie.ids.imdb // ""),
                (.movie.ids.tmdb // ""),
                (.watched_at // ""),
                (.movie.ids.trakt // ""),
                (.rating // "")
            ] | 
            @csv' "$history_file" > "$tmp_file" 2>>"${log}"
        
        if [ $? -ne 0 ]; then
            echo -e "⚠️ Alternative processing also failed" | tee -a "${log}"
            return 1
        fi
    fi
    
    # Check if tmp_file was created and has content
    if [ ! -s "$tmp_file" ]; then
        echo -e "⚠️ Error: No data extracted from history_movies.json" | tee -a "${log}"
        return 1
    fi
    
    # Create raw output with headers
    echo "Title,Year,imdbID,tmdbID,WatchedDate,TraktID,Rating" > "$raw_output"
    cat "$tmp_file" >> "$raw_output"
    
    # If ratings_file is available, extract all ratings into a lookup map
    local title_ratings={}
    local imdb_ratings={}
    local tmdb_ratings={}
    local ratings_file="${history_file%/*}/johandev-ratings_movies.json"
    
    if [ -f "$ratings_file" ] && [ -s "$ratings_file" ]; then
        echo "DEBUG: Found ratings file at $ratings_file, parsing additional ratings..." | tee -a "${log}"
        # Create lookup maps by title, IMDb ID, and TMDB ID
        title_ratings=$(jq -c 'reduce .[] as $item ({}; 
            if $item.movie.title != null then .[$item.movie.title] = $item.rating else . end)' "$ratings_file" 2>/dev/null)
        
        imdb_ratings=$(jq -c 'reduce .[] as $item ({}; 
            if $item.movie.ids.imdb != null then .[$item.movie.ids.imdb] = $item.rating else . end)' "$ratings_file" 2>/dev/null)
        
        tmdb_ratings=$(jq -c 'reduce .[] as $item ({}; 
            if $item.movie.ids.tmdb != null then .[$item.movie.ids.tmdb | tostring] = $item.rating else . end)' "$ratings_file" 2>/dev/null)
        
        echo "DEBUG: Created additional lookups from ratings file" | tee -a "${log}"
    fi
    
    # Process the extracted data to add ratings and deduplicate
    local processed_count=0
    local ratings_found=0
    local existing_ids=()
    
    # Add CSV header to output
    echo "Title,Year,imdbID,tmdbID,WatchedDate,Rating10,Rewatch" > "$csv_output"
    
    while IFS=, read -r title year imdb_id tmdb_id watched_at trakt_id direct_rating; do
        # Clean quotes from CSV fields if present
        title=$(echo "$title" | sed -e 's/^"//' -e 's/"$//')
        year=$(echo "$year" | sed -e 's/^"//' -e 's/"$//')
        imdb_id=$(echo "$imdb_id" | sed -e 's/^"//' -e 's/"$//')
        tmdb_id=$(echo "$tmdb_id" | sed -e 's/^"//' -e 's/"$//')
        watched_at=$(echo "$watched_at" | sed -e 's/^"//' -e 's/"$//')
        trakt_id=$(echo "$trakt_id" | sed -e 's/^"//' -e 's/"$//')
        direct_rating=$(echo "$direct_rating" | sed -e 's/^"//' -e 's/"$//')
        
        # Skip entries with missing key data
        if [ -z "$title" ] || [ "$title" = "null" ] || [ -z "$watched_at" ] || [ "$watched_at" = "null" ]; then
            continue
        fi
        
        # Format watched date (keep only YYYY-MM-DD)
        watched_at=$(echo "$watched_at" | awk -F'T' '{print $1}')
        
        # Get rating from multiple sources with priority:
        # 1. Direct rating from history item
        # 2. Ratings lookup from trakt_id
        # 3. Title, IMDb, or TMDB lookups
        rating=""
        
        # Method 1: Check if we have a direct rating
        if [ -n "$direct_rating" ] && [ "$direct_rating" != "null" ]; then
            rating="$direct_rating"
            ((ratings_found++))
            echo "DEBUG: Using direct rating for $title: $rating" | tee -a "${log}"
        fi
        
        # Method 2: Check in ratings lookup by Trakt ID
        if [ -z "$rating" ] && [ -f "$ratings_lookup" ] && [ -n "$trakt_id" ] && [ "$trakt_id" != "null" ]; then
            local lookup_rating=$(jq -r --arg id "$trakt_id" '.[$id] // ""' "$ratings_lookup" 2>/dev/null)
            if [ -n "$lookup_rating" ] && [ "$lookup_rating" != "null" ]; then
                rating="$lookup_rating"
                ((ratings_found++))
                echo "DEBUG: Found rating from lookup for $title ($trakt_id): $rating" | tee -a "${log}"
            fi
        fi
        
        # Method 3: Check in ratings directly from the ratings file by title
        if [ -z "$rating" ] && [ -n "$title_ratings" ] && [ "$title_ratings" != "{}" ]; then
            local title_rating=$(echo "$title_ratings" | jq -r --arg title "$title" '.[$title] // ""')
            if [ -n "$title_rating" ] && [ "$title_rating" != "null" ]; then
                rating="$title_rating"
                ((ratings_found++))
                echo "DEBUG: Found rating by title for $title: $rating" | tee -a "${log}"
            fi
        fi
        
        # Method 4: Check by IMDb ID
        if [ -z "$rating" ] && [ -n "$imdb_ratings" ] && [ "$imdb_ratings" != "{}" ] && [ -n "$imdb_id" ] && [ "$imdb_id" != "null" ]; then
            local imdb_rating=$(echo "$imdb_ratings" | jq -r --arg id "$imdb_id" '.[$id] // ""')
            if [ -n "$imdb_rating" ] && [ "$imdb_rating" != "null" ]; then
                rating="$imdb_rating"
                ((ratings_found++))
                echo "DEBUG: Found rating by IMDb ID for $title: $rating" | tee -a "${log}"
            fi
        fi
        
        # Method 5: Check by TMDB ID
        if [ -z "$rating" ] && [ -n "$tmdb_ratings" ] && [ "$tmdb_ratings" != "{}" ] && [ -n "$tmdb_id" ] && [ "$tmdb_id" != "null" ]; then
            local tmdb_rating=$(echo "$tmdb_ratings" | jq -r --arg id "$tmdb_id" '.[$id] // ""')
            if [ -n "$tmdb_rating" ] && [ "$tmdb_rating" != "null" ]; then
                rating="$tmdb_rating"
                ((ratings_found++))
                echo "DEBUG: Found rating by TMDB ID for $title: $rating" | tee -a "${log}"
            fi
        fi
        
        # Method 6: Direct lookup from ratings file if all else fails
        if [ -z "$rating" ] && [ -f "${history_file%/*}/johandev-ratings_movies.json" ] && [ -n "$trakt_id" ] && [ "$trakt_id" != "null" ]; then
            rating=$(jq -r --arg tid "$trakt_id" '.[] | select(.movie.ids.trakt | tostring == $tid) | .rating' "${history_file%/*}/johandev-ratings_movies.json" 2>/dev/null | head -n1)
            if [ -n "$rating" ] && [ "$rating" != "null" ]; then
                ((ratings_found++))
                echo "DEBUG: Found direct rating for $title ($trakt_id): $rating" | tee -a "${log}"
            fi
        fi
        
        # Get plays count for rewatch flag
        rewatch="false"
        if [ -n "$trakt_id" ] && [ "$trakt_id" != "null" ]; then
            play_count=$(jq -r --arg id "$trakt_id" '.[$id] // "0"' "$plays_lookup" 2>/dev/null)
            if [ -n "$play_count" ] && [ "$play_count" != "null" ] && [ "$play_count" -gt 1 ]; then
                rewatch="true"
            fi
        fi
        
        # Add data to CSV file if not a duplicate
        # Filter by unique imdb or tmdb ids if available, otherwise by title+year
        local item_id=""
        if [ -n "$imdb_id" ] && [ "$imdb_id" != "null" ]; then
            item_id="imdb:$imdb_id"
        elif [ -n "$tmdb_id" ] && [ "$tmdb_id" != "null" ]; then
            item_id="tmdb:$tmdb_id"
        elif [ -n "$title" ] && [ -n "$year" ]; then
            item_id="title:$title:$year"
        else
            # Skip if no identifier available
            continue
        fi
        
        # Check if movie is already processed (simple deduplication)
        if [[ " ${existing_ids[@]} " =~ " ${item_id} " ]]; then
            continue
        fi
        
        # Ensure proper quoting and format for IMDb ID
        if [[ ! "$imdb_id" =~ ^tt ]] && [ -n "$imdb_id" ] && [ "$imdb_id" != "null" ]; then
            imdb_id="tt$imdb_id"
        fi
        
        # Ensure proper quoting and format
        echo "\"$title\",\"$year\",\"$imdb_id\",\"$tmdb_id\",\"$watched_at\",\"$rating\",\"$rewatch\"" >> "$csv_output"
        
        # Add to processed IDs
        existing_ids+=("$item_id")
        ((processed_count++))
    done < "$tmp_file"
    
    # Clean up temporary files
    rm -f "$tmp_file" "$ratings_direct_file"
    
    # Report results
    echo "Movies history: $processed_count movies processed" | tee -a "${log}"
    echo "Ratings found: $ratings_found ratings added to CSV" | tee -a "${log}"
    
    if [ "$processed_count" -eq 0 ]; then
        echo -e "⚠️ Movies history: No valid movies extracted from history" | tee -a "${log}"
        return 1
    fi
    
    return 0
}

# Process watched movies (used in complete mode or when history is missing)
process_watched_movies() {
    local watched_file="$1"
    local ratings_lookup="$2"
    local output_csv="$3"
    local existing_ids_file="$4"
    local raw_output_file="$5" 
    local is_fallback="$6"  # true if this is a fallback for missing history
    local log_file="$7"
    
    if [ ! -f "$watched_file" ]; then
        echo "WARNING: Watched movies file not found, skipping" | tee -a "${log_file}"
        return 1
    fi
    
    # Count total movies in the watched file
    local watched_count=$(jq '. | length' "$watched_file" 2>/dev/null || echo "0")
    echo "📊 Found $watched_count movies in watched file" | tee -a "${log_file}"
    
    # Check if ratings_lookup exists to avoid jq errors
    local has_ratings=false
    if [ -f "$ratings_lookup" ] && [ -s "$ratings_lookup" ]; then
        has_ratings=true
        echo "📊 Using ratings lookup for watched movies processing" | tee -a "${log_file}"
        echo "📊 Sample ratings: $(jq -r 'to_entries | .[0:3] | map("\(.key):\(.value)") | join(", ")' "$ratings_lookup" 2>/dev/null || echo "Error reading ratings")" | tee -a "${log_file}"
    else
        echo "⚠️ WARNING: Ratings lookup file not found or empty, proceeding without ratings" | tee -a "${log_file}"
    fi
    
    # Create a temporary file for the extracted data
    local tmp_file=$(mktemp)
    
    if [ "$is_fallback" = "true" ]; then
        echo "DEBUG: No history found. Processing watched_movies file with ratings..." | tee -a "${log_file}"
        
        # Add CSV header to output if needed
        if [ ! -s "$output_csv" ]; then
            echo "Title,Year,imdbID,tmdbID,WatchedDate,Rating10,Rewatch" > "$output_csv"
        fi
        
        # Extract data from watched file to temporary file using jq
        jq -r '.[] | 
            # Handle null values safely with defaults
            . as $item | 
            {
                title: (try ($item.movie.title) catch null // "Unknown Title"),
                year: (try ($item.movie.year | tostring) catch null // ""),
                imdb_id: (try ($item.movie.ids.imdb) catch null // ""),
                tmdb_id: (try ($item.movie.ids.tmdb | tostring) catch null // ""),
                watched_at: (try ($item.last_watched_at) catch null // ""),
                trakt_id: (try ($item.movie.ids.trakt | tostring) catch null // ""),
                plays: (try ($item.plays | tostring) catch null // "1")
            } | 
            # Build CSV line with safe values
            [.title, .year, .imdb_id, .tmdb_id, .watched_at, .trakt_id, .plays] | 
            @csv' "$watched_file" > "$tmp_file"
        
        # Process the file line by line
        while IFS=, read -r title year imdb_id tmdb_id watched_at trakt_id plays; do
            # Clean quotes from CSV fields if present
            title=$(echo "$title" | sed -e 's/^"//' -e 's/"$//')
            year=$(echo "$year" | sed -e 's/^"//' -e 's/"$//')
            imdb_id=$(echo "$imdb_id" | sed -e 's/^"//' -e 's/"$//')
            tmdb_id=$(echo "$tmdb_id" | sed -e 's/^"//' -e 's/"$//')
            watched_at=$(echo "$watched_at" | sed -e 's/^"//' -e 's/"$//')
            trakt_id=$(echo "$trakt_id" | sed -e 's/^"//' -e 's/"$//')
            plays=$(echo "$plays" | sed -e 's/^"//' -e 's/"$//')
            
            # Skip entries with missing key data
            if [ -z "$title" ] || [ "$title" = "null" ]; then
                continue
            fi
            
            # Format watched date (keep only YYYY-MM-DD)
            watched_at=$(echo "$watched_at" | awk -F'T' '{print $1}')
            
            # Determine rewatch status
            rewatch="false"
            if [ -n "$plays" ] && [ "$plays" != "null" ] && [ "$plays" -gt 1 ]; then
                rewatch="true"
            fi
            
            # Get rating if available
            rating=""
            if [ "$has_ratings" = "true" ] && [ -n "$trakt_id" ] && [ "$trakt_id" != "null" ]; then
                rating=$(jq -r --arg id "$trakt_id" '.[$id] // ""' "$ratings_lookup" 2>/dev/null)
                if [ -n "$rating" ] && [ "$rating" != "null" ]; then
                    echo "DEBUG: Found rating for $title ($trakt_id): $rating" >> "${log_file}"
                fi
            fi
            
            # Add to CSV, ensuring proper format for IMDb ID
            if [[ ! "$imdb_id" =~ ^tt ]]; then
                imdb_id="tt$imdb_id"
            fi
            
            echo "\"$title\",\"$year\",\"$imdb_id\",\"$tmdb_id\",\"$watched_at\",\"$rating\",\"$rewatch\"" >> "$output_csv"
        done < "$tmp_file"
        
        # Report results
        local processed_count=$(grep -c "," "$output_csv")
        echo "Movies watched: $processed_count movies processed (fallback mode)" | tee -a "${log_file}"
    else
        echo "DEBUG: Processing watched_movies file with ratings (complete mode)..." | tee -a "${log_file}"
        
        # Create tracking file for existing IDs if it doesn't exist
        if [ ! -f "$existing_ids_file" ]; then
            echo "DEBUG: Creating new tracking file for existing IDs" | tee -a "${log_file}"
            touch "$existing_ids_file"
            # Extract all movie IDs from existing CSV to avoid duplicates - skip header
            if [ -s "$output_csv" ]; then
                sed 1d "$output_csv" | awk -F, '{print $3}' | sed 's/"//g' > "$existing_ids_file"
            fi
        fi
        
        # Extract data from watched file to temporary file
        jq -r '.[] | 
            # Handle null values safely with defaults
            . as $item | 
            {
                title: (try ($item.movie.title) catch null // "Unknown Title"),
                year: (try ($item.movie.year | tostring) catch null // ""),
                imdb_id: (try ($item.movie.ids.imdb) catch null // ""),
                tmdb_id: (try ($item.movie.ids.tmdb | tostring) catch null // ""),
                watched_at: (try ($item.last_watched_at) catch null // ""),
                trakt_id: (try ($item.movie.ids.trakt | tostring) catch null // ""),
                plays: (try ($item.plays | tostring) catch null // "1")
            } | 
            # Build CSV line with safe values
            [.title, .year, .imdb_id, .tmdb_id, .watched_at, .trakt_id, .plays] | 
            @csv' "$watched_file" > "$tmp_file"
        
        # Process the file line by line
        local added_count=0
        local duplicate_count=0
        
        while IFS=, read -r title year imdb_id tmdb_id watched_at trakt_id plays; do
            # Clean quotes from CSV fields if present
            title=$(echo "$title" | sed -e 's/^"//' -e 's/"$//')
            year=$(echo "$year" | sed -e 's/^"//' -e 's/"$//')
            imdb_id=$(echo "$imdb_id" | sed -e 's/^"//' -e 's/"$//')
            tmdb_id=$(echo "$tmdb_id" | sed -e 's/^"//' -e 's/"$//')
            watched_at=$(echo "$watched_at" | sed -e 's/^"//' -e 's/"$//')
            trakt_id=$(echo "$trakt_id" | sed -e 's/^"//' -e 's/"$//')
            plays=$(echo "$plays" | sed -e 's/^"//' -e 's/"$//')
            
            # Skip entries with missing key data
            if [ -z "$title" ] || [ "$title" = "null" ]; then
                continue
            fi
            
            # Format watched date (keep only YYYY-MM-DD)
            watched_at=$(echo "$watched_at" | awk -F'T' '{print $1}')
            
            # Ensure proper IMDb ID format
            if [[ ! "$imdb_id" =~ ^tt ]]; then
                imdb_id="tt$imdb_id"
            fi
            
            # Check if this movie is already in our list
            if grep -q "$imdb_id" "$existing_ids_file"; then
                ((duplicate_count++))
                continue
            fi
            
            # Determine rewatch status
            rewatch="false"
            if [ -n "$plays" ] && [ "$plays" != "null" ] && [ "$plays" -gt 1 ]; then
                rewatch="true"
            fi
            
            # Get rating if available
            rating=""
            if [ "$has_ratings" = "true" ] && [ -n "$trakt_id" ] && [ "$trakt_id" != "null" ]; then
                rating=$(jq -r --arg id "$trakt_id" '.[$id] // ""' "$ratings_lookup" 2>/dev/null)
                if [ -n "$rating" ] && [ "$rating" != "null" ]; then
                    echo "DEBUG: Found rating for $title ($trakt_id): $rating" >> "${log_file}"
                fi
            fi
            
            # Add to CSV
            echo "\"$title\",\"$year\",\"$imdb_id\",\"$tmdb_id\",\"$watched_at\",\"$rating\",\"$rewatch\"" >> "$output_csv"
            
            # Add to tracking file
            echo "$imdb_id" >> "$existing_ids_file"
            
            ((added_count++))
        done < "$tmp_file"
        
        # Report results
        local total_count=$(grep -c "," "$output_csv")
        echo "📊 Added $added_count new movies from watched list" | tee -a "${log_file}"
        echo "📊 Skipped $duplicate_count duplicate movies" | tee -a "${log_file}"
        echo "Total movies after combining history and watched list: $total_count movies processed" | tee -a "${log_file}"
    fi
    
    # Clean up
    rm -f "$tmp_file"
    
    return 0
}

# Create backup archive
create_backup_archive() {
    local backup_dir="$1"
    local log_file="$2"
    
    debug_msg "Creating backup archive" "$log_file"
    # Generate a unique backup archive name
    backup_archive_name="backup-$(date '+%Y%m%d%H%M%S').tar.gz"
    # Create the archive
    tar -czvf "${backup_dir}/${backup_archive_name}" -C "$(dirname "${backup_dir}")" "$(basename "${backup_dir}")" > /dev/null 2>&1
    echo -e "Backup completed: ${backup_dir}/${backup_archive_name}" | tee -a "${log_file}"
}

# Improved function to deduplicate movies based on IMDb ID
deduplicate_movies() {
    local input_csv="$1"
    local output_csv="$2"
    local log_file="$3"
    
    echo "🔄 Deduplicating movies in CSV file..." | tee -a "${log_file}"
    
    # Create a temporary file for the deduplicated content
    local temp_csv="${input_csv}.dedup"
    
    # Keep header line
    head -n 1 "$input_csv" > "$temp_csv"
    
    # Track IMDb IDs we've seen
    local seen_ids=()
    local id_file=$(mktemp)
    
    # Count total movies
    local total_lines=$(wc -l < "$input_csv")
    local total_movies=$((total_lines - 1))
    echo "📊 Total movies before deduplication: $total_movies" | tee -a "${log_file}"
    
    # Process each line (skipping header)
    cat "$input_csv" | tail -n +2 | while IFS=, read -r title year imdb tmdb date rating rewatch; do
        # Extract IMDb ID (remove quotes if present)
        clean_imdb=$(echo "$imdb" | sed 's/"//g' | sed 's/^tt//g')
        
        # Skip if no IMDb ID
        if [ -z "$clean_imdb" ]; then
            echo "⚠️ Skipping movie with no IMDb ID: $title ($year)" | tee -a "${log_file}"
            continue
        fi
        
        # Use a file to track seen IDs (more reliable than array in a subshell)
        if ! grep -q "^$clean_imdb$" "$id_file"; then
            echo "$clean_imdb" >> "$id_file"
            # Ensure IMDb ID has tt prefix
            formatted_imdb="\"tt${clean_imdb}\""
            echo "${title},${year},${formatted_imdb},${tmdb},${date},${rating},${rewatch}" >> "$temp_csv"
        fi
    done
    
    # Move the deduplicated file to the output
    mv "$temp_csv" "$output_csv"
    
    # Count deduplicated movies
    local dedup_lines=$(wc -l < "$output_csv")
    local dedup_movies=$((dedup_lines - 1))
    echo "📊 Total movies after deduplication: $dedup_movies" | tee -a "${log_file}"
    echo "🔄 Removed $((total_movies - dedup_movies)) duplicate entries" | tee -a "${log_file}"
    
    # Clean up
    rm -f "$id_file"
    
    return 0
}

# Function to limit the number of movies in the CSV if LIMIT_FILMS is set
limit_movies_in_csv() {
    local input_csv="$1"
    local output_csv="$2"
    local log_file="$3"
    
    # Default to no limit if LIMIT_FILMS is not set
    local limit=${LIMIT_FILMS:-0}
    
    # Check if input file exists and is not empty
    if [ ! -f "$input_csv" ]; then
        echo "❌ ERROR: Input CSV file does not exist: $input_csv" | tee -a "${log_file}"
        return 1
    fi
    
    # Ensure input file has a header
    local header="Title,Year,imdbID,tmdbID,WatchedDate,Rating10,Rewatch"
    local first_line=$(head -n 1 "$input_csv")
    
    # Create a temporary file for processing
    local temp_input_csv="${input_csv}.with_header"
    
    # If the first line is not the expected header, add it
    if [ "$first_line" != "$header" ]; then
        echo "⚠️ CSV file is missing header, adding it..." | tee -a "${log_file}"
        echo "$header" > "$temp_input_csv"
        cat "$input_csv" >> "$temp_input_csv"
    else
        cp "$input_csv" "$temp_input_csv"
    fi
    
    # Check if limit is a positive number
    if [[ "$limit" =~ ^[0-9]+$ ]] && [ "$limit" -gt 0 ]; then
        echo "🎯 Limiting CSV to the most recent $limit movies..." | tee -a "${log_file}"
        
        # Create temporary files for processing
        local temp_csv="${input_csv}.limited"
        local prep_csv="${input_csv}.prep"
        local sorted_csv="${input_csv}.sorted"
        local final_csv="${input_csv}.final"
        
        # Keep header line in output
        echo "$header" > "$temp_csv"
        
        # Show sample of dates in the file for debugging
        echo "🔍 Sample of watch dates in CSV:" | tee -a "${log_file}"
        tail -n +2 "$temp_input_csv" | cut -d, -f5 | tr -d '"' | sort | uniq | head -n 5 | tee -a "${log_file}"
        
        # Count lines from input (minus header if present)
        local input_count=$(tail -n +2 "$temp_input_csv" | wc -l)
        echo "📊 Total movies before limiting: $input_count" | tee -a "${log_file}"
        
        echo "# Preprocessing dates for sorting..." | tee -a "${log_file}"
        
        # Create a preprocessed CSV with validated dates and a sort key in the first column
        # Skip the header as we'll add it back later
        tail -n +2 "$temp_input_csv" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
            # Clean the date (remove quotes)
            clean_date=$(echo "$date" | tr -d '"')
            
            # Check if the date is in YYYY-MM-DD format
            if [[ "$clean_date" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
                # Valid ISO date - use it as sort key
                echo "$clean_date|${title}|${year}|${imdb}|${tmdb}|${date}|${rating}|${rewatch}" >> "$prep_csv"
            else
                # Invalid or empty date - use old date to sort at the end
                echo "1970-01-01|${title}|${year}|${imdb}|${tmdb}|${date}|${rating}|${rewatch}" >> "$prep_csv"
                echo "⚠️ Invalid date format found: '$date' for movie: $title. Using default." | tee -a "${log_file}"
            fi
        done
        
        # Sort by date in reverse order and take only the top N entries
        echo "# Sorting movies by date (newest first)..." | tee -a "${log_file}"
        if [ -f "$prep_csv" ]; then
            # Sort lines by the date prefix (newest first) and keep only the top N
            sort -r "$prep_csv" | head -n "$limit" > "$sorted_csv"
            
            # Now convert back to CSV format by removing the sort key
            cat "$sorted_csv" | while IFS='|' read -r sort_key title year imdb tmdb date rating rewatch; do
                # Extract TMDB ID for debugging
                tmdb_clean=""
                if [[ "$tmdb" =~ ([0-9]+) ]]; then
                    tmdb_clean=${BASH_REMATCH[1]}
                    echo "🔍 Movie: $title, TMDB: $tmdb_clean, Rating: $rating" >> "${log_file}"
                fi
                
                # Reconstitute the CSV line
                echo "${title},${year},${imdb},${tmdb},${date},${rating},${rewatch}" >> "$temp_csv"
            done
        else
            echo "⚠️ No valid data to process after preprocessing" | tee -a "${log_file}"
        fi
        
        # Move the limited file to a temporary final file
        cp "$temp_csv" "$final_csv"
        
        # Find all possible ratings files
        echo "# Looking for ratings files in backup directories..." | tee -a "${log_file}"
        local backup_dir=$(dirname "$(dirname "$log_file")")/backup
        echo "📊 Searching in backup directory: $backup_dir" | tee -a "${log_file}"
        
        # Find the most recent ratings file
        local ratings_file=$(find "$backup_dir" -name '*ratings_movies.json' -type f -print0 | 
                     xargs -0 ls -t | head -n1)
        
        if [ -n "$ratings_file" ] && [ -f "$ratings_file" ]; then
            echo "📊 Found ratings file: $ratings_file" | tee -a "${log_file}"
            local ratings_count=$(jq '. | length' "$ratings_file" 2>/dev/null || echo "unknown")
            echo "📊 Ratings file contains $ratings_count ratings" | tee -a "${log_file}"
        else
            echo "⚠️ No ratings file found in backup directories" | tee -a "${log_file}"
        fi
        
        # Apply comprehensive ratings lookup if ratings file exists
        if [ -f "$ratings_file" ] && [ -s "$ratings_file" ]; then
            echo "# Finding ratings for all selected films..." | tee -a "${log_file}"
            
            # Create corrected file with header
            echo "$header" > "$output_csv"
            
            # Create temporary lookup files for faster access
            local title_ratings_file=$(mktemp)
            local imdb_ratings_file=$(mktemp)
            local tmdb_ratings_file=$(mktemp)
            
            echo "# Creating ratings lookup maps..." | tee -a "${log_file}"
            # Create lookup by title
            jq -r '.[] | [.movie.title, .rating] | @tsv' "$ratings_file" > "$title_ratings_file"
            
            # Create lookup by IMDb ID
            jq -r '.[] | select(.movie.ids.imdb != null) | [.movie.ids.imdb, .rating] | @tsv' "$ratings_file" > "$imdb_ratings_file"
            
            # Create lookup by TMDB ID
            jq -r '.[] | select(.movie.ids.tmdb != null) | [.movie.ids.tmdb | tostring, .rating] | @tsv' "$ratings_file" > "$tmdb_ratings_file"
            
            # Show sample of lookups
            echo "📊 Sample of title ratings lookup (first 3 entries):" | tee -a "${log_file}"
            head -n 3 "$title_ratings_file" | tee -a "${log_file}"
            
            # Variables to track rating matches
            local ratings_found=0
            local ratings_total=0
            
            # Process the limited CSV and find ratings
            tail -n +2 "$final_csv" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
                ((ratings_total++))
                
                # Clean fields (remove quotes)
                clean_title=$(echo "$title" | tr -d '"')
                clean_imdb=$(echo "$imdb" | tr -d '"')
                clean_tmdb=$(echo "$tmdb" | tr -d '"')
                clean_rating=$(echo "$rating" | tr -d '"')
                
                # Default to existing rating
                local final_rating="$clean_rating"
                
                # Only look for a new rating if current one is empty
                if [ -z "$final_rating" ] || [ "$final_rating" = "null" ]; then
                    # Try exact title match
                    local found_rating=$(grep -F "$clean_title" "$title_ratings_file" | head -n1 | cut -f2)
                    
                    if [ -n "$found_rating" ]; then
                        final_rating="$found_rating"
                        ((ratings_found++))
                        echo "✅ Found rating $final_rating for '$clean_title' by title match" | tee -a "${log_file}"
                    else
                        # Try by IMDb ID
                        if [ -n "$clean_imdb" ] && [ "$clean_imdb" != "null" ]; then
                            imdb_rating=$(grep -F "$clean_imdb" "$imdb_ratings_file" | head -n1 | cut -f2)
                            if [ -n "$imdb_rating" ]; then
                                final_rating="$imdb_rating"
                                ((ratings_found++))
                                echo "✅ Found rating $final_rating for '$clean_title' by IMDb ID ($clean_imdb)" | tee -a "${log_file}"
                            fi
                        fi
                        
                        # Try by TMDB ID if still no rating
                        if { [ -z "$final_rating" ] || [ "$final_rating" = "null" ]; } && [ -n "$clean_tmdb" ] && [ "$clean_tmdb" != "null" ]; then
                            tmdb_rating=$(grep -F "$clean_tmdb" "$tmdb_ratings_file" | head -n1 | cut -f2)
                            if [ -n "$tmdb_rating" ]; then
                                final_rating="$tmdb_rating"
                                ((ratings_found++))
                                echo "✅ Found rating $final_rating for '$clean_title' by TMDB ID ($clean_tmdb)" | tee -a "${log_file}"
                            fi
                        fi
                        
                        # If still no rating, try direct lookup from the ratings file
                        if [ -z "$final_rating" ] || [ "$final_rating" = "null" ]; then
                            # Try a case-insensitive grep search
                            local grep_result=$(grep -i "\"title\": \"$clean_title\"" "$ratings_file" -A 10 | grep -m 1 "\"rating\":" | awk -F: '{print $2}' | tr -d ' ,')
                            if [ -n "$grep_result" ]; then
                                final_rating="$grep_result"
                                ((ratings_found++))
                                echo "✅ Found rating $final_rating for '$clean_title' by direct file search" | tee -a "${log_file}"
                            fi
                        fi
                    fi
                else
                    echo "ℹ️ Using existing rating $final_rating for '$clean_title'" | tee -a "${log_file}"
                    ((ratings_found++))
                fi
                
                # Reconstruct the line with the rating
                echo "\"$clean_title\",\"$year\",\"$clean_imdb\",\"$clean_tmdb\",\"$date\",\"$final_rating\",\"$rewatch\"" >> "$output_csv"
            done
            
            # Report ratings stats
            echo "📊 Ratings found: $ratings_found out of $ratings_total movies" | tee -a "${log_file}"
            
            # Clean up
            rm -f "$title_ratings_file" "$imdb_ratings_file" "$tmdb_ratings_file"
        else
            # No ratings file available, just use the original ratings
            echo "⚠️ No valid ratings file found, using existing ratings" | tee -a "${log_file}"
            cp "$final_csv" "$output_csv"
        fi
        
        # Count final movies
        local final_count=$(tail -n +2 "$output_csv" | wc -l)
        echo "📊 Keeping only the $final_count most recent movies" | tee -a "${log_file}"
        
        # Debug: Show the dates that were kept
        echo "📅 Watch dates of kept movies (newest first):" | tee -a "${log_file}"
        echo "---------------------------------------------" | tee -a "${log_file}"
        # Print the first 10 entries with title and date
        tail -n +2 "$output_csv" | head -n 10 | while IFS=, read -r title year imdb tmdb date rating rewatch; do
            echo "Film: $title - Date: $date - Rating: $rating" | tee -a "${log_file}"
        done
        echo "---------------------------------------------" | tee -a "${log_file}"
        
        # Clean up
        rm -f "$temp_input_csv" "$temp_csv" "$prep_csv" "$sorted_csv" "$final_csv"
        
        return 0
    else
        # No limit needed
        echo "📊 No movie limit applied (LIMIT_FILMS=$limit)" | tee -a "${log_file}"
        
        # Just ensure header and copy
        cp "$temp_input_csv" "$output_csv"
        
        # Clean up
        rm -f "$temp_input_csv"
        
        return 0
    fi
}

# Add this function call at the end of process_data function 