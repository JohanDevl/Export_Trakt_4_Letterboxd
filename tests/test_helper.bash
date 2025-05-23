#!/usr/bin/env bash

# Determine directory containing this script
TEST_HELPER_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Set up the test environment
REPO_ROOT="$( cd "${TEST_HELPER_DIR}/.." >/dev/null 2>&1 && pwd )"
LIB_DIR="${REPO_ROOT}/lib"
TESTS_DIR="${REPO_ROOT}/tests"
MOCKS_DIR="${TESTS_DIR}/mocks"
TEST_DATA_DIR="${TESTS_DIR}/data"

# Export directories for use in test files
export REPO_ROOT
export LIB_DIR
export TESTS_DIR
export MOCKS_DIR
export TEST_DATA_DIR

# Load testing libraries - using relative paths from this file
load "${TEST_HELPER_DIR}/helpers/bats-support/load"
load "${TEST_HELPER_DIR}/helpers/bats-assert/load"
load "${TEST_HELPER_DIR}/helpers/bats-file/load"

# Setup the test environment before each test
setup() {
    # Create a temporary directory for test artifacts
    TEST_TEMP_DIR="$(mktemp -d)"
    export TEST_TEMP_DIR
    
    # Set up environment variables for testing
    export TRAKT_CLIENT_ID="test_client_id"
    export TRAKT_CLIENT_SECRET="test_client_secret"
    export TRAKT_REDIRECT_URI="urn:ietf:wg:oauth:2.0:oob"
    export TEST_MODE="true"
    
    # Create mock directories mirroring the main project structure
    mkdir -p "${TEST_TEMP_DIR}/config"
    mkdir -p "${TEST_TEMP_DIR}/logs"
    mkdir -p "${TEST_TEMP_DIR}/backup"
    mkdir -p "${TEST_TEMP_DIR}/TEMP"
    mkdir -p "${TEST_TEMP_DIR}/copy"
}

# Clean up after each test
teardown() {
    # Remove the temporary directory contents first
    if [ -d "${TEST_TEMP_DIR}" ]; then
        # First try to remove all contents
        find "${TEST_TEMP_DIR}" -mindepth 1 -delete 2>/dev/null
        
        # Then try to remove the directory itself
        rmdir "${TEST_TEMP_DIR}" 2>/dev/null || true
    fi
}

# Helper function to create a mock config file
create_mock_config() {
    cat > "${TEST_TEMP_DIR}/config/.config.cfg" << EOF
TRAKT_CLIENT_ID="test_client_id"
TRAKT_CLIENT_SECRET="test_client_secret"
TRAKT_ACCESS_TOKEN="test_access_token"
TRAKT_REFRESH_TOKEN="test_refresh_token"
TRAKT_EXPIRES_IN="7889238"
TRAKT_CREATED_AT="1600000000"
DEBUG_MODE="true"
LOG_LEVEL="DEBUG"
EOF
}

# Helper function to load a mock JSON response
load_mock_response() {
    local mock_file="$1"
    local target_file="$2"
    
    cp "${MOCKS_DIR}/${mock_file}" "${target_file}"
} 