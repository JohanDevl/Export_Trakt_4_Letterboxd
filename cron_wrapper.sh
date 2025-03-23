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

# Log start time with export option
export_option="${1:-complete}"
start_time=$(date +"%Y-%m-%d %H:%M:%S")
echo -e "\n${BOLD}================================================================================${NC}" >> "$LOGFILE"
echo -e "${BLUE}[CRON] Starting Trakt to Letterboxd Export at ${start_time}${NC} 🚀" >> "$LOGFILE"
echo -e "${YELLOW}Exporting your Trakt data with option '${export_option}'...${NC}" >> "$LOGFILE"
echo -e "${BOLD}================================================================================${NC}\n" >> "$LOGFILE"

# Run the export script
"$EXPORT_SCRIPT" "$export_option" >> "$LOGFILE" 2>&1
exit_code=$?

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
    
    echo -e "\n${BOLD}================================================================================${NC}" >> "$LOGFILE"
    echo -e "${GREEN}[CRON] Export completed successfully at ${end_time}${NC} ✅" >> "$LOGFILE"
    echo -e "${BLUE}Your Letterboxd import file is ready in the copy directory!${NC}" >> "$LOGFILE"
    echo -e "${GREEN}Exported ${BOLD}${movie_count}${NC}${GREEN} movies to CSV file${NC}" >> "$LOGFILE"
    echo -e "${YELLOW}File: ${SCRIPT_DIR}/copy/letterboxd_import.csv (${file_size} size)${NC}" >> "$LOGFILE"
    echo -e "${BOLD}================================================================================${NC}\n" >> "$LOGFILE"
  else
    echo -e "\n${BOLD}================================================================================${NC}" >> "$LOGFILE"
    echo -e "${GREEN}[CRON] Export completed successfully at ${end_time}${NC} ✅" >> "$LOGFILE"
    echo -e "${YELLOW}Warning: CSV file not found at expected location.${NC}" >> "$LOGFILE"
    echo -e "${BOLD}================================================================================${NC}\n" >> "$LOGFILE"
  fi
else
  # If export failed
  echo -e "\n${BOLD}================================================================================${NC}" >> "$LOGFILE"
  echo -e "${RED}[CRON] Export failed at ${end_time} with exit code ${exit_code}${NC} ❌" >> "$LOGFILE"
  echo -e "${YELLOW}Check the logs for details: ${LOGFILE}${NC}" >> "$LOGFILE"
  echo -e "${BOLD}================================================================================${NC}\n" >> "$LOGFILE"
fi

# Exit with the same exit code as the export script
exit $exit_code 