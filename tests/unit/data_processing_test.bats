#!/usr/bin/env bats

# Load the testing helper
load '../test_helper'

# Load the actual code under test
load "${LIB_DIR}/data_processing.sh"

# Setup the test data
setup_test_data() {
    # Create a test ratings file
    cat > "${TEST_TEMP_DIR}/test_ratings.json" << EOF
[
  {
    "rated_at": "2023-06-15T21:36:49.000Z",
    "rating": 8,
    "type": "movie",
    "movie": {
      "title": "Inception",
      "year": 2010,
      "ids": {
        "trakt": 16662,
        "slug": "inception-2010",
        "imdb": "tt1375666",
        "tmdb": 27205
      }
    }
  },
  {
    "rated_at": "2023-06-16T18:22:13.000Z",
    "rating": 9,
    "type": "movie",
    "movie": {
      "title": "The Shawshank Redemption",
      "year": 1994,
      "ids": {
        "trakt": 231,
        "slug": "the-shawshank-redemption-1994",
        "imdb": "tt0111161",
        "tmdb": 278
      }
    }
  }
]
EOF

    # Create a test watched movies file
    cat > "${TEST_TEMP_DIR}/test_watched.json" << EOF
[
  {
    "plays": 2,
    "last_watched_at": "2023-06-15T21:36:49.000Z",
    "movie": {
      "title": "Inception",
      "year": 2010,
      "ids": {
        "trakt": 16662,
        "slug": "inception-2010",
        "imdb": "tt1375666",
        "tmdb": 27205
      }
    }
  },
  {
    "plays": 1,
    "last_watched_at": "2023-06-16T18:22:13.000Z",
    "movie": {
      "title": "The Matrix",
      "year": 1999,
      "ids": {
        "trakt": 481,
        "slug": "the-matrix-1999",
        "imdb": "tt0133093",
        "tmdb": 603
      }
    }
  }
]
EOF

    # Create a test history file
    cat > "${TEST_TEMP_DIR}/test_history.json" << EOF
[
  {
    "id": 123456789,
    "watched_at": "2023-06-20T19:30:15.000Z",
    "action": "watch",
    "type": "movie",
    "movie": {
      "title": "Inception",
      "year": 2010,
      "ids": {
        "trakt": 16662,
        "slug": "inception-2010",
        "imdb": "tt1375666",
        "tmdb": 27205
      }
    }
  },
  {
    "id": 123456790,
    "watched_at": "2023-06-19T20:15:30.000Z",
    "action": "watch",
    "type": "movie",
    "movie": {
      "title": "The Matrix",
      "year": 1999,
      "ids": {
        "trakt": 481,
        "slug": "the-matrix-1999",
        "imdb": "tt0133093",
        "tmdb": 603
      }
    }
  }
]
EOF
}

# Test for create_ratings_lookup function
@test "create_ratings_lookup should create a lookup JSON from ratings" {
    # Setup test data
    setup_test_data
    
    # Files for the test
    local ratings_file="${TEST_TEMP_DIR}/test_ratings.json"
    local output_file="${TEST_TEMP_DIR}/ratings_lookup.json"
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "$log_file"
    
    # Run the function
    run create_ratings_lookup "$ratings_file" "$output_file" "$log_file"
    
    # Check the return code
    assert_success
    
    # Check if the output file exists
    assert_file_exists "$output_file"
    
    # Check if it contains the expected data
    run jq -r '.["16662"]' "$output_file"
    assert_output "8"
    
    run jq -r '.["231"]' "$output_file"
    assert_output "9"
}

# Test for create_ratings_lookup with missing file
@test "create_ratings_lookup should create empty lookup when file is missing" {
    # Files for the test
    local ratings_file="${TEST_TEMP_DIR}/missing_file.json"
    local output_file="${TEST_TEMP_DIR}/empty_ratings_lookup.json"
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "$log_file"
    
    # Run the function
    run create_ratings_lookup "$ratings_file" "$output_file" "$log_file"
    
    # Should return error code
    assert_failure
    
    # Check if the output file exists with empty JSON
    assert_file_exists "$output_file"
    
    # Check if it contains empty JSON
    run cat "$output_file"
    assert_output "{}"
}

# Test for create_plays_count_lookup function
@test "create_plays_count_lookup should create a lookup JSON from watched" {
    # Setup test data
    setup_test_data
    
    # Files for the test
    local watched_file="${TEST_TEMP_DIR}/test_watched.json"
    local output_file="${TEST_TEMP_DIR}/plays_lookup.json"
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "$log_file"
    
    # Run the function
    run create_plays_count_lookup "$watched_file" "$output_file" "$log_file"
    
    # Check the return code
    assert_success
    
    # Check if the output file exists
    assert_file_exists "$output_file"
    
    # Check if it contains the expected data
    run jq -r '.["tt1375666"]' "$output_file"
    assert_output "2"
    
    run jq -r '.["tt0133093"]' "$output_file"
    assert_output "1"
}

# Test for create_plays_count_lookup with missing file
@test "create_plays_count_lookup should create empty lookup when file is missing" {
    # Files for the test
    local watched_file="${TEST_TEMP_DIR}/missing_file.json"
    local output_file="${TEST_TEMP_DIR}/empty_plays_lookup.json"
    local log_file="${TEST_TEMP_DIR}/test.log"
    touch "$log_file"
    
    # Run the function
    run create_plays_count_lookup "$watched_file" "$output_file" "$log_file"
    
    # Should return error code
    assert_failure
    
    # Check if the output file exists with empty JSON
    assert_file_exists "$output_file"
    
    # Check if it contains empty JSON
    run cat "$output_file"
    assert_output "{}"
} 