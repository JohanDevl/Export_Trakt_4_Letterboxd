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

# Add this function call at the end of process_data function 