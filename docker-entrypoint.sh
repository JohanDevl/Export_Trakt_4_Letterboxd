#!/bin/bash
set -e

# Create config directory if it doesn't exist
mkdir -p /app/config

# Move the example config file to the config directory if it doesn't exist there
if [ ! -f /app/config/.config.cfg.example ]; then
    echo "Moving example config file to config directory..."
    cp /app/.config.cfg.example /app/config/.config.cfg.example
    # Remove the original example file after copying
    rm -f /app/.config.cfg.example
fi

# Check if config file exists
if [ ! -f /app/config/.config.cfg ]; then
    echo "Config file not found. Creating from template..."
    cp /app/config/.config.cfg.example /app/config/.config.cfg
    echo "Please edit /app/config/.config.cfg with your Trakt API credentials."
fi

# Remove any existing symlink or config file in the root directory
if [ -L /app/.config.cfg ] || [ -f /app/.config.cfg ]; then
    rm -f /app/.config.cfg
fi

# Ensure the config file is writable
set +e  # Don't exit on error
chmod -f 644 /app/config/.config.cfg
chmod -f 644 /app/config/.config.cfg.example

# Create necessary directories
mkdir -p /app/logs /app/copy /app/brain_ops /app/backup /app/TEMP
chmod -f -R 755 /app/logs /app/copy /app/brain_ops /app/backup /app/TEMP
set -e  # Resume exit on error

# Make scripts executable
chmod +x /app/Export_Trakt_4_Letterboxd.sh /app/setup_trakt.sh

# Update scripts to use the config file in the config directory
sed -i 's|CONFIG_FILE="${SCRIPT_DIR}/.config.cfg"|CONFIG_FILE="/app/config/.config.cfg"|g' /app/setup_trakt.sh
sed -i 's|source ${SCRIPT_DIR}/.config.cfg|source /app/config/.config.cfg|g' /app/Export_Trakt_4_Letterboxd.sh

# Display help message
echo "=== Export Trakt 4 Letterboxd ==="
echo ""
echo "Available commands:"
echo "  setup_trakt.sh - Configure Trakt API authentication"
echo "  Export_Trakt_4_Letterboxd.sh [option] - Export Trakt data"
echo ""
echo "Options for Export_Trakt_4_Letterboxd.sh:"
echo "  normal (default) - Export rated movies, episodes, history, and watchlist"
echo "  initial - Export only rated and watched movies"
echo "  complet - Export all available data"
echo ""

# Execute command if provided, otherwise start shell
if [ $# -gt 0 ]; then
    exec "$@"
else
    exec /bin/bash
fi 