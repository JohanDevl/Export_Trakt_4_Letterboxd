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
        
        # Verify JSON is valid before processing
        if ! jq empty "$ratings_file" 2>/dev/null; then
            echo "⚠️ WARNING: Invalid JSON in ratings file, creating empty lookup" | tee -a "${log_file}"
            echo "{}" > "$output_file"
            return 1
        fi
        
        # Ensure output directory exists
        local output_dir=$(dirname "$output_file")
        mkdir -p "$output_dir" 2>/dev/null
        
        # Create the lookup with proper error handling
        if ! jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt | tostring] = $item.rating)' "$ratings_file" > "$output_file" 2>/dev/null; then
            echo "⚠️ WARNING: Failed to process ratings file, creating empty lookup" | tee -a "${log_file}"
            echo "{}" > "$output_file"
            return 1
        fi
        
        # Verify the lookup file was created successfully
        if [ ! -s "$output_file" ]; then
            echo "⚠️ WARNING: Ratings lookup file is empty, creating basic JSON" | tee -a "${log_file}"
            echo "{}" > "$output_file"
            return 1
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
        
        # Create the lookup with proper error handling
        if ! jq -c 'reduce .[] as $item ({}; if $item.movie.ids.trakt != null then .[$item.movie.ids.trakt | tostring] = $item.plays else . end)' "$watched_file" > "$output_file" 2>/dev/null; then
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
    
    # Get number of movies in history
    local movie_count=$(jq length "$history_file")
    if [ "$movie_count" -eq 0 ]; then
        echo -e "⚠️ Movies history: No movies found in history" | tee -a "${log}"
        return 1
    fi
    
    echo "DEBUG: Processing history_movies file with ratings..." >> "${log}"
    
    # Create temporary file for processed data
    local tmp_file=$(mktemp)
    
    # More robust JQ processing to handle null values
    jq -r '.[] | 
        # Handle null values safely with defaults
        . as $item | 
        {
            title: (try ($item.movie.title) catch null // "Unknown Title"),
            year: (try ($item.movie.year | tostring) catch null // ""),
            imdb_id: (try ($item.movie.ids.imdb) catch null // ""),
            tmdb_id: (try ($item.movie.ids.tmdb | tostring) catch null // ""),
            watched_at: (try ($item.watched_at) catch null // ""),
            trakt_id: (try ($item.movie.ids.trakt | tostring) catch null // "")
        } | 
        # Build CSV line with safe values
        [.title, .year, .imdb_id, .tmdb_id, .watched_at, .trakt_id] | 
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
                (.movie.ids.trakt // "")
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
    echo "Title,Year,imdbID,tmdbID,WatchedDate,TraktID" > "$raw_output"
    cat "$tmp_file" >> "$raw_output"
    
    # Process the extracted data to add ratings and deduplicate
    local processed_count=0
    local existing_ids=()
    local movie_title=""
    local movie_year=""
    local movie_imdb=""
    local movie_tmdb=""
    local movie_watched=""
    local movie_trakt=""
    
    while IFS=, read -r title year imdb_id tmdb_id watched_at trakt_id; do
        # Clean quotes from CSV fields if present
        title=$(echo "$title" | sed -e 's/^"//' -e 's/"$//')
        year=$(echo "$year" | sed -e 's/^"//' -e 's/"$//')
        imdb_id=$(echo "$imdb_id" | sed -e 's/^"//' -e 's/"$//')
        tmdb_id=$(echo "$tmdb_id" | sed -e 's/^"//' -e 's/"$//')
        watched_at=$(echo "$watched_at" | sed -e 's/^"//' -e 's/"$//')
        trakt_id=$(echo "$trakt_id" | sed -e 's/^"//' -e 's/"$//')
        
        # Skip entries with missing key data
        if [ -z "$title" ] || [ "$title" = "null" ] || [ -z "$watched_at" ] || [ "$watched_at" = "null" ]; then
            continue
        fi
        
        # Format watched date (keep only YYYY-MM-DD)
        watched_at=$(echo "$watched_at" | awk -F'T' '{print $1}')
        
        # Get rating from ratings lookup
        rating=""
        if [ -n "$trakt_id" ] && [ "$trakt_id" != "null" ]; then
            rating=$(jq -r --arg id "$trakt_id" '.[$id] // empty' "$ratings_lookup" 2>/dev/null)
        fi
        
        # Scale rating to 1-10 if it exists (Trakt uses 1-10, but we need to ensure it's in that range)
        if [ -n "$rating" ] && [ "$rating" != "null" ]; then
            # Ensure rating is an integer between 1 and 10
            if ! [[ "$rating" =~ ^[0-9]+$ ]] || [ "$rating" -lt 1 ] || [ "$rating" -gt 10 ]; then
                # Try to convert if it's a decimal
                rating=$(echo "$rating" | awk '{printf "%.0f", $1}')
                # Check if now valid
                if ! [[ "$rating" =~ ^[0-9]+$ ]] || [ "$rating" -lt 1 ] || [ "$rating" -gt 10 ]; then
                    rating=""
                fi
            fi
        else
            rating=""
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
        
        # Add to CSV
        echo "$title,$year,$imdb_id,$tmdb_id,$watched_at,$rating,$rewatch" >> "$csv_output"
        
        # Add to processed IDs
        existing_ids+=("$item_id")
        ((processed_count++))
    done < "$tmp_file"
    
    # Clean up temporary file
    rm -f "$tmp_file"
    
    # Report results
    echo "Movies history: $processed_count movies processed" | tee -a "${log}"
    
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
    
    # Check if ratings_lookup exists to avoid jq errors
    local ratings_param=""
    if [ -f "$ratings_lookup" ]; then
        ratings_param="--slurpfile ratings $ratings_lookup"
    else
        echo "⚠️ WARNING: Ratings lookup file not found, proceeding without ratings" | tee -a "${log_file}"
    fi
    
    if [ "$is_fallback" = "true" ]; then
        echo "DEBUG: No history found. Processing watched_movies file with ratings..." | tee -a "${log_file}"
        
        # Extract watched movies with their date and rating - safely handle missing ratings file
        if [ -n "$ratings_param" ]; then
            jq -r "$ratings_param" '.[] | 
                [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
                 (if .last_watched_at then .last_watched_at | split("T")[0] else "" end),
                 ($ratings[0][.movie.ids.trakt | tostring] // ""),
                 (if .plays > 1 then "true" else "false" end)] | 
                @csv' "$watched_file" > "$raw_output_file"
        else
            # Process without ratings if ratings file is missing
            jq -r '.[] | 
                [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
                 (if .last_watched_at then .last_watched_at | split("T")[0] else "" end),
                 "",
                 (if .plays > 1 then "true" else "false" end)] | 
                @csv' "$watched_file" > "$raw_output_file"
        fi
        
        # Process the CSV line by line to properly format
        cat "$raw_output_file" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
            # Remove any existing "tt" prefix from IMDb ID to avoid duplication
            imdb=$(echo "$imdb" | sed 's/^"*tt//g' | sed 's/"*$//g')
            # Properly quote title and format IMDb ID with tt prefix
            echo "${title},${year},\"tt${imdb}\",${tmdb},${date},${rating},${rewatch}" >> "$output_csv"
        done
        
        echo "Movies watched: $(wc -l < "$output_csv") movies processed" | tee -a "${log_file}"
    else
        echo "DEBUG: Processing watched_movies file with ratings (complete mode)..." | tee -a "${log_file}"
        
        # Extract all movie IDs from existing CSV to avoid duplicates
        awk -F, '{print $3}' "$output_csv" | sed 's/"//g' > "$existing_ids_file"
        
        # Use watched count from the API for rewatch status - safely handle missing ratings file
        if [ -n "$ratings_param" ]; then
            jq -r "$ratings_param" '.[] | 
                [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
                 (if .last_watched_at then .last_watched_at | split("T")[0] else "" end),
                 ($ratings[0][.movie.ids.trakt | tostring] // ""),
                 (if .plays > 1 then "true" else "false" end)] | 
                @csv' "$watched_file" > "$raw_output_file"
        else
            # Process without ratings if ratings file is missing
            jq -r '.[] | 
                [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
                 (if .last_watched_at then .last_watched_at | split("T")[0] else "" end),
                 "",
                 (if .plays > 1 then "true" else "false" end)] | 
                @csv' "$watched_file" > "$raw_output_file"
        fi
        
        # Process the CSV line by line, only adding movies not already in the history
        cat "$raw_output_file" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
            # Format IMDb ID consistently
            clean_imdb=$(echo "$imdb" | sed 's/^"*tt//g' | sed 's/"*$//g')
            formatted_imdb="tt${clean_imdb}"
            
            # Check if this movie is already in our list
            if ! grep -q "$formatted_imdb" "$existing_ids_file"; then
                # Add this movie to our output
                echo "${title},${year},\"${formatted_imdb}\",${tmdb},${date},${rating},${rewatch}" >> "$output_csv"
                # Add to our tracking of existing IDs
                echo "$formatted_imdb" >> "$existing_ids_file"
            fi
        done
        
        echo "Total movies after combining history and watched list: $(wc -l < "$output_csv") movies processed" | tee -a "${log_file}"
    fi
    
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
        
        # Keep header line in output
        echo "$header" > "$temp_csv"
        
        # Show sample of dates in the file for debugging
        echo "🔍 Sample of watch dates in CSV:" | tee -a "${log_file}"
        tail -n +2 "$temp_input_csv" | cut -d, -f5 | tr -d '"' | sort | uniq | head -n 5 | tee -a "${log_file}"
        
        # Count lines from input (minus header if present)
        local input_count=$(tail -n +2 "$temp_input_csv" | wc -l)
        echo "📊 Total movies before limiting: $input_count" | tee -a "${log_file}"
        
        echo "# Preprocessing dates for sorting..." | tee -a "${log_file}"
        
        # Create a preprocessed CSV with validated dates
        # Skip the header as we'll add it back later
        tail -n +2 "$temp_input_csv" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
            # Clean the date (remove quotes)
            clean_date=$(echo "$date" | tr -d '"')
            
            # Check if the date is in YYYY-MM-DD format
            if [[ "$clean_date" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
                # Valid ISO date - use it as is
                echo "${title},${year},${imdb},${tmdb},\"${clean_date}\",${rating},${rewatch}" >> "$prep_csv"
            else
                # Invalid or empty date - use old date to sort at the end
                echo "${title},${year},${imdb},${tmdb},\"1970-01-01\",${rating},${rewatch}" >> "$prep_csv"
                echo "⚠️ Invalid date format found: '$date' for movie: $title. Using default." | tee -a "${log_file}"
            fi
        done
        
        # Sort by date in reverse order and take only the top N entries
        echo "# Sorting movies by date (newest first)..." | tee -a "${log_file}"
        if [ -f "$prep_csv" ]; then
            cat "$prep_csv" | sort -t, -k5,5r | head -n "$limit" >> "$temp_csv"
        else
            echo "⚠️ No valid data to process after preprocessing" | tee -a "${log_file}"
        fi
        
        # Move the limited file to the output
        cp "$temp_csv" "$output_csv"
        
        # Count final movies
        local final_count=$(tail -n +2 "$output_csv" | wc -l)
        echo "📊 Keeping only the $final_count most recent movies" | tee -a "${log_file}"
        
        # Debug: Show the dates that were kept
        echo "📅 Watch dates of kept movies (newest first):" | tee -a "${log_file}"
        echo "---------------------------------------------" | tee -a "${log_file}"
        # Print the first 10 entries with title and date
        tail -n +2 "$output_csv" | head -n 10 | while IFS=, read -r title year imdb tmdb date rating rewatch; do
            echo "Film: $title - Date: $date" | tee -a "${log_file}"
        done
        echo "---------------------------------------------" | tee -a "${log_file}"
        
        # Clean up
        rm -f "$temp_input_csv" "$temp_csv" "$prep_csv"
        
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