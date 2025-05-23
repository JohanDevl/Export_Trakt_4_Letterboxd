#!/bin/bash
#
# Installation script for Export Trakt 4 Letterboxd
#

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

# Get script directory (resolving symlinks)
SCRIPT_DIR="$( cd "$( dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")" )" && pwd )"

echo -e "${BLUE}${BOLD}==================================================${NC}"
echo -e "${GREEN}${BOLD}  Export Trakt 4 Letterboxd - Installation Script  ${NC}"
echo -e "${BLUE}${BOLD}==================================================${NC}"
echo

# Create essential directories
echo -e "${YELLOW}Creating required directories...${NC}"
directories=("lib" "config" "logs" "backup" "TEMP" "copy")

for dir in "${directories[@]}"; do
    if [ ! -d "${SCRIPT_DIR}/${dir}" ]; then
        mkdir -p "${SCRIPT_DIR}/${dir}"
        echo -e "  - Created directory: ${CYAN}${dir}${NC}"
    else
        echo -e "  - Directory already exists: ${CYAN}${dir}${NC}"
    fi
done
echo -e "${GREEN}✓ Directories setup complete${NC}"
echo

# Check for required dependencies
echo -e "${YELLOW}Checking required dependencies...${NC}"
dependencies=("curl" "jq" "sed" "awk")
missing_deps=()

for cmd in "${dependencies[@]}"; do
    if ! command -v "$cmd" &> /dev/null; then
        echo -e "  - ${RED}✗ $cmd not found${NC}"
        missing_deps+=("$cmd")
    else
        echo -e "  - ${GREEN}✓ $cmd found: $(command -v "$cmd")${NC}"
    fi
done

if [ ${#missing_deps[@]} -gt 0 ]; then
    echo -e "\n${RED}${BOLD}Missing dependencies:${NC} ${missing_deps[*]}"
    echo -e "${YELLOW}Please install the missing dependencies before continuing.${NC}"
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "\n${CYAN}On macOS, you can install them with:${NC}"
        echo "  brew install ${missing_deps[*]}"
    elif [[ -f /etc/debian_version ]]; then
        echo -e "\n${CYAN}On Debian/Ubuntu, you can install them with:${NC}"
        echo "  sudo apt update && sudo apt install ${missing_deps[*]}"
    elif [[ -f /etc/fedora-release ]]; then
        echo -e "\n${CYAN}On Fedora, you can install them with:${NC}"
        echo "  sudo dnf install ${missing_deps[*]}"
    fi
    
    exit 1
fi
echo -e "${GREEN}✓ All dependencies are installed${NC}"
echo

# Check if config file exists, create from example if it doesn't
echo -e "${YELLOW}Setting up configuration...${NC}"
if [ ! -f "${SCRIPT_DIR}/config/.config.cfg" ]; then
    if [ -f "${SCRIPT_DIR}/config/.config.cfg.example" ]; then
        cp "${SCRIPT_DIR}/config/.config.cfg.example" "${SCRIPT_DIR}/config/.config.cfg"
        echo -e "  - ${CYAN}Created config file from example${NC}"
        echo -e "  - ${YELLOW}${BOLD}IMPORTANT: Edit ${SCRIPT_DIR}/config/.config.cfg with your Trakt.tv credentials${NC}"
    else
        echo -e "  - ${RED}No config example found. Creating minimal config...${NC}"
        cat > "${SCRIPT_DIR}/config/.config.cfg" << EOF
# Trakt API Configuration
API_URL="https://api.trakt.tv"
API_KEY=""
API_SECRET=""
REDIRECT_URI="urn:ietf:wg:oauth:2.0:oob"
ACCESS_TOKEN=""
REFRESH_TOKEN=""
USERNAME=""

# Script paths
DOSLOG="${SCRIPT_DIR}/logs"
DOSCOPY="${SCRIPT_DIR}/copy"
BACKUP_DIR="${SCRIPT_DIR}/backup"
TEMP_DIR="${SCRIPT_DIR}/TEMP"
EOF
        echo -e "  - ${YELLOW}${BOLD}IMPORTANT: Edit ${SCRIPT_DIR}/config/.config.cfg with your Trakt.tv credentials${NC}"
    fi
else
    echo -e "  - ${CYAN}Config file already exists${NC}"
fi
echo -e "${GREEN}✓ Configuration setup complete${NC}"
echo

# Set correct file permissions
echo -e "${YELLOW}Setting file permissions...${NC}"
chmod 755 "${SCRIPT_DIR}/Export_Trakt_4_Letterboxd.sh"
chmod 755 "${SCRIPT_DIR}/lib/"*.sh 2>/dev/null || echo "  - No library files to set permissions yet"
chmod 644 "${SCRIPT_DIR}/config/.config.cfg" 2>/dev/null || echo "  - No config file to set permissions yet"
echo -e "${GREEN}✓ File permissions set${NC}"
echo

# Final setup message
echo -e "${BLUE}${BOLD}==================================================${NC}"
echo -e "${GREEN}${BOLD}Installation Complete!${NC}"
echo -e "${BLUE}${BOLD}==================================================${NC}"
echo
echo -e "${CYAN}Next steps:${NC}"
echo -e "1. Edit ${YELLOW}${SCRIPT_DIR}/config/.config.cfg${NC} with your Trakt API credentials"
echo -e "2. Run ${YELLOW}./setup_trakt.sh${NC} to authenticate with Trakt"
echo -e "3. Run ${YELLOW}./Export_Trakt_4_Letterboxd.sh${NC} to export your data"
echo
echo -e "${CYAN}Available options:${NC}"
echo -e "  - ${YELLOW}normal${NC}: Export movie history (default)"
echo -e "  - ${YELLOW}initial${NC}: Export only essential data for first-time users"
echo -e "  - ${YELLOW}complete${NC}: Export all data (history, ratings, watchlist, etc.)"
echo
echo -e "${BLUE}For more information, see the README.md file${NC}"
echo

exit 0 