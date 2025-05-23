#!/bin/bash

# Debug script for monitoring Docker container logs in real-time
# Export Trakt 4 Letterboxd

echo "=== Docker Logs Debug Script ==="
echo ""

# Function to display usage
usage() {
    echo "Usage: $0 [service-profile] [options]"
    echo ""
    echo "Service profiles:"
    echo "  schedule-15min  - Every 15 minutes (testing)"
    echo "  schedule-6h     - Every 6 hours (production)"
    echo "  schedule-daily  - Daily at 2:30 AM"
    echo "  schedule-weekly - Weekly on Sundays"
    echo ""
    echo "Options:"
    echo "  --follow, -f    - Follow logs in real-time (default)"
    echo "  --tail N        - Show last N lines (default: 50)"
    echo "  --since TIME    - Show logs since TIME (e.g., '10m', '1h', '2023-01-01T10:00:00')"
    echo "  --timestamps    - Show timestamps"
    echo "  --help, -h      - Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 schedule-15min"
    echo "  $0 schedule-6h --tail 100"
    echo "  $0 schedule-daily --since 1h"
    echo ""
}

# Default values
PROFILE="schedule-15min"
FOLLOW=true
TAIL=50
SINCE=""
TIMESTAMPS=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        schedule-15min|schedule-6h|schedule-daily|schedule-weekly)
            PROFILE="$1"
            shift
            ;;
        --follow|-f)
            FOLLOW=true
            shift
            ;;
        --tail)
            TAIL="$2"
            shift 2
            ;;
        --since)
            SINCE="$2"
            shift 2
            ;;
        --timestamps)
            TIMESTAMPS=true
            shift
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

echo "Monitoring logs for profile: $PROFILE"
echo ""

# Check if service is running
if ! docker compose --profile "$PROFILE" ps | grep -q "Up"; then
    echo "⚠️  Service with profile '$PROFILE' is not running."
    echo ""
    echo "To start the service:"
    echo "  docker compose --profile $PROFILE up -d"
    echo ""
    echo "Current running services:"
    docker compose ps --filter "status=running"
    exit 1
fi

# Build docker logs command
LOG_CMD="docker compose --profile $PROFILE logs"

if [ "$FOLLOW" = true ]; then
    LOG_CMD="$LOG_CMD --follow"
fi

if [ -n "$TAIL" ]; then
    LOG_CMD="$LOG_CMD --tail=$TAIL"
fi

if [ -n "$SINCE" ]; then
    LOG_CMD="$LOG_CMD --since=$SINCE"
fi

if [ "$TIMESTAMPS" = true ]; then
    LOG_CMD="$LOG_CMD --timestamps"
fi

echo "Command: $LOG_CMD"
echo ""
echo "🔍 Monitoring logs... (Press Ctrl+C to stop)"
echo "============================================"
echo ""

# Execute the command
eval $LOG_CMD 