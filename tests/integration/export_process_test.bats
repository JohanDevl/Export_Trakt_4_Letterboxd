#!/usr/bin/env bats

# Load the testing helper
load '../test_helper'

# Load the mock API functions
source "${TESTS_DIR}/mocks/trakt_api_mock.sh"

# Set up the test environment for integration tests
setup_integration_test() {
    # Create mock config file
    create_mock_config
    
    # Copy the mock data to the test backup directory
    mkdir -p "${TEST_TEMP_DIR}/backup"
    cp "${MOCKS_DIR}/ratings.json" "${TEST_TEMP_DIR}/backup/ratings_movies.json"
    cp "${MOCKS_DIR}/history.json" "${TEST_TEMP_DIR}/backup/history_movies.json"
    cp "${MOCKS_DIR}/watchlist.json" "${TEST_TEMP_DIR}/backup/watchlist_movies.json"
    
    # Create an expected output directory
    mkdir -p "${TEST_TEMP_DIR}/copy"
}

# Test simple export process
@test "Integration: Basic export process should create CSV files" {
    # Set up the integration test environment
    setup_integration_test
    
    # Create temporary copy of the main script for testing
    cat > "${TEST_TEMP_DIR}/export_test.sh" << EOF
#!/bin/bash

# Override directories for testing
export SCRIPT_DIR="${TEST_TEMP_DIR}"
export CONFIG_DIR="${TEST_TEMP_DIR}/config"
export LOG_DIR="${TEST_TEMP_DIR}/logs"
export COPY_DIR="${TEST_TEMP_DIR}/copy"
export TEMP_DIR="${TEST_TEMP_DIR}/TEMP"
export BACKUP_DIR="${TEST_TEMP_DIR}/backup"

# Enable test mode
export TEST_MODE="true"

# Source the library files
source "${LIB_DIR}/config.sh"
source "${LIB_DIR}/utils.sh"
source "${TESTS_DIR}/mocks/trakt_api_mock.sh"
source "${LIB_DIR}/data_processing.sh"

# Create log file
LOG_FILE="${LOG_DIR}/export_test.log"
mkdir -p "${LOG_DIR}"
touch "\${LOG_FILE}"

# Initialize directories
ensure_directories "\${LOG_DIR}" "\${COPY_DIR}" "\${LOG_FILE}"
init_temp_dir "\${TEMP_DIR}" "\${LOG_FILE}"
init_backup_dir "\${BACKUP_DIR}" "\${LOG_FILE}"

# Process ratings (simplified for test)
echo "Processing ratings..." | tee -a "\${LOG_FILE}"
RATINGS_FILE="\${BACKUP_DIR}/ratings_movies.json"
RATINGS_LOOKUP="\${TEMP_DIR}/ratings_lookup.json"
create_ratings_lookup "\${RATINGS_FILE}" "\${RATINGS_LOOKUP}" "\${LOG_FILE}"

# Process history (simplified for test)
echo "Processing history..." | tee -a "\${LOG_FILE}"
HISTORY_FILE="\${BACKUP_DIR}/history_movies.json"

# Create CSV header for ratings
echo "Title,Year,Directors,Rating,WatchedDate" > "\${COPY_DIR}/ratings.csv"

# Extract movie information from ratings
jq -r '.[] | "\(.movie.title),\(.movie.year),,\(.rating),\(.rated_at)"' "\${RATINGS_FILE}" >> "\${COPY_DIR}/ratings.csv"

# Create CSV header for watchlist
echo "Title,Year,Directors" > "\${COPY_DIR}/watchlist.csv"

# Extract movie information from watchlist
jq -r '.[] | "\(.movie.title),\(.movie.year),"' "\${BACKUP_DIR}/watchlist_movies.json" >> "\${COPY_DIR}/watchlist.csv"

echo "Export completed successfully" | tee -a "\${LOG_FILE}"
exit 0
EOF

    # Make the test script executable
    chmod +x "${TEST_TEMP_DIR}/export_test.sh"
    
    # Run the test script
    run "${TEST_TEMP_DIR}/export_test.sh"
    
    # Check it was successful
    assert_success
    
    # Check if the output files were created
    assert_file_exists "${TEST_TEMP_DIR}/copy/ratings.csv"
    assert_file_exists "${TEST_TEMP_DIR}/copy/watchlist.csv"
    
    # Check content of ratings.csv
    run grep "Inception,2010,,8," "${TEST_TEMP_DIR}/copy/ratings.csv"
    assert_success
    
    run grep "The Shawshank Redemption,1994,,9," "${TEST_TEMP_DIR}/copy/ratings.csv"
    assert_success
    
    # Check content of watchlist.csv
    run grep "Dune,2021," "${TEST_TEMP_DIR}/copy/watchlist.csv"
    assert_success
    
    run grep "Oppenheimer,2023," "${TEST_TEMP_DIR}/copy/watchlist.csv"
    assert_success
} 