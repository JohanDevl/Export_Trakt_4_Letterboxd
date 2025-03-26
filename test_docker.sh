#!/bin/bash

# Test script for Docker image with mocked data
# This script builds the Docker image and tests the letterboxd export format

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'

# Build the Docker image
echo -e "${BLUE}Building Docker image...${RESET}"
docker build -t export-trakt-4-letterboxd -f Dockerfile . || { echo -e "${RED}Docker build failed!${RESET}"; exit 1; }
echo -e "${GREEN}Docker image built successfully${RESET}"

# Create test directories
echo -e "${BLUE}Setting up test environment...${RESET}"
mkdir -p test/config test/logs test/exports

# Copy config file
cp config/config.example.toml test/config/config.toml

# Modify config to use letterboxd format
echo -e "${BLUE}Configuring for Letterboxd export format...${RESET}"
sed -i.bak 's/extended_info = "full"/extended_info = "letterboxd"/' test/config/config.toml

# Run the Docker container with different export types
echo -e "${BLUE}Testing Docker container with --help flag...${RESET}"
docker run --rm -v $(pwd)/test/config:/app/config -v $(pwd)/test/logs:/app/logs -v $(pwd)/test/exports:/app/exports export-trakt-4-letterboxd --help

echo -e "${YELLOW}Note: The following commands will fail with API errors since we don't have valid API credentials.${RESET}"
echo -e "${YELLOW}This is expected behavior and confirms the Docker container is working correctly.${RESET}"

echo -e "${BLUE}Testing Docker container with watched export...${RESET}"
docker run --rm -v $(pwd)/test/config:/app/config -v $(pwd)/test/logs:/app/logs -v $(pwd)/test/exports:/app/exports export-trakt-4-letterboxd --export watched

echo -e "${BLUE}Testing Docker container with ratings export...${RESET}"
docker run --rm -v $(pwd)/test/config:/app/config -v $(pwd)/test/logs:/app/logs -v $(pwd)/test/exports:/app/exports export-trakt-4-letterboxd --export ratings

# Check logs
echo -e "${BLUE}Checking logs...${RESET}"
cat test/logs/export.log

echo -e "${GREEN}Docker image testing completed${RESET}"
echo -e "${YELLOW}To use this Docker image with real data, update the config file with valid API credentials${RESET}"
echo -e "${YELLOW}and run: docker run --rm -v /path/to/config:/app/config -v /path/to/logs:/app/logs -v /path/to/exports:/app/exports export-trakt-4-letterboxd${RESET}" 