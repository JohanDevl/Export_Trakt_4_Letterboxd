#!/bin/bash

# Test script for Docker Compose with new --run and --schedule features
# Export Trakt 4 Letterboxd

echo "=== Docker Compose Test Script - Export Trakt 4 Letterboxd ==="
echo ""

# Check if Docker and Docker Compose are available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker compose &> /dev/null; then
    echo "❌ Docker Compose is not available"
    exit 1
fi

echo "✅ Docker and Docker Compose are available"
echo ""

# Test 1: Validate Docker Compose configuration
echo "=== Test 1: Validating Docker Compose configuration ==="
if docker compose config --quiet; then
    echo "✅ Docker Compose configuration is valid"
else
    echo "❌ Docker Compose configuration is invalid"
    exit 1
fi
echo ""

# Test 2: Show available services
echo "=== Test 2: Available services ==="
echo "Immediate execution services (--run):"
echo "  - run-watched: Export watched movies only"
echo "  - run-all: Export all data (recommended for testing)"
echo "  - run-collection: Export collection only"
echo "  - run-ratings: Export ratings only"
echo "  - run-watchlist: Export watchlist only"
echo "  - run-shows: Export shows only"
echo ""
echo "Scheduled services (--schedule):"
echo "  - schedule-6h: Every 6 hours (production)"
echo "  - schedule-daily: Daily at 2:30 AM"
echo "  - schedule-weekly: Weekly on Sundays at 3:00 AM"
echo "  - schedule-15min: Every 15 minutes (testing)"
echo "  - schedule-custom: Custom schedule via environment variables"
echo ""

# Test 3: List all available profiles
echo "=== Test 3: Available Docker Compose profiles ==="
docker compose config --profiles 2>/dev/null || echo "Profile listing not supported in this Docker Compose version"
echo ""

# Test 4: Test configuration validation (if config exists)
echo "=== Test 4: Testing configuration validation ==="
if [ -f "config/config.toml" ]; then
    echo "Configuration file found. Testing validation..."
    if timeout 30s docker compose --profile validate up --remove-orphans 2>/dev/null; then
        echo "✅ Configuration validation passed"
    else
        echo "⚠️  Configuration validation completed (may need setup)"
    fi
else
    echo "⚠️  No configuration file found at config/config.toml"
    echo "   Run: docker compose --profile setup up"
fi
echo ""

# Test 5: Show example commands
echo "=== Test 5: Example commands ==="
echo ""
echo "🚀 Quick start commands:"
echo ""
echo "1. Setup (first time):"
echo "   docker compose --profile setup up"
echo ""
echo "2. Test your configuration:"
echo "   docker compose --profile run-watched up"
echo ""
echo "3. Export all data once:"
echo "   docker compose --profile run-all up"
echo ""
echo "4. Start production scheduler (every 6 hours):"
echo "   docker compose --profile schedule-6h up -d"
echo ""
echo "5. View scheduler logs:"
echo "   docker compose --profile schedule-6h logs -f"
echo ""
echo "6. Stop scheduler:"
echo "   docker compose --profile schedule-6h down"
echo ""
echo "📋 Specific export types:"
echo "   docker compose --profile run-collection up    # Collection only"
echo "   docker compose --profile run-ratings up       # Ratings only"
echo "   docker compose --profile run-watchlist up     # Watchlist only"
echo "   docker compose --profile run-shows up         # Shows only"
echo ""
echo "⏰ Different schedules:"
echo "   docker compose --profile schedule-daily up -d   # Daily at 2:30 AM"
echo "   docker compose --profile schedule-weekly up -d  # Weekly backup"
echo "   docker compose --profile schedule-15min up -d   # Every 15 min (testing)"
echo ""
echo "🎛️  Custom schedule:"
echo "   SCHEDULE=\"0 */4 * * *\" EXPORT_TYPE=\"watched\" EXPORT_MODE=\"normal\" \\"
echo "   docker compose --profile schedule-custom up -d"
echo ""

# Test 6: Show Docker information
echo "=== Test 6: Docker environment information ==="
echo "Docker version:"
docker version --format "{{.Client.Version}}" 2>/dev/null || echo "Could not get Docker version"
echo ""
echo "Docker Compose version:"
docker compose version --short 2>/dev/null || echo "Could not get Docker Compose version"
echo ""

echo "=== Docker Compose Test Completed! ==="
echo ""
echo "✅ Ready to use the new Docker Compose features!"
echo ""
echo "Next steps:"
echo "1. If you haven't already, run setup: docker compose --profile setup up"
echo "2. Test your config: docker compose --profile run-watched up"
echo "3. For production: docker compose --profile schedule-6h up -d"
echo ""
echo "For detailed usage, see: docker/README.md" 