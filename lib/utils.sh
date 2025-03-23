#!/bin/bash
#
# Utility and debugging functions
#

# Debug messaging function
debug_msg() {
    local message="$1"
    local log_file="$2"
    
    echo -e "DEBUG: $message" | tee -a "${log_file}"
}

# File manipulation debug function
debug_file_info() {
    local file="$1"
    local message="$2"
    local log_file="$3"
    
    echo "📄 $message:" | tee -a "${log_file}"
    if [ -f "$file" ]; then
        echo "  - File exists: ✅" | tee -a "${log_file}"
        echo "  - File size: $(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "unknown") bytes" | tee -a "${log_file}"
        echo "  - File permissions: $(ls -la "$file" | awk '{print $1}')" | tee -a "${log_file}"
        echo "  - Owner: $(ls -la "$file" | awk '{print $3":"$4}')" | tee -a "${log_file}"
        
        # Check if file is readable
        if [ -r "$file" ]; then
            echo "  - File is readable: ✅" | tee -a "${log_file}"
        else
            echo "  - File is readable: ❌" | tee -a "${log_file}"
        fi
        
        # Check if file is writable
        if [ -w "$file" ]; then
            echo "  - File is writable: ✅" | tee -a "${log_file}"
        else
            echo "  - File is writable: ❌" | tee -a "${log_file}"
        fi
        
        # Check if file has content
        if [ -s "$file" ]; then
            echo "  - File has content: ✅" | tee -a "${log_file}"
            echo "  - First line: $(head -n 1 "$file" 2>/dev/null || echo "Cannot read file")" | tee -a "${log_file}"
            echo "  - Line count: $(wc -l < "$file" 2>/dev/null || echo "Cannot count lines")" | tee -a "${log_file}"
        else
            echo "  - File has content: ❌ (empty file)" | tee -a "${log_file}"
        fi
    else
        echo "  - File exists: ❌ (not found)" | tee -a "${log_file}"
        echo "  - Directory exists: $(if [ -d "$(dirname "$file")" ]; then echo "✅"; else echo "❌"; fi)" | tee -a "${log_file}"
        echo "  - Directory permissions: $(ls -la "$(dirname "$file")" 2>/dev/null | head -n 1 | awk '{print $1}' || echo "Cannot access directory")" | tee -a "${log_file}"
    fi
    echo "-----------------------------------" | tee -a "${log_file}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for required dependencies
check_dependencies() {
    local log_file="$1"
    local missing=0
    
    echo "🔍 Checking required dependencies:" | tee -a "${log_file}"
    
    for cmd in curl jq sed awk; do
        if command_exists "$cmd"; then
            echo "  - $cmd: ✅" | tee -a "${log_file}"
        else
            echo "  - $cmd: ❌ (missing)" | tee -a "${log_file}"
            missing=1
        fi
    done
    
    if [ $missing -eq 1 ]; then
        echo "❌ Some required dependencies are missing. Please install them before continuing." | tee -a "${log_file}"
        return 1
    else
        echo "✅ All required dependencies are installed." | tee -a "${log_file}"
        return 0
    fi
}

# Print progress bar
progress_bar() {
    local current="$1"
    local total="$2"
    local prefix="$3"
    local log_file="$4"
    local width=50
    local percentage=$((current * 100 / total))
    local completed=$((width * current / total))
    local remaining=$((width - completed))
    
    printf "\r%s [%s%s] %d%%" "$prefix" "$(printf "%${completed}s" | tr ' ' '=')" "$(printf "%${remaining}s" | tr ' ' ' ')" "$percentage"
    
    if [ "$current" -eq "$total" ]; then
        echo ""
        echo "$prefix completed (100%)" | tee -a "${log_file}"
    fi
}

# Error handling function
handle_error() {
    local error_message="$1"
    local error_code="$2"
    local log_file="$3"
    
    echo "❌ ERROR: $error_message" | tee -a "${log_file}"
    
    if [ -n "$error_code" ]; then
        exit "$error_code"
    fi
} 