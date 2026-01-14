#!/bin/sh
set -e

# Display version
echo "Export Trakt 4 Letterboxd - Scheduler"
echo "======================================"

# If EXPORT_SCHEDULE variable is defined, launch the scheduler
if [ -n "$EXPORT_SCHEDULE" ]; then
    echo "Schedule configured: $EXPORT_SCHEDULE"
    echo "Export mode: ${EXPORT_MODE:-complete}"
    echo "Export type: ${EXPORT_TYPE:-all}"

    # Launch the program with schedule command
    exec /app/export-trakt schedule
else
    echo "No EXPORT_SCHEDULE defined. Exiting."
    exit 1
fi 