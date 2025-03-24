#!/bin/bash
#
# Data processing functions for Trakt to Letterboxd export
#

# Create ratings lookup file from ratings_movies.json
create_ratings_lookup() {
    local ratings_file="$1"
    local output_file="$2"
    local log_file="$3"
    
    if [ -f "$ratings_file" ]; then
        echo "DEBUG: Creating ratings lookup file..." | tee -a "${log_file}"
        jq -c 'reduce .[] as $item ({}; .[$item.movie.ids.trakt | tostring] = $item.rating)' "$ratings_file" > "$output_file"
        return 0
    else
        echo "{}" > "$output_file"
        echo "WARNING: Ratings file not found, creating empty lookup" | tee -a "${log_file}"
        return 1
    fi
}

# Create plays count lookup from watched_movies.json
create_plays_count_lookup() {
    local watched_file="$1"
    local output_file="$2"
    local log_file="$3"
    
    if [ -f "$watched_file" ]; then
        echo "DEBUG: Creating plays count lookup from watched_movies..." | tee -a "${log_file}"
        jq -c 'reduce .[] as $item ({}; if $item.movie.ids.imdb != null then .[$item.movie.ids.imdb] = $item.plays else . end)' "$watched_file" > "$output_file"
        return 0
    else
        echo "{}" > "$output_file"
        echo "WARNING: Watched movies file not found, creating empty lookup" | tee -a "${log_file}"
        return 1
    fi
}

# Process history movies with ratings
process_history_movies() {
    local history_file="$1"
    local ratings_lookup="$2"
    local plays_lookup="$3"
    local output_csv="$4"
    local raw_output_file="$5"
    local log_file="$6"
    
    if [ -f "$history_file" ]; then
        echo "DEBUG: Processing history_movies file with ratings..." | tee -a "${log_file}"
        
        # Extract watched movies with their date, rating, and rewatch status
        jq -r --slurpfile ratings "$ratings_lookup" --slurpfile plays "$plays_lookup" '.[] | 
            [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
             (if .watched_at then .watched_at | split("T")[0] else "" end),
             ($ratings[0][.movie.ids.trakt | tostring] // ""),
             (if ($plays[0][.movie.ids.imdb] // 1) > 1 then "true" else "false" end)] | 
            @csv' "$history_file" > "$raw_output_file"
        
        # Process the CSV line by line to properly format
        cat "$raw_output_file" | while IFS=, read -r title year imdb tmdb date rating rewatch; do
            # Remove any existing "tt" prefix from IMDb ID to avoid duplication
            imdb=$(echo "$imdb" | sed 's/^"*tt//g' | sed 's/"*$//g')
            # Properly quote title and format IMDb ID with tt prefix
            echo "${title},${year},\"tt${imdb}\",${tmdb},${date},${rating},${rewatch}" >> "$output_csv"
        done
        
        echo "Movies history: $(wc -l < "$output_csv") movies processed" | tee -a "${log_file}"
        return 0
    else
        echo "WARNING: History movies file not found, skipping" | tee -a "${log_file}"
        return 1
    fi
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
    
    if [ "$is_fallback" = "true" ]; then
        echo "DEBUG: No history found. Processing watched_movies file with ratings..." | tee -a "${log_file}"
        
        # Extract watched movies with their date and rating
        jq -r --slurpfile ratings "$ratings_lookup" '.[] | 
            [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
             (if .last_watched_at then .last_watched_at | split("T")[0] else "" end),
             ($ratings[0][.movie.ids.trakt | tostring] // ""),
             (if .plays > 1 then "true" else "false" end)] | 
            @csv' "$watched_file" > "$raw_output_file"
        
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
        
        # Use watched count from the API for rewatch status
        jq -r --slurpfile ratings "$ratings_lookup" '.[] | 
            [.movie.title, .movie.year, .movie.ids.imdb, .movie.ids.tmdb, 
             (if .last_watched_at then .last_watched_at | split("T")[0] else "" end),
             ($ratings[0][.movie.ids.trakt | tostring] // ""),
             (if .plays > 1 then "true" else "false" end)] | 
            @csv' "$watched_file" > "$raw_output_file"
        
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