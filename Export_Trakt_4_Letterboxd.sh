#!/bin/bash
#
# Export_Trakt_4_Letterboxd - Main Script
# This script exports your Trakt.tv watch history to a CSV format compatible with Letterboxd import.
# Author: Johan
#

# Get script directory (resolving symlinks)
SCRIPT_DIR="$( cd "$( dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")" )" && pwd )"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
SAISPAS='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Initialize logging
CONFIG_DIR="${SCRIPT_DIR}/config"
LOG_TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
LOG="${SCRIPT_DIR}/logs/Export_Trakt_4_Letterboxd_${LOG_TIMESTAMP}.log"

# Create logs directory if it doesn't exist
mkdir -p "${SCRIPT_DIR}/logs"

# Source the main module
if [ -f "${SCRIPT_DIR}/lib/main.sh" ]; then
    source "${SCRIPT_DIR}/lib/main.sh"
else
    echo -e "${RED}ERROR: Main module not found. Did you run the setup script?${NC}" | tee -a "${LOG}"
    exit 1
fi

# Import modules
import_modules "$SCRIPT_DIR"

# Default global variables
TEMP_DIR="${SCRIPT_DIR}/TEMP"
DOSLOG="${SCRIPT_DIR}/logs"
DOSCOPY="${SCRIPT_DIR}/copy"
BACKUP_DIR="${SCRIPT_DIR}/backup"

# Load configuration (needed for language settings)
[ -f "${CONFIG_DIR}/.config.cfg" ] && source "${CONFIG_DIR}/.config.cfg"

# Initialize internationalization - use language from config or auto-detect
init_i18n "$SCRIPT_DIR" "$LOG"

# Log header
echo "===============================================================" | tee -a "${LOG}"
echo -e "${GREEN}$(_ "welcome") - $(_ "starting")${NC}" | tee -a "${LOG}"
echo "===============================================================" | tee -a "${LOG}"
echo -e "${BLUE}$(date) - $(_ "script_execution_start")${NC}" | tee -a "${LOG}"

# Example of using internationalization
echo -e "${CYAN}$(_ "WELCOME") - $(_ "running_in") $(hostname)${NC}" | tee -a "${LOG}"
echo -e "${YELLOW}$(_ "language_set"): ${LANGUAGE:-$(_ "auto_detected")}${NC}" | tee -a "${LOG}"

# Check if we are running in Docker
if [ -f "/.dockerenv" ]; then
    echo -e "${CYAN}$(_ "running_docker")${NC}" | tee -a "${LOG}"
    # Docker-specific settings can be added here
fi

# Parse command line argument (if any)
OPTION="$1"
echo -e "${YELLOW}$(_ "script_option"): ${OPTION:-$(_ "none")}${NC}" | tee -a "${LOG}"

# Run the export process
run_export "$SCRIPT_DIR" "$OPTION"

# Exit with success
exit 0
