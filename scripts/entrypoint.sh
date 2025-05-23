#!/bin/sh
set -e

# Afficher la version
echo "Export Trakt 4 Letterboxd - Scheduler"
echo "======================================"

# Si la variable EXPORT_SCHEDULE est définie, lancer le scheduler
if [ -n "$EXPORT_SCHEDULE" ]; then
    echo "Schedule configured: $EXPORT_SCHEDULE"
    echo "Export mode: ${EXPORT_MODE:-complete}"
    echo "Export type: ${EXPORT_TYPE:-all}"
    
    # Lancer le programme avec la commande schedule
    exec /app/export-trakt schedule
else
    echo "No EXPORT_SCHEDULE defined. Exiting."
    exit 1
fi 