#!/usr/bin/env bash
#
# Run all tests and generate coverage report
#

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$(dirname "${SCRIPT_DIR}")" && pwd)"
BATS_DIR="${SCRIPT_DIR}/bats"
TEST_RESULTS_DIR="${REPO_ROOT}/test-results"

# Create test results directory
mkdir -p "${TEST_RESULTS_DIR}"

# Output title
echo "🧪 Running tests for Export_Trakt_4_Letterboxd..."
echo "=============================================="

# Check if we have the required tools
check_dependencies() {
    local missing=0
    
    # Check for bats
    if [ ! -d "${BATS_DIR}" ]; then
        echo "❌ bats-core not found (run: git submodule update --init --recursive)"
        missing=1
    fi
    
    # Check for jq
    if ! command -v jq &> /dev/null; then
        echo "❌ jq not found (install with your package manager)"
        missing=1
    fi
    
    # Check for kcov if coverage is requested
    if [ "$1" = "coverage" ] && ! command -v kcov &> /dev/null; then
        echo "❌ kcov not found (install with your package manager for test coverage)"
        missing=1
    fi
    
    if [ $missing -eq 1 ]; then
        echo "Please install the missing dependencies and try again."
        exit 1
    fi
}

# Run the tests
run_tests() {
    echo "🔍 Running unit tests..."
    "${BATS_DIR}/bin/bats" "${SCRIPT_DIR}/unit" | tee "${TEST_RESULTS_DIR}/unit_tests.log"
    local unit_status=${PIPESTATUS[0]}
    
    echo -e "\n🔍 Running integration tests..."
    "${BATS_DIR}/bin/bats" "${SCRIPT_DIR}/integration" | tee "${TEST_RESULTS_DIR}/integration_tests.log"
    local integration_status=${PIPESTATUS[0]}
    
    # Return success only if all tests passed
    if [ $unit_status -eq 0 ] && [ $integration_status -eq 0 ]; then
        echo -e "\n✅ All tests passed!"
        return 0
    else
        echo -e "\n❌ Some tests failed!"
        return 1
    fi
}

# Generate coverage report
generate_coverage() {
    echo -e "\n📊 Generating test coverage report..."
    
    # Create coverage directory
    mkdir -p "${TEST_RESULTS_DIR}/coverage"
    
    # Run the tests with kcov for coverage reporting
    kcov --include-path="${REPO_ROOT}/lib" \
         "${TEST_RESULTS_DIR}/coverage" \
         "${BATS_DIR}/bin/bats" "${SCRIPT_DIR}/unit" "${SCRIPT_DIR}/integration"
    
    echo "Coverage report generated at: ${TEST_RESULTS_DIR}/coverage/index.html"
}

# Main execution
check_dependencies "$1"

# Run the tests
run_tests
TEST_STATUS=$?

# Generate coverage if requested
if [ "$1" = "coverage" ]; then
    generate_coverage
fi

echo -e "\n📝 Test summary:"
echo "Unit tests: $(grep "tests," "${TEST_RESULTS_DIR}/unit_tests.log" | tail -n 1)"
echo "Integration tests: $(grep "tests," "${TEST_RESULTS_DIR}/integration_tests.log" | tail -n 1)"

exit $TEST_STATUS 