#!/usr/bin/env bats

# Load the testing helper
load '../test_helper'

# Load the actual code under test
load "${LIB_DIR}/config.sh"

# Test for init_temp_dir function
@test "init_temp_dir should create an empty directory" {
    local temp_test_dir="${TEST_TEMP_DIR}/TEMP_TEST"
    local log_file="${TEST_TEMP_DIR}/logs/test.log"
    mkdir -p "$(dirname "$log_file")"
    touch "$log_file"
    
    # Create a file in the directory to verify it gets cleared
    mkdir -p "$temp_test_dir"
    touch "$temp_test_dir/test_file.txt"
    
    # Call the function
    run init_temp_dir "$temp_test_dir" "$log_file"
    
    # Verify the directory exists
    assert_dir_exists "$temp_test_dir"
    
    # Verify the directory is empty (the test file should be deleted)
    run ls -A "$temp_test_dir"
    assert_output ""
}

# Test for ensure_directories function
@test "ensure_directories should create directories if they don't exist" {
    local log_dir="${TEST_TEMP_DIR}/test_logs"
    local copy_dir="${TEST_TEMP_DIR}/test_copy"
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "$log_file"
    
    # Remove directories if they exist from previous tests
    rm -rf "$log_dir" "$copy_dir"
    
    # Call the function
    run ensure_directories "$log_dir" "$copy_dir" "$log_file"
    
    # Verify the directories were created
    assert_dir_exists "$log_dir"
    assert_dir_exists "$copy_dir"
}

# Test for detect_os_sed function
@test "detect_os_sed should return correct sed command for the OS" {
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "$log_file"
    
    # Save the original OSTYPE value
    local original_ostype="$OSTYPE"
    
    # Test with macOS
    export OSTYPE="darwin20.0"
    run detect_os_sed "$log_file"
    assert_line --index 0 "sed -i ''"
    
    # Test with Linux
    export OSTYPE="linux-gnu"
    run detect_os_sed "$log_file"
    assert_line --index 0 "sed -i"
    
    # Restore original OSTYPE
    export OSTYPE="$original_ostype"
}

# Test for init_backup_dir function
@test "init_backup_dir should create backup directory if it doesn't exist" {
    local backup_dir="${TEST_TEMP_DIR}/test_backup"
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "$log_file"
    
    # Remove the directory if it exists from previous tests
    rm -rf "$backup_dir"
    
    # Call the function
    run init_backup_dir "$backup_dir" "$log_file"
    
    # Verify the directory was created
    assert_dir_exists "$backup_dir"
    
    # Verify the function returns the backup directory path
    assert_line --index 3 "$backup_dir"
}

# Test for load_config function
@test "load_config should load configuration from the correct location" {
    local script_dir="${TEST_TEMP_DIR}"
    local config_dir="${script_dir}/config"
    local log_file="${TEST_TEMP_DIR}/test.log"
    
    # Create config directory and file
    mkdir -p "$config_dir"
    cat > "${config_dir}/.config.cfg" << EOF
TRAKT_CLIENT_ID="test_client_id"
TRAKT_CLIENT_SECRET="test_client_secret"
TEST_VARIABLE="test_value"
EOF
    
    # Source the load_config function with script_dir and log_file arguments
    source "${LIB_DIR}/config.sh"
    load_config "$script_dir" "$log_file"
    
    # Check if the variables were loaded
    assert_equal "$TRAKT_CLIENT_ID" "test_client_id"
    assert_equal "$TRAKT_CLIENT_SECRET" "test_client_secret"
    assert_equal "$TEST_VARIABLE" "test_value"
} 