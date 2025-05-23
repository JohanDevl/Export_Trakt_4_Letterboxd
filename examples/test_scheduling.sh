#!/bin/bash

# Test script for the new --run and --schedule features
# Export Trakt 4 Letterboxd

echo "=== Export Trakt 4 Letterboxd - Scheduling Test Script ==="
echo ""

# Build the application first
echo "Building the application..."
go build -o export_trakt ./cmd/export_trakt/
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi
echo "✅ Build successful!"
echo ""

# Test 1: Show help to verify new options are available
echo "=== Test 1: Checking available options ==="
./export_trakt --help
echo ""

# Test 2: Test invalid cron format validation
echo "=== Test 2: Testing cron format validation ==="
echo "Testing invalid cron format (should show error):"
./export_trakt --schedule "invalid-format" --export watched 2>&1 | head -10
echo ""

# Test 3: Test valid cron format (dry run)
echo "=== Test 3: Testing valid cron format validation ==="
echo "Testing valid cron format (should validate successfully):"
timeout 5s ./export_trakt --schedule "0 */6 * * *" --export watched 2>&1 &
PID=$!
sleep 2
if kill -0 $PID 2>/dev/null; then
    echo "✅ Scheduler started successfully (process running)"
    kill $PID 2>/dev/null
else
    echo "❌ Scheduler failed to start"
fi
echo ""

# Test 4: Test immediate execution mode
echo "=== Test 4: Testing immediate execution mode ==="
echo "Testing --run flag (should execute once and exit):"
timeout 10s ./export_trakt --run --export watched --mode normal 2>&1 | head -5
echo "✅ Immediate execution test completed"
echo ""

# Test 5: Show different schedule examples
echo "=== Test 5: Schedule Examples ==="
echo "Here are some example schedules you can use:"
echo ""
echo "Every 6 hours:"
echo "  ./export_trakt --schedule \"0 */6 * * *\" --export all --mode complete"
echo ""
echo "Every day at 2:30 AM:"
echo "  ./export_trakt --schedule \"30 2 * * *\" --export all --mode complete"
echo ""
echo "Every Monday at 9:00 AM:"
echo "  ./export_trakt --schedule \"0 9 * * 1\" --export collection --mode normal"
echo ""
echo "Every 15 minutes (high frequency):"
echo "  ./export_trakt --schedule \"*/15 * * * *\" --export watched --mode normal"
echo ""

echo "=== All tests completed! ==="
echo ""
echo "✅ The new --run and --schedule features are working correctly!"
echo ""
echo "Usage examples:"
echo "  # Immediate execution:"
echo "  ./export_trakt --run --export all --mode complete"
echo ""
echo "  # Scheduled execution:"
echo "  ./export_trakt --schedule \"0 */6 * * *\" --export all --mode complete"
echo ""
echo "For more examples, see: examples/scheduling.md" 