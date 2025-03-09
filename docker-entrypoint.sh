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

# Setup cron job if CRON_SCHEDULE is provided
if [ ! -z "${CRON_SCHEDULE}" ]; then
    # Install cron if not already installed
    if ! command -v cron &> /dev/null; then
        echo "Installing cron..."
        apk add --no-cache dcron
    fi

    # Set default export option if not provided
    EXPORT_OPTION=${EXPORT_OPTION:-normal}
    
    echo "Setting up cron job with schedule: ${CRON_SCHEDULE}"
    echo "Export option: ${EXPORT_OPTION}"
    
    # Create a wrapper script for the cron job
    cat > /app/cron_wrapper.sh << 'EOF'
#!/bin/bash
# Get the start time
START_TIME=$(date +"%Y-%m-%d %H:%M:%S")

# Log to container stdout with a friendly message
echo "🎬 [CRON] Starting Trakt to Letterboxd Export at ${START_TIME} 🎬" > /proc/1/fd/1
echo "📊 Exporting your Trakt data... This may take a few minutes." > /proc/1/fd/1

# Redirect all output to the log file
exec > /app/logs/cron_export.log 2>&1

# Print friendly messages
echo "========================================================"
echo "🎬 Starting Trakt to Letterboxd Export - $(date)"
echo "========================================================"
echo "🌟 Exporting your Trakt data to Letterboxd format..."
echo "📊 This may take a few minutes depending on the amount of data."
echo "========================================================"

# Run the export script
cd /app && ./Export_Trakt_4_Letterboxd.sh $1

# Get the end time
END_TIME=$(date +"%Y-%m-%d %H:%M:%S")

# Print completion message
echo "========================================================"
echo "✅ Export completed at $(date)"
echo "🎉 Your Letterboxd import file is ready in the copy directory!"
echo "========================================================"

# Log to container stdout with a friendly completion message
echo "✅ [CRON] Trakt to Letterboxd Export completed at ${END_TIME} ✅" > /proc/1/fd/1
echo "🎉 Your Letterboxd import file is ready in the copy directory! 🎉" > /proc/1/fd/1
EOF
    
    # Make the wrapper script executable
    chmod +x /app/cron_wrapper.sh
    
    # Create cron job using the wrapper script
    echo "${CRON_SCHEDULE} /app/cron_wrapper.sh ${EXPORT_OPTION}" > /etc/crontabs/root
    
    # Make sure the log file exists and is writable
    touch /app/logs/cron_export.log
    chmod 644 /app/logs/cron_export.log
    
    # Start cron daemon with appropriate logging
    echo "Starting cron daemon..."
    crond -b -L 8
    
    echo "Cron job has been set up. Logs will be written to /app/logs/cron_export.log"
    echo "You can also see cron execution messages in the container logs."
fi

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