#!/bin/bash

# Run tests and create coverage profile
go test -coverprofile=coverage.txt -covermode=atomic ./pkg/...

# For debug purposes, print the coverage
go tool cover -func=coverage.txt

# Check if we meet the 70% coverage threshold
COVERAGE=$(go tool cover -func=coverage.txt | grep total | awk '{print $3}' | sed 's/%//')
THRESHOLD=70.0

echo "Current test coverage: $COVERAGE%"
echo "Required threshold: $THRESHOLD%"

if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
  echo "Test coverage is below threshold"
  exit 1
else
  echo "Test coverage meets or exceeds threshold"
  exit 0
fi 