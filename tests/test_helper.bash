#!/usr/bin/env bash

# Set up the test environment
REPO_ROOT="$(cd "$(dirname "${BATS_TEST_DIRNAME}")" && pwd)"
LIB_DIR="${REPO_ROOT}/lib"
TESTS_DIR="${REPO_ROOT}/tests"
MOCKS_DIR="${TESTS_DIR}/mocks"
TEST_DATA_DIR="${TESTS_DIR}/data"

# Load testing libraries
load "${TESTS_DIR}/helpers/bats-support/load"
load "${TESTS_DIR}/helpers/bats-assert/load"
load "${TESTS_DIR}/helpers/bats-file/load"

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
    # Remove the temporary directory
    rm -rf "${TEST_TEMP_DIR}"
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