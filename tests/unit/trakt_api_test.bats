#!/usr/bin/env bats

# Load the testing helper
load '../test_helper'

# Load the API mock functions
source "${TESTS_DIR}/mocks/trakt_api_mock.sh"

# Test for API mocking of get_trakt_ratings
@test "get_trakt_ratings should return ratings from mock data" {
    # Set up test mode and output file
    export TEST_MODE="true"
    local output_file="${TEST_TEMP_DIR}/test_ratings.json"
    
    # Call function
    get_trakt_ratings "movies" "$output_file"
    
    # Verify the file was created with mock data
    assert_file_exists "$output_file"
    
    # Check if the file contains expected data
    run jq -r '.[0].movie.title' "$output_file"
    assert_output "Inception"
    
    run jq -r '.[1].movie.title' "$output_file"
    assert_output "The Shawshank Redemption"
    
    run jq -r '.[2].movie.title' "$output_file"
    assert_output "The Dark Knight"
}

# Test for API mocking of get_trakt_history
@test "get_trakt_history should return history from mock data" {
    # Set up test mode and output file
    export TEST_MODE="true"
    local output_file="${TEST_TEMP_DIR}/test_history.json"
    
    # Call function
    get_trakt_history "2023-01-01" "$output_file"
    
    # Verify the file was created with mock data
    assert_file_exists "$output_file"
    
    # Check if the file contains expected data
    run jq -r '.[0].movie.title' "$output_file"
    assert_output "Inception"
    
    run jq -r '.[1].movie.title' "$output_file"
    assert_output "The Matrix"
    
    run jq -r '.[2].movie.title' "$output_file"
    assert_output "Pulp Fiction"
}

# Test for API mocking of get_trakt_watchlist
@test "get_trakt_watchlist should return watchlist from mock data" {
    # Set up test mode and output file
    export TEST_MODE="true"
    local output_file="${TEST_TEMP_DIR}/test_watchlist.json"
    
    # Call function
    get_trakt_watchlist "movies" "$output_file"
    
    # Verify the file was created with mock data
    assert_file_exists "$output_file"
    
    # Check if the file contains expected data
    run jq -r '.[0].movie.title' "$output_file"
    assert_output "Dune"
    
    run jq -r '.[1].movie.title' "$output_file"
    assert_output "Oppenheimer"
    
    run jq -r '.[2].movie.title' "$output_file"
    assert_output "The Batman"
}

# Test for check_token_validity
@test "check_token_validity should return success in test mode" {
    # Set up test mode
    export TEST_MODE="true"
    
    # Call function
    run check_token_validity
    
    # It should return success (0) in test mode
    assert_success
}

# Test for refresh_token
@test "refresh_token should return success message in test mode" {
    # Set up test mode
    export TEST_MODE="true"
    
    # Call function 
    run refresh_token
    
    # It should return success (0) in test mode
    assert_success
    
    # It should output the mock message
    assert_output "Token refreshed (mock)"
} 