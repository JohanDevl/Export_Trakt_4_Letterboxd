#!/bin/bash
#
# Cron wrapper for Export_Trakt_4_Letterboxd.sh
#

# Set script variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXPORT_SCRIPT="${SCRIPT_DIR}/Export_Trakt_4_Letterboxd.sh"
LOGFILE="${SCRIPT_DIR}/logs/cron_export.log"
CONFIG_FILE="${SCRIPT_DIR}/config/.config.cfg"

# Load config if it exists
if [ -f "$CONFIG_FILE" ]; then
  source "$CONFIG_FILE"
fi

# Make sure log directory exists
mkdir -p "$(dirname "$LOGFILE")"

# Define colors for better log readability
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

# Function to log to both console and file
log_both() {
  echo -e "$1" | tee -a "$LOGFILE"
}

# Log start time with export option
export_option="${1:-complete}"
start_time=$(date +"%Y-%m-%d %H:%M:%S")

# Visual separator for better readability
log_both "\n${BOLD}================================================================================${NC}"
log_both "${BLUE}[CRON] Starting Trakt to Letterboxd Export at ${start_time}${NC} 🚀"
log_both "${YELLOW}Exporting your Trakt data with option '${export_option}'...${NC}"
log_both "${BOLD}================================================================================${NC}\n"

# Run the export script and capture output
# We use a temporary file to capture the output
temp_output_file=$(mktemp)
"$EXPORT_SCRIPT" "$export_option" > "$temp_output_file" 2>&1
exit_code=$?

# Append the script output to our log file
cat "$temp_output_file" >> "$LOGFILE"

# Get end time
end_time=$(date +"%Y-%m-%d %H:%M:%S")

# If export was successful
if [ $exit_code -eq 0 ]; then
  # Try to count the number of films exported by checking the CSV file
  csv_file="${SCRIPT_DIR}/copy/letterboxd_import.csv"
  if [ -f "$csv_file" ]; then
    # Count movies (subtract 1 for header)
    movie_count=$(($(wc -l < "$csv_file") - 1))
    file_size=$(du -h "$csv_file" | cut -f1)
    
    log_both "\n${BOLD}================================================================================${NC}"
    log_both "${GREEN}[CRON] Export completed successfully at ${end_time}${NC} ✅"
    log_both "${BLUE}Your Letterboxd import file is ready in the copy directory!${NC}"
    log_both "${GREEN}Exported ${BOLD}${movie_count}${NC}${GREEN} movies to CSV file${NC}"
    log_both "${YELLOW}File: ${SCRIPT_DIR}/copy/letterboxd_import.csv (${file_size} size)${NC}"
    log_both "${BOLD}================================================================================${NC}\n"
  else
    log_both "\n${BOLD}================================================================================${NC}"
    log_both "${GREEN}[CRON] Export completed successfully at ${end_time}${NC} ✅"
    log_both "${YELLOW}Warning: CSV file not found at expected location.${NC}"
    log_both "${BOLD}================================================================================${NC}\n"
  fi
else
  # If export failed
  log_both "\n${BOLD}================================================================================${NC}"
  log_both "${RED}[CRON] Export failed at ${end_time} with exit code ${exit_code}${NC} ❌"
  log_both "${YELLOW}Check the logs for details: ${LOGFILE}${NC}"
  log_both "${BOLD}================================================================================${NC}\n"
fi

# Clean up
rm -f "$temp_output_file"

# Exit with the same exit code as the export script
exit $exit_code 