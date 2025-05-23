#!/bin/bash
#
# Internationalization (i18n) support for Export Trakt 4 Letterboxd
#

# Get script directory - script is in lib/, so go up one level
SCRIPT_DIR="$( cd "$( dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")" )" && pwd )"
BASE_DIR="$(dirname "$SCRIPT_DIR")"  # This should be the root directory of the project

# Used for error message if language initialization fails
MSG_ERROR_MISSING_LANG_FILE="Error: Language file not found. Using English defaults."

# Global variables
LANG_DIR=""
CURRENT_LANG="en"
AVAILABLE_LANGS=("en" "fr" "es" "de" "it")

# Initialize the i18n system
init_i18n() {
    local script_dir="$1"
    local log_file="$2"
    
    # Set language directory path correctly
    LANG_DIR="${BASE_DIR}/locales"
    
    echo "DEBUG: Language directory set to: $LANG_DIR" | tee -a "${log_file}"
    
    # Check if language directory exists
    if [ ! -d "$LANG_DIR" ]; then
        echo "Creating language directory: $LANG_DIR" | tee -a "${log_file}"
        mkdir -p "$LANG_DIR"
    fi
    
    # Load language from config or use default
    if [ -n "$LANGUAGE" ]; then
        set_language "$LANGUAGE" "$log_file"
    else
        # Try to detect system language
        detect_system_language "$log_file"
    fi
    
    # Load messages for the current language
    load_language_messages "$log_file"
    
    echo "Internationalization initialized. Current language: $CURRENT_LANG" | tee -a "${log_file}"
}

# Set language to use
set_language() {
    local lang="$1"
    local log_file="$2"
    
    # Check if the language is supported
    local is_supported=0
    for supported_lang in "${AVAILABLE_LANGS[@]}"; do
        if [ "$lang" == "$supported_lang" ]; then
            is_supported=1
            break
        fi
    done
    
    if [ $is_supported -eq 1 ]; then
        CURRENT_LANG="$lang"
        echo "Language set to: $CURRENT_LANG" | tee -a "${log_file}"
    else
        echo "Warning: Language '$lang' is not supported. Using default (en)." | tee -a "${log_file}"
        CURRENT_LANG="en"
    fi
}

# Detect system language
detect_system_language() {
    local log_file="$1"
    local system_lang=""
    
    # Try to get system language
    if [ -n "$LANG" ]; then
        system_lang="${LANG:0:2}"
    elif [ -n "$LC_ALL" ]; then
        system_lang="${LC_ALL:0:2}"
    elif [ -n "$LC_MESSAGES" ]; then
        system_lang="${LC_MESSAGES:0:2}"
    fi
    
    echo "Detected system language: ${system_lang:-unknown}" | tee -a "${log_file}"
    
    # If we got a valid language and it's supported, use it
    if [ -n "$system_lang" ]; then
        set_language "$system_lang" "$log_file"
    else
        # Otherwise use default
        CURRENT_LANG="en"
        echo "Using default language: en" | tee -a "${log_file}"
    fi
}

# Load messages for the current language
load_language_messages() {
    local log_file="$1"
    local messages_file="${LANG_DIR}/${CURRENT_LANG}/LC_MESSAGES/messages.sh"
    
    echo "DEBUG: Looking for messages file at: $messages_file" | tee -a "${log_file}"
    
    # Start with loading default English messages
    local default_messages_file="${LANG_DIR}/en/LC_MESSAGES/messages.sh"
    if [ -f "$default_messages_file" ]; then
        echo "DEBUG: Loading default English messages from: $default_messages_file" | tee -a "${log_file}"
        source "$default_messages_file"
    else
        echo "Warning: Default English messages file not found at $default_messages_file. Using internal defaults." | tee -a "${log_file}"
        load_default_messages
    fi
    
    # If current language is not English, load the specific language file to override defaults
    if [ "$CURRENT_LANG" != "en" ] && [ -f "$messages_file" ]; then
        echo "DEBUG: Loading language specific messages from: $messages_file" | tee -a "${log_file}"
        source "$messages_file"
        echo "Loaded language messages from: $messages_file" | tee -a "${log_file}"
    elif [ "$CURRENT_LANG" != "en" ]; then
        echo "Warning: No messages file found for language '$CURRENT_LANG' at $messages_file. Using English defaults." | tee -a "${log_file}"
    fi
}

# Load default (English) messages if translation file is not found
load_default_messages() {
    # Define default English messages
    MSG_welcome="Welcome to Export Trakt 4 Letterboxd"
    MSG_starting="Starting script"
    MSG_script_execution_start="Script execution started"
    MSG_processing_option="Processing option"
    MSG_no_option="No option provided, using default"
    MSG_retrieving_info="Retrieving information"
    MSG_checking_dependencies="Checking required dependencies"
    MSG_missing_dependencies="Some required dependencies are missing. Please install them before continuing."
    MSG_all_dependencies_installed="All required dependencies are installed."
    MSG_environment_info="Environment information"
    MSG_existing_csv_check="Existing CSV file check"
    MSG_error="ERROR"
    MSG_warning="WARNING"
    MSG_success="SUCCESS"
    MSG_script_complete="Script execution completed"
    MSG_running_docker="Running in Docker container"
    MSG_script_option="Script option"
    MSG_none="none"
    MSG_user="User"
    MSG_working_directory="Working directory"
    MSG_script_directory="Script directory"
    MSG_copy_directory="Copy directory"
    MSG_log_directory="Log directory"
    MSG_backup_directory="Backup directory"
    MSG_os_type="OS Type"
    MSG_file_exists="File exists"
    MSG_file_is_readable="File is readable"
    MSG_file_is_writable="File is writable"
    MSG_file_has_content="File has content"
    MSG_file_exists_not="File not found"
    MSG_directory_exists="Directory exists"
    MSG_directory_permissions="Directory permissions"
    MSG_created_backup_directory="Created backup directory"
    MSG_backup_directory_exists="Backup directory exists"
    MSG_backup_directory_writable="Backup directory is writable"
    MSG_backup_directory_not_writable="WARNING: Backup directory is not writable. Check permissions."
    MSG_language_set="Language set to"
    MSG_running_in="running on"
    MSG_auto_detected="auto-detected"
}

# Get a translated message
get_message() {
    local message_key="$1"
    local default_message="$2"
    local var_name="MSG_${message_key}"
    
    # If the variable exists, return its value
    if [ -n "${!var_name}" ]; then
        echo "${!var_name}"
    else
        # Otherwise return default message if provided
        if [ -n "$default_message" ]; then
            echo "$default_message"
        else
            # If no default message, return the key itself
            echo "$message_key"
        fi
    fi
}

# Translate a message (alias for get_message)
_() {
    get_message "$@"
}

# List available languages
list_languages() {
    local log_file="$1"
    
    echo "Available languages:" | tee -a "${log_file}"
    for lang in "${AVAILABLE_LANGS[@]}"; do
        if [ "$lang" == "$CURRENT_LANG" ]; then
            echo "  - $lang (current)" | tee -a "${log_file}"
        else
            echo "  - $lang" | tee -a "${log_file}"
        fi
    done
}

# Create a new language file template
create_language_template() {
    local lang="$1"
    local log_file="$2"
    
    # Check if language code is valid
    if [ -z "$lang" ] || [ ${#lang} -ne 2 ]; then
        echo "Invalid language code. Please use a 2-letter ISO language code (e.g., 'en', 'fr')." | tee -a "${log_file}"
        return 1
    fi
    
    # Create directory if it doesn't exist
    local lang_dir="${LANG_DIR}/${lang}/LC_MESSAGES"
    if [ ! -d "$lang_dir" ]; then
        mkdir -p "$lang_dir"
        echo "Created directory: $lang_dir" | tee -a "${log_file}"
    fi
    
    # Create template file
    local template_file="${lang_dir}/messages.sh"
    if [ -f "$template_file" ]; then
        echo "Warning: File already exists: $template_file" | tee -a "${log_file}"
        read -p "Overwrite? (y/N): " confirm
        if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
            echo "Aborted." | tee -a "${log_file}"
            return 1
        fi
    fi
    
    # Get the English template as the base
    local en_template="${LANG_DIR}/en/LC_MESSAGES/messages.sh"
    if [ -f "$en_template" ]; then
        cp "$en_template" "$template_file"
        echo "Created language template from English messages: $template_file" | tee -a "${log_file}"
    else
        # Create template content
        echo "#!/bin/bash" > "$template_file"
        echo "#" >> "$template_file"
        echo "# Language: $lang" >> "$template_file"
        echo "#" >> "$template_file"
        echo "" >> "$template_file"
        echo "# Define messages for $lang" >> "$template_file"
        echo "# Variables must start with MSG_ to be recognized by the system" >> "$template_file"
        echo "" >> "$template_file"
        
        # Add default messages
        load_default_messages
        
        # Get a list of all message variables
        local msg_vars=$(set | grep '^MSG_' | cut -d= -f1)
        
        # Add each message key with its English value for translation
        for var in $msg_vars; do
            # Get the message value
            local value=${!var}
            echo "$var=\"$value\"" >> "$template_file"
        done
        
        echo "Created language template from default messages: $template_file" | tee -a "${log_file}"
    fi
    
    return 0
} 