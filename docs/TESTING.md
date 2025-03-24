# Testing Framework Documentation

This document provides detailed information about the testing framework implemented for Export_Trakt_4_Letterboxd.

## Overview

The testing framework is designed to ensure code quality, prevent regressions, and make it easier to add new features. It consists of unit tests for core functions, integration tests for the export process, and a test runner script that can generate code coverage reports.

## Testing Structure

The tests are organized in the following directory structure:

```
tests/
├── unit/                   # Unit tests for individual functions
│   ├── config_test.bats    # Tests for config.sh
│   ├── trakt_api_test.bats # Tests for trakt_api.sh
│   └── data_processing_test.bats # Tests for data_processing.sh
├── integration/            # Integration tests
│   └── export_process_test.bats # Test for complete export process
├── mocks/                  # Mock API responses
│   ├── ratings.json        # Mock ratings data
│   ├── history.json        # Mock history data
│   ├── watchlist.json      # Mock watchlist data
│   └── trakt_api_mock.sh   # Mock API functions
├── helpers/                # Bats helper libraries
│   ├── bats-assert/        # Assertion library for Bats
│   ├── bats-support/       # Support functions for Bats
│   └── bats-file/          # File-related assertions for Bats
├── bats/                   # Bats core test framework
├── data/                   # Test data files
├── test_helper.bash        # Common setup for all tests
└── run_tests.sh            # Script to run tests and generate coverage reports
```

## Dependencies

The testing framework relies on the following tools:

1. **Bats (Bash Automated Testing System)**: A TAP-compliant testing framework for Bash
2. **jq**: A lightweight and flexible command-line JSON processor
3. **kcov** (optional): For generating code coverage reports

## Installation

The Bats testing framework and its helper libraries are installed as Git submodules. To initialize them:

```bash
git submodule update --init --recursive
```

For jq and kcov, install them using your package manager:

```bash
# Debian/Ubuntu
apt-get install jq kcov

# macOS
brew install jq kcov
```

## Running Tests

### Basic Test Run

To run all tests:

```bash
./tests/run_tests.sh
```

This will execute all unit and integration tests and provide a summary of the results.

### Code Coverage Reports

To generate a code coverage report:

```bash
./tests/run_tests.sh coverage
```

The coverage report will be available in HTML format at `test-results/coverage/index.html`.

## Writing Tests

### Unit Tests

Unit tests should test individual functions in isolation. Example:

```bash
@test "function_name should do something specific" {
    # Setup test environment
    local input="test input"
    local expected="expected output"

    # Run the function
    run function_name "$input"

    # Assert the results
    assert_success
    assert_output "$expected"
}
```

### Integration Tests

Integration tests should test the interaction between multiple components:

```bash
@test "Integration: Export process should produce valid CSV files" {
    # Setup the integration test environment
    setup_integration_test

    # Run the export process
    run ./export_script.sh

    # Verify the output
    assert_success
    assert_file_exists "output.csv"
}
```

## Mocking API Calls

The framework includes mock functions for the Trakt API to avoid making real API calls during tests. To use the mock functions:

```bash
# Load the mock API functions
source "${TESTS_DIR}/mocks/trakt_api_mock.sh"

# Enable test mode
export TEST_MODE="true"

# Now API calls will use mock data
```

## Test Helper Functions

Common test helper functions are defined in `test_helper.bash`:

- `setup()`: Called before each test to set up the test environment
- `teardown()`: Called after each test to clean up
- `create_mock_config()`: Creates a mock configuration file for testing
- `load_mock_response()`: Loads a mock API response

## Continuous Integration

The tests are automatically run in the GitHub Actions CI/CD pipeline for every pull request. The workflow is defined in `.github/workflows/docker-test.yml`.

## Best Practices

When writing tests, follow these best practices:

1. **Isolation**: Tests should be independent of each other
2. **Cleanup**: Always clean up temporary files in the teardown function
3. **Mock External Dependencies**: Use mock functions for external APIs and services
4. **Test Edge Cases**: Include tests for error conditions and edge cases
5. **Keep Tests Fast**: Tests should run quickly to provide rapid feedback
6. **Descriptive Names**: Use descriptive test names that explain what is being tested

## Troubleshooting

Common issues and solutions:

- **Test not found**: Ensure the test file is executable and follows the naming convention `*_test.bats`
- **Bats command not found**: Run `git submodule update --init --recursive` to install Bats
- **jq not found**: Install jq using your package manager
- **Coverage report not generated**: Install kcov and ensure it's in your PATH
