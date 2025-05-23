#!/usr/bin/env bash
# Mock functions for trakt_api.sh

# Mock for get_trakt_ratings
get_trakt_ratings() {
    local type="$1"
    local output_file="$2"
    
    if [[ "$TEST_MODE" == "true" ]]; then
        # In test mode, use the mock data
        cp "${MOCKS_DIR}/ratings.json" "$output_file"
        return 0
    fi
    
    # Original function would be called here if not in test mode
    return 1
}

# Mock for get_trakt_history
get_trakt_history() {
    local start_date="$1"
    local output_file="$2"
    
    if [[ "$TEST_MODE" == "true" ]]; then
        # In test mode, use the mock data
        cp "${MOCKS_DIR}/history.json" "$output_file"
        return 0
    fi
    
    # Original function would be called here if not in test mode
    return 1
}

# Mock for get_trakt_watchlist
get_trakt_watchlist() {
    local type="$1"
    local output_file="$2"
    
    if [[ "$TEST_MODE" == "true" ]]; then
        # In test mode, use the mock data
        cp "${MOCKS_DIR}/watchlist.json" "$output_file"
        return 0
    fi
    
    # Original function would be called here if not in test mode
    return 1
}

# Mock for refresh_token
refresh_token() {
    if [[ "$TEST_MODE" == "true" ]]; then
        # In test mode, just pretend we refreshed the token
        echo "Token refreshed (mock)"
        return 0
    fi
    
    # Original function would be called here if not in test mode
    return 1
}

# Mock for check_token_validity
check_token_validity() {
    if [[ "$TEST_MODE" == "true" ]]; then
        # In test mode, pretend the token is valid
        return 0
    fi
    
    # Original function would be called here if not in test mode
    return 1
}

# Export the mock functions
export -f get_trakt_ratings
export -f get_trakt_history
export -f get_trakt_watchlist
export -f refresh_token
export -f check_token_validity 