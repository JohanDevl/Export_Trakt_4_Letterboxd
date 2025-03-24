#!/usr/bin/env bats

# Load the testing helper with a simple relative path
load "../test_helper"

# Test initialization of temporary directory
@test "init_temp_dir creates and cleans directory" {
    # Source the function under test
    source "${LIB_DIR}/config.sh"
    
    # Create a test directory
    local test_dir="${TEST_TEMP_DIR}/init_test"
    mkdir -p "${test_dir}"
    
    # Add some files
    touch "${test_dir}/file1.txt"
    mkdir -p "${test_dir}/subdir"
    touch "${test_dir}/subdir/file2.txt"
    
    # Create a log file
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "${log_file}"
    
    # Run the function under test
    run init_temp_dir "${test_dir}" "${log_file}"
    
    # Check that it succeeded
    assert_success
    
    # Check that directory exists but is empty
    assert_dir_exists "${test_dir}"
    run find "${test_dir}" -type f
    assert_output ""
}

@test "ensure_directories creates directories when they don't exist" {
    # Source the function under test
    source "${LIB_DIR}/config.sh"
    
    # Define test directories
    local dir1="${TEST_TEMP_DIR}/dir1"
    local dir2="${TEST_TEMP_DIR}/dir2"
    
    # Create a log file
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "${log_file}"
    
    # Ensure the directories don't exist initially
    rm -rf "${dir1}" "${dir2}"
    
    # Run the function under test
    run ensure_directories "${dir1}" "${dir2}" "${log_file}"
    
    # Check that it succeeded
    assert_success
    
    # Check that directories were created
    assert_dir_exists "${dir1}"
    assert_dir_exists "${dir2}"
}

@test "detect_os_sed returns correct sed command for OS" {
    # Source the function under test
    source "${LIB_DIR}/config.sh"
    
    # Create a log file
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "${log_file}"
    
    # Run the function under test
    run detect_os_sed "${log_file}"
    
    # Check that it succeeded
    assert_success
    
    # On macOS, it should return 'sed -i "" "s///"'
    # On Linux, it should return 'sed -i "s///"'
    if [[ "$OSTYPE" == "darwin"* ]]; then
        assert_output --partial "sed -i ''"
    else
        assert_output --partial "sed -i"
    fi
}

@test "init_backup_dir creates and preserves directory" {
    # Source the function under test
    source "${LIB_DIR}/config.sh"
    
    # Create a test backup directory
    local backup_dir="${TEST_TEMP_DIR}/backup_test"
    mkdir -p "${backup_dir}"
    
    # Create a log file
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "${log_file}"
    
    # Add a file that should be preserved
    touch "${backup_dir}/ratings_movies.json"
    
    # Run the function under test
    run init_backup_dir "${backup_dir}" "${log_file}"
    
    # Check that it succeeded
    assert_success
    
    # Check that directory exists
    assert_dir_exists "${backup_dir}"
    
    # Check that the existing file was preserved
    assert_file_exists "${backup_dir}/ratings_movies.json"
}

@test "load_config loads configuration values" {
    # Create a mock config file with environment variables
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
    
    # Create a log file
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "${log_file}"
    
    # Define a wrapper function that sources config.sh and calls load_config
    test_load_config() {
        source "${LIB_DIR}/config.sh"
        load_config "${TEST_TEMP_DIR}" "${log_file}"
        echo "TRAKT_CLIENT_ID=${TRAKT_CLIENT_ID}"
        echo "TRAKT_CLIENT_SECRET=${TRAKT_CLIENT_SECRET}"
        echo "TRAKT_ACCESS_TOKEN=${TRAKT_ACCESS_TOKEN}"
        echo "TRAKT_REFRESH_TOKEN=${TRAKT_REFRESH_TOKEN}"
        echo "DEBUG_MODE=${DEBUG_MODE}"
    }
    
    # Run the wrapper function
    run test_load_config
    
    # Check that it succeeded
    assert_success
    
    # Check that config values were loaded correctly
    assert_output --partial "TRAKT_CLIENT_ID=test_client_id"
    assert_output --partial "TRAKT_CLIENT_SECRET=test_client_secret"
    assert_output --partial "TRAKT_ACCESS_TOKEN=test_access_token"
    assert_output --partial "TRAKT_REFRESH_TOKEN=test_refresh_token"
    assert_output --partial "DEBUG_MODE=true"
} 