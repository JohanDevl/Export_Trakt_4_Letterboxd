#!/bin/bash
#
# Translation Management Utility
# This script helps manage language files for Export Trakt 4 Letterboxd
#

# Enable error handling
set -o pipefail

# We use PWD to get the current working directory
SCRIPT_DIR="$(pwd)"
LANG_DIR="${SCRIPT_DIR}/locales"
DEFAULT_LANG="en"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Debug flag (set to true to enable debug output)
DEBUG=true

# Debug information function
debug_log() {
    if [ "$DEBUG" = true ]; then
        echo -e "${BLUE}DEBUG:${NC} $1"
    fi
}

# Error function
error_log() {
    echo -e "${RED}ERROR:${NC} $1" >&2
}

# Success function
success_log() {
    echo -e "${GREEN}SUCCESS:${NC} $1"
}

# Warning function
warning_log() {
    echo -e "${YELLOW}WARNING:${NC} $1"
}

# Initialize environment
initialize() {
    debug_log "Working directory: ${SCRIPT_DIR}"
    debug_log "Language directory: ${LANG_DIR}"

    if [ ! -d "$LANG_DIR" ]; then
        mkdir -p "$LANG_DIR"
        success_log "Created language directory: $LANG_DIR"
    fi

    # Check for default language
    if [ ! -d "${LANG_DIR}/${DEFAULT_LANG}" ]; then
        warning_log "Default language directory (${DEFAULT_LANG}) not found"
    fi

    # Add debug information
    debug_log "Looking for language files in: $LANG_DIR"
    if [ "$DEBUG" = true ]; then
        find "$LANG_DIR" -type f -name "messages.sh" -print
    fi
}

# Function to display help
show_help() {
    cat << EOF
Translation Management Utility
------------------------------
Usage: $0 [command] [options]

Commands:
  list                      List available languages
  create <language_code>    Create a new language template
  update                    Update all language files with new strings from default language
  status                    Show translation status for all languages
  check                     Validate translation files for errors
  export <format>           Export translations to a specified format (json, po)
  import <file>             Import translations from external file
  help                      Display this help message

Examples:
  $0 list
  $0 create de
  $0 update
  $0 status
  $0 check
  $0 export json
  $0 import translations.po

Language codes should be 2-letter ISO language codes (e.g., 'en', 'fr', 'es', 'de').
EOF
}

# Initialize the environment
initialize

# Function to load default language messages (as the base for all translations)
load_default_messages() {
    local default_file="${LANG_DIR}/${DEFAULT_LANG}/LC_MESSAGES/messages.sh"
    debug_log "Trying to load default language (${DEFAULT_LANG}) messages from: $default_file"
    debug_log "File exists? $([ -f "$default_file" ] && echo "Yes" || echo "No")"
    
    if [ -f "$default_file" ]; then
        if [ "$DEBUG" = true ]; then
            debug_log "File content preview:"
            head -n 20 "$default_file"
        fi
        
        # Backup any existing MSG_ variables to avoid conflicts
        local old_msg_vars=$(set | grep '^MSG_' | cut -d= -f1)
        for var in $old_msg_vars; do
            unset "$var"
        done
        
        source "$default_file"
        success_log "Loaded ${DEFAULT_LANG} messages as reference"
        
        # Get information about loaded messages
        debug_log "Example message values:"
        debug_log "Value of MSG_WELCOME: ${MSG_WELCOME}"
        debug_log "Value of MSG_ERROR: ${MSG_ERROR}"
        
        # Count MSG_ variables using a different approach
        msg_count=$(set | grep '^MSG_' | wc -l)
        debug_log "Found $msg_count message variables"
        
        return 0
    else
        error_log "Default language file not found at: $default_file"
        error_log "Please make sure the default language file exists."
        return 1
    fi
}

# Function to list available languages with enhanced output
list_languages() {
    local count=0
    echo "Available languages:"
    echo "------------------"
    
    if [ "$DEBUG" = true ]; then
        debug_log "Searching in subdirectories of $LANG_DIR"
        ls -la "$LANG_DIR"
    fi
    
    for lang_dir in "${LANG_DIR}"/*; do
        if [ "$DEBUG" = true ]; then
            debug_log "Checking directory: $lang_dir"
        fi
        
        if [ -d "$lang_dir" ]; then
            message_file="${lang_dir}/LC_MESSAGES/messages.sh"
            
            if [ "$DEBUG" = true ]; then
                debug_log "Looking for message file: $message_file (exists: $([ -f "$message_file" ] && echo "Yes" || echo "No"))"
            fi
            
            if [ -f "${message_file}" ]; then
                lang_code=$(basename "$lang_dir")
                # Get language name if possible
                lang_name=$(grep '# Language:' "$message_file" | sed 's/# Language: //' | head -1)
                
                if [ -z "$lang_name" ]; then
                    lang_name="$lang_code"
                fi
                
                # Check if this is the default language
                if [ "$lang_code" = "$DEFAULT_LANG" ]; then
                    echo -e "  - ${GREEN}$lang_code${NC} ($lang_name) [DEFAULT]"
                else
                    echo "  - $lang_code ($lang_name)"
                fi
                count=$((count + 1))
            fi
        fi
    done
    
    if [ $count -eq 0 ]; then
        warning_log "No language files found."
    else
        echo ""
        success_log "Total: $count language(s)"
    fi
}

# Function to create a new language template with better formatting
create_language() {
    local lang="$1"
    
    # Check if language code is valid
    if [ -z "$lang" ] || [ ${#lang} -ne 2 ]; then
        error_log "Invalid language code. Please use a 2-letter ISO language code (e.g., 'en', 'fr')."
        return 1
    fi
    
    # Check if the language already exists
    if [ -f "${LANG_DIR}/${lang}/LC_MESSAGES/messages.sh" ]; then
        warning_log "Language '$lang' already exists."
        read -p "Do you want to overwrite it? (y/N): " confirm
        if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
            warning_log "Operation cancelled by user."
            return 1
        fi
    fi
    
    # Load default language messages as reference
    if ! load_default_messages; then
        return 1
    fi
    
    # Create directory if it doesn't exist
    local lang_dir="${LANG_DIR}/${lang}/LC_MESSAGES"
    if [ ! -d "$lang_dir" ]; then
        mkdir -p "$lang_dir"
        success_log "Created directory: $lang_dir"
    fi
    
    # Create template file
    local template_file="${lang_dir}/messages.sh"
    
    # Get a list of all message variables
    # Use set command to get all variables and filter for MSG_ prefix
    local msg_vars=$(set | grep '^MSG_' | cut -d= -f1 | sort)
    
    # Get categories from the default language file
    local categories=$(grep "^# " "${LANG_DIR}/${DEFAULT_LANG}/LC_MESSAGES/messages.sh" | grep -v "^# Language:")
    
    # Create template content
    {
        echo "#!/bin/bash"
        echo "#"
        echo "# Language: $lang"
        echo "#"
        echo ""
        echo "# Define messages for $lang"
        echo "# Variables must start with MSG_ to be recognized by the system"
        echo ""
        
        # Get the current content of the default language file to preserve categories and comments
        local current_section=""
        while IFS= read -r line; do
            # If this is a category comment, add it to the template
            if [[ $line =~ ^#[[:space:]]+(.*)[[:space:]]*$ && ! "$line" =~ ^#[[:space:]]+Language: ]]; then
                current_section=$(echo "$line" | sed 's/^#[[:space:]]*//g')
                echo "$line"
            # If this is a variable definition, add it with its English value
            elif [[ $line =~ ^MSG_.*= ]]; then
                var_name=$(echo "$line" | cut -d= -f1)
                eng_value=${!var_name}
                echo "$var_name=\"$eng_value\""
            # If this is an empty line, preserve it for readability
            elif [[ -z "$line" ]]; then
                echo ""
            fi
        done < "${LANG_DIR}/${DEFAULT_LANG}/LC_MESSAGES/messages.sh"
    } > "$template_file"
    
    # Make file executable
    chmod +x "$template_file"
    
    success_log "Created language template: $template_file"
    echo "Please edit this file to add your translations."
}

# Function to update all language files with new strings from default language
update_languages() {
    # Load default language messages as reference
    if ! load_default_messages; then
        return 1
    fi
    
    echo "Updating language files..."
    
    # Get default language message keys and values
    default_keys=()
    default_values=()
    
    for var in $(set | grep '^MSG_' | cut -d= -f1 | sort); do
        default_keys+=("$var")
        default_values+=("${!var}")
    done
    
    # Get categories from the default language file for proper formatting
    local categories=$(grep "^# " "${LANG_DIR}/${DEFAULT_LANG}/LC_MESSAGES/messages.sh" | grep -v "^# Language:")
    
    # Count total languages updated
    local updated_count=0
    local skipped_count=0
    
    # Update each language file
    for lang_dir in "${LANG_DIR}"/*; do
        if [ -d "$lang_dir" ] && [ "$lang_dir" != "${LANG_DIR}/${DEFAULT_LANG}" ]; then
            lang_code=$(basename "$lang_dir")
            lang_file="${lang_dir}/LC_MESSAGES/messages.sh"
            
            if [ -f "$lang_file" ]; then
                echo "Updating $lang_code language file..."
                
                # Back up the original file
                cp "$lang_file" "${lang_file}.bak"
                success_log "Created backup: ${lang_file}.bak"
                
                # Clear previous message variables
                for var in $(set | grep '^MSG_' | cut -d= -f1); do
                    unset "$var"
                done
                
                # Load current translations
                source "$lang_file"
                
                # Create new file content with proper formatting
                local temp_file="${lang_file}.new"
                
                # Get the current content of the default language file to preserve categories and comments
                {
                    echo "#!/bin/bash"
                    echo "#"
                    echo "# Language: $lang_code"
                    echo "#"
                    echo ""
                    echo "# Define messages for $lang_code"
                    echo "# Variables must start with MSG_ to be recognized by the system"
                    echo ""
                    
                    local current_section=""
                    while IFS= read -r line; do
                        # If this is a category comment, add it to the template
                        if [[ $line =~ ^#[[:space:]]+(.*)[[:space:]]*$ && ! "$line" =~ ^#[[:space:]]+Language: ]]; then
                            current_section=$(echo "$line" | sed 's/^#[[:space:]]*//g')
                            echo "$line"
                        # If this is a variable definition, add it with its translation or English value
                        elif [[ $line =~ ^(MSG_[A-Za-z0-9_]+)= ]]; then
                            var_name="${BASH_REMATCH[1]}"
                            eng_value=${!var_name}
                            
                            if [ -n "${!var_name}" ]; then
                                # Use existing translation
                                echo "$var_name=\"${!var_name}\""
                            else
                                # Add new key with English value as comment
                                echo "$var_name=\"$eng_value\" # TODO: Translate this"
                            fi
                        # If this is an empty line, preserve it for readability
                        elif [[ -z "$line" ]]; then
                            echo ""
                        fi
                    done < "${LANG_DIR}/${DEFAULT_LANG}/LC_MESSAGES/messages.sh"
                } > "$temp_file"
                
                # Replace old file with new one
                mv "$temp_file" "$lang_file"
                chmod +x "$lang_file"
                success_log "Updated $lang_file"
                updated_count=$((updated_count + 1))
            else
                warning_log "Language file not found for $lang_code, skipping."
                skipped_count=$((skipped_count + 1))
            fi
        fi
    done
    
    echo "-----------------------------"
    echo "Update summary:"
    echo "  - Languages updated: $updated_count"
    echo "  - Languages skipped: $skipped_count"
    success_log "Language files updated successfully."
}

# Function to validate translation files for errors
check_translation_files() {
    echo "Validating translation files..."
    echo "-------------------------------"
    
    local error_count=0
    local warning_count=0
    
    # Check each language file
    for lang_dir in "${LANG_DIR}"/*; do
        if [ -d "$lang_dir" ]; then
            lang_code=$(basename "$lang_dir")
            lang_file="${lang_dir}/LC_MESSAGES/messages.sh"
            
            if [ -f "$lang_file" ]; then
                echo "Checking $lang_code language file..."
                
                # Check if file has execute permission
                if [ ! -x "$lang_file" ]; then
                    warning_log "$lang_file does not have execute permission."
                    chmod +x "$lang_file"
                    success_log "Fixed permissions for $lang_file"
                    warning_count=$((warning_count + 1))
                fi
                
                # Check for syntax errors in the shell script
                bash -n "$lang_file"
                if [ $? -ne 0 ]; then
                    error_log "Syntax error in $lang_file"
                    error_count=$((error_count + 1))
                    continue
                fi
                
                # Check for duplicate keys
                local duplicates=$(grep -oE "^MSG_[A-Za-z0-9_]+" "$lang_file" | sort | uniq -d)
                if [ -n "$duplicates" ]; then
                    warning_log "Duplicate keys found in $lang_file:"
                    echo "$duplicates"
                    warning_count=$((warning_count + 1))
                fi
                
                # Check for untranslated strings
                local untranslated=$(grep -c "# TODO: Translate this" "$lang_file")
                if [ $untranslated -gt 0 ]; then
                    warning_log "$lang_file has $untranslated untranslated strings."
                    warning_count=$((warning_count + 1))
                fi
                
                # Check for missing quotes
                local missing_quotes=$(grep -E '^MSG_[A-Za-z0-9_]+=[^"]' "$lang_file" | grep -v '=""')
                if [ -n "$missing_quotes" ]; then
                    error_log "Missing quotes in $lang_file:"
                    echo "$missing_quotes"
                    error_count=$((error_count + 1))
                fi
                
                # Success for this file
                if [ $error_count -eq 0 ] && [ $warning_count -eq 0 ]; then
                    success_log "$lang_file is valid."
                fi
            else
                error_log "Language file not found: $lang_file"
                error_count=$((error_count + 1))
            fi
        fi
    done
    
    echo "-----------------------------"
    echo "Validation summary:"
    echo "  - Errors: $error_count"
    echo "  - Warnings: $warning_count"
    
    if [ $error_count -eq 0 ] && [ $warning_count -eq 0 ]; then
        success_log "All language files are valid."
        return 0
    elif [ $error_count -eq 0 ]; then
        warning_log "Language files have warnings but no errors."
        return 0
    else
        error_log "Language files have errors that must be fixed."
        return 1
    fi
}

# Function to show translation status with enhanced output
show_status() {
    # Load default language messages as reference
    if ! load_default_messages; then
        return 1
    fi
    
    # Count total keys using set instead of env
    local total_keys=$(set | grep -c '^MSG_')
    debug_log "Total message keys found: $total_keys"
    
    if [ $total_keys -eq 0 ]; then
        error_log "No message keys found in default language file."
        return 1
    fi
    
    # Get all keys from default language
    local default_keys=$(set | grep '^MSG_' | cut -d= -f1 | sort)
    
    echo "Translation Status:"
    echo "------------------"
    
    # Track overall statistics
    local total_languages=0
    local fully_translated=0
    local partially_translated=0
    local not_translated=0
    
    for lang_dir in "${LANG_DIR}"/*; do
        if [ -d "$lang_dir" ]; then
            lang_code=$(basename "$lang_dir")
            lang_file="${lang_dir}/LC_MESSAGES/messages.sh"
            
            if [ -f "$lang_file" ]; then
                debug_log "Analyzing file: $lang_file"
                total_languages=$((total_languages + 1))
                
                # Count message keys in this language file
                local msg_count=$(grep -c "^MSG_.*=" "$lang_file")
                local translated_count=$(grep -c "^MSG_.*=" "$lang_file" | grep -v "# TODO: Translate this")
                local untranslated_count=$(grep -c "# TODO: Translate this" "$lang_file")
                local percentage=0
                
                if [ $total_keys -gt 0 ]; then
                    percentage=$((translated_count * 100 / total_keys))
                    
                    # Colorize output based on percentage
                    if [ $percentage -eq 100 ]; then
                        echo -e "  - ${GREEN}$lang_code: $translated_count/$total_keys ($percentage%) translated${NC}"
                        fully_translated=$((fully_translated + 1))
                    elif [ $percentage -ge 75 ]; then
                        echo -e "  - ${YELLOW}$lang_code: $translated_count/$total_keys ($percentage%) translated${NC}"
                        partially_translated=$((partially_translated + 1))
                    else
                        echo -e "  - ${RED}$lang_code: $translated_count/$total_keys ($percentage%) translated${NC}"
                        partially_translated=$((partially_translated + 1))
                    fi
                    
                    # Vérifier la présence de variables en minuscules et majuscules
                    local upper_case=$(grep -c "^MSG_[A-Z]" "$lang_file")
                    local lower_case=$(grep -c "^MSG_[a-z]" "$lang_file")
                    echo "      Upper case keys: $upper_case, Lower case keys: $lower_case"
                    
                    # Identifier les doublons éventuels
                    echo "      Checking for duplicate keys..."
                    local duplicates=$(grep -oE "^MSG_[A-Za-z0-9_]+" "$lang_file" | sort | uniq -d | wc -l)
                    if [ $duplicates -gt 0 ]; then
                        warning_log "      $duplicates duplicate keys found"
                        # Afficher les clés dupliquées
                        grep -oE "^MSG_[A-Za-z0-9_]+" "$lang_file" | sort | uniq -d
                    fi
                    
                    # Show untranslated strings count if any
                    if [ $untranslated_count -gt 0 ]; then
                        warning_log "      $untranslated_count untranslated strings"
                    fi
                    
                    # Check for extra keys (not in default language)
                    local extra_keys=$(comm -13 <(echo "$default_keys" | sort) <(grep -oE "^MSG_[A-Za-z0-9_]+" "$lang_file" | sort | uniq))
                    if [ -n "$extra_keys" ]; then
                        warning_log "      Extra keys not in default language:"
                        echo "$extra_keys" | sed 's/^/        - /'
                    fi
                else
                    error_log "  - $lang_code: No message keys found in default language"
                    not_translated=$((not_translated + 1))
                fi
            else
                error_log "  - $lang_code: No messages file found"
                not_translated=$((not_translated + 1))
            fi
        fi
    done
    
    echo ""
    echo "Status Summary:"
    echo "  - Total languages: $total_languages"
    echo "  - Fully translated (100%): $fully_translated"
    echo "  - Partially translated: $partially_translated"
    echo "  - Not translated: $not_translated"
    
    return 0
}

# Function to export translations to different formats
export_translations() {
    local format="$1"
    
    if [ -z "$format" ]; then
        error_log "Please specify an export format (json, po)"
        return 1
    fi
    
    # Create export directory if it doesn't exist
    local export_dir="${SCRIPT_DIR}/exports"
    if [ ! -d "$export_dir" ]; then
        mkdir -p "$export_dir"
        success_log "Created export directory: $export_dir"
    fi
    
    case "$format" in
        json)
            export_json "$export_dir"
            ;;
        po)
            export_po "$export_dir"
            ;;
        *)
            error_log "Unsupported export format: $format. Use 'json' or 'po'."
            return 1
            ;;
    esac
}

# Function to export to JSON format
export_json() {
    local export_dir="$1"
    
    echo "Exporting translations to JSON format..."
    
    # Get all available languages
    local languages=()
    for lang_dir in "${LANG_DIR}"/*; do
        if [ -d "$lang_dir" ] && [ -f "${lang_dir}/LC_MESSAGES/messages.sh" ]; then
            languages+=($(basename "$lang_dir"))
        fi
    done
    
    # Create JSON for each language
    for lang in "${languages[@]}"; do
        local lang_file="${LANG_DIR}/${lang}/LC_MESSAGES/messages.sh"
        local json_file="${export_dir}/${lang}.json"
        
        # Clear any existing MSG_ variables to avoid conflicts
        for var in $(set | grep '^MSG_' | cut -d= -f1); do
            unset "$var"
        done
        
        # Source the language file to get all definitions
        source "$lang_file"
        
        # Start JSON content
        echo "{" > "$json_file"
        
        # Add each message key to JSON
        local msg_vars=$(set | grep '^MSG_' | cut -d= -f1 | sort)
        local first=true
        for var in $msg_vars; do
            local value=${!var}
            # Escape quotes in the value
            value=$(echo "$value" | sed 's/"/\\"/g')
            
            if [ "$first" = true ]; then
                echo "  \"$var\": \"$value\"" >> "$json_file"
                first=false
            else
                echo "  ,\"$var\": \"$value\"" >> "$json_file"
            fi
        done
        
        # End JSON content
        echo "}" >> "$json_file"
        
        success_log "Created $json_file"
    done
    
    # Create a combined JSON with all languages
    local all_json="${export_dir}/all_translations.json"
    echo "{" > "$all_json"
    
    local first_lang=true
    for lang in "${languages[@]}"; do
        if [ "$first_lang" = true ]; then
            echo "  \"$lang\": $(cat "${export_dir}/${lang}.json")" >> "$all_json"
            first_lang=false
        else
            echo "  ,\"$lang\": $(cat "${export_dir}/${lang}.json")" >> "$all_json"
        fi
    done
    
    echo "}" >> "$all_json"
    
    success_log "Created combined translations file: $all_json"
    success_log "Translations successfully exported to JSON format in $export_dir"
}

# Function to export to PO format (gettext compatible)
export_po() {
    local export_dir="$1"
    
    echo "Exporting translations to PO format..."
    
    # Make sure default language exists
    if [ ! -f "${LANG_DIR}/${DEFAULT_LANG}/LC_MESSAGES/messages.sh" ]; then
        error_log "Default language file not found: ${LANG_DIR}/${DEFAULT_LANG}/LC_MESSAGES/messages.sh"
        return 1
    fi
    
    # Load default language as a reference
    for var in $(set | grep '^MSG_' | cut -d= -f1); do
        unset "$var"
    done
    source "${LANG_DIR}/${DEFAULT_LANG}/LC_MESSAGES/messages.sh"
    
    # Get default messages
    local default_keys=()
    local default_values=()
    for var in $(set | grep '^MSG_' | cut -d= -f1 | sort); do
        default_keys+=("$var")
        default_values+=("${!var}")
    done
    
    # Current date in PO format
    local po_date=$(date +"%Y-%m-%d %H:%M%z")
    
    # Get all available languages except default
    for lang_dir in "${LANG_DIR}"/*; do
        if [ -d "$lang_dir" ] && [ "$(basename "$lang_dir")" != "$DEFAULT_LANG" ]; then
            local lang=$(basename "$lang_dir")
            local lang_file="${lang_dir}/LC_MESSAGES/messages.sh"
            local po_file="${export_dir}/${lang}.po"
            
            if [ -f "$lang_file" ]; then
                # Clear any existing MSG_ variables to avoid conflicts
                for var in $(set | grep '^MSG_' | cut -d= -f1); do
                    unset "$var"
                done
                
                # Source the language file to get all definitions
                source "$lang_file"
                
                # Create PO header
                cat > "$po_file" << EOF
# Translation for Export Trakt 4 Letterboxd
# Copyright (C) $(date +"%Y") Export Trakt 4 Letterboxd
# This file is distributed under the same license as the Export Trakt 4 Letterboxd package.
#
msgid ""
msgstr ""
"Project-Id-Version: Export Trakt 4 Letterboxd 1.0\\n"
"Report-Msgid-Bugs-To: \\n"
"POT-Creation-Date: $po_date\\n"
"PO-Revision-Date: $po_date\\n"
"Last-Translator: Automatic export\\n"
"Language-Team: $lang\\n"
"Language: $lang\\n"
"MIME-Version: 1.0\\n"
"Content-Type: text/plain; charset=UTF-8\\n"
"Content-Transfer-Encoding: 8bit\\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\\n"

EOF
                
                # Add each message to PO file
                for i in "${!default_keys[@]}"; do
                    local key="${default_keys[$i]}"
                    local default_value="${default_values[$i]}"
                    local translated_value="${!key}"
                    
                    # If the translation exists, add it to the PO file
                    echo "#: ${key}" >> "$po_file"
                    echo "msgid \"$default_value\"" >> "$po_file"
                    echo "msgstr \"$translated_value\"" >> "$po_file"
                    echo "" >> "$po_file"
                done
                
                success_log "Created $po_file"
            fi
        fi
    done
    
    # Create a POT (template) file
    local pot_file="${export_dir}/template.pot"
    
    # Create POT header
    cat > "$pot_file" << EOF
# Translation template for Export Trakt 4 Letterboxd
# Copyright (C) $(date +"%Y") Export Trakt 4 Letterboxd
# This file is distributed under the same license as the Export Trakt 4 Letterboxd package.
#
msgid ""
msgstr ""
"Project-Id-Version: Export Trakt 4 Letterboxd 1.0\\n"
"Report-Msgid-Bugs-To: \\n"
"POT-Creation-Date: $po_date\\n"
"PO-Revision-Date: YEAR-MO-DA HO:MI+ZONE\\n"
"Last-Translator: FULL NAME <EMAIL@ADDRESS>\\n"
"Language-Team: LANGUAGE <LL@li.org>\\n"
"Language: \\n"
"MIME-Version: 1.0\\n"
"Content-Type: text/plain; charset=UTF-8\\n"
"Content-Transfer-Encoding: 8bit\\n"

EOF
    
    # Add each message to POT file
    for i in "${!default_keys[@]}"; do
        local key="${default_keys[$i]}"
        local default_value="${default_values[$i]}"
        
        echo "#: ${key}" >> "$pot_file"
        echo "msgid \"$default_value\"" >> "$pot_file"
        echo "msgstr \"\"" >> "$pot_file"
        echo "" >> "$pot_file"
    done
    
    success_log "Created POT template: $pot_file"
    success_log "Translations successfully exported to PO format in $export_dir"
}

# Function to import translations from external file
import_translations() {
    local import_file="$1"
    
    if [ -z "$import_file" ] || [ ! -f "$import_file" ]; then
        error_log "Please specify a valid import file"
        return 1
    fi
    
    echo "Importing translations from $import_file..."
    
    # Determine the file format based on extension
    local file_ext="${import_file##*.}"
    
    case "$file_ext" in
        po)
            import_po "$import_file"
            ;;
        json)
            import_json "$import_file"
            ;;
        *)
            error_log "Unsupported import format: $file_ext. Use '.po' or '.json' files."
            return 1
            ;;
    esac
}

# Function to import from PO format
import_po() {
    local import_file="$1"
    
    # Try to determine language from PO file
    local lang=$(grep '"Language:' "$import_file" | sed 's/.*"Language: \([^\\]*\).*/\1/' | head -1)
    
    if [ -z "$lang" ]; then
        error_log "Could not determine language from PO file"
        read -p "Please specify the language code (e.g., 'fr', 'de'): " lang
        
        if [ -z "$lang" ] || [ ${#lang} -ne 2 ]; then
            error_log "Invalid language code"
            return 1
        fi
    fi
    
    # Check if the language already exists
    local lang_file="${LANG_DIR}/${lang}/LC_MESSAGES/messages.sh"
    if [ ! -f "$lang_file" ]; then
        warning_log "Language file not found: $lang_file"
        read -p "Create a new language file? (y/N): " confirm
        if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
            create_language "$lang"
        else
            error_log "Import cancelled"
            return 1
        fi
    fi
    
    # Load default language to get keys
    if ! load_default_messages; then
        return 1
    fi
    
    # Get default messages mapping
    declare -A default_messages
    for var in $(set | grep '^MSG_' | cut -d= -f1); do
        default_messages["${!var}"]="$var"
    done
    
    # Parse PO file and update language file
    local temp_file="${lang_file}.new"
    cp "$lang_file" "$temp_file"
    
    # Parse each message block in the PO file
    local current_msgid=""
    local current_msgstr=""
    local in_msgid=false
    local in_msgstr=false
    
    while IFS= read -r line; do
        if [[ $line =~ ^msgid\ \"(.*)\"$ ]]; then
            current_msgid="${BASH_REMATCH[1]}"
            in_msgid=true
            in_msgstr=false
        elif [[ $line =~ ^\"(.*)\"$ ]] && [ "$in_msgid" = true ]; then
            current_msgid="$current_msgid${BASH_REMATCH[1]}"
        elif [[ $line =~ ^msgstr\ \"(.*)\"$ ]]; then
            current_msgstr="${BASH_REMATCH[1]}"
            in_msgid=false
            in_msgstr=true
        elif [[ $line =~ ^\"(.*)\"$ ]] && [ "$in_msgstr" = true ]; then
            current_msgstr="$current_msgstr${BASH_REMATCH[1]}"
        elif [[ -z "$line" ]] && [ -n "$current_msgid" ] && [ -n "$current_msgstr" ]; then
            # End of a message block, update the translation
            local var_name="${default_messages[$current_msgid]}"
            if [ -n "$var_name" ]; then
                # Escape quotes in the value
                current_msgstr=$(echo "$current_msgstr" | sed 's/"/\\"/g')
                
                # Update the variable in the language file
                sed -i.bak "s|^$var_name=.*|$var_name=\"$current_msgstr\"|" "$temp_file"
            fi
            
            current_msgid=""
            current_msgstr=""
            in_msgid=false
            in_msgstr=false
        fi
    done < "$import_file"
    
    # Handle the last message block
    if [ -n "$current_msgid" ] && [ -n "$current_msgstr" ]; then
        local var_name="${default_messages[$current_msgid]}"
        if [ -n "$var_name" ]; then
            # Escape quotes in the value
            current_msgstr=$(echo "$current_msgstr" | sed 's/"/\\"/g')
            
            # Update the variable in the language file
            sed -i.bak "s|^$var_name=.*|$var_name=\"$current_msgstr\"|" "$temp_file"
        fi
    fi
    
    # Replace the original file
    mv "$temp_file" "$lang_file"
    rm -f "${lang_file}.bak"
    chmod +x "$lang_file"
    
    success_log "Imported translations from $import_file to $lang_file"
}

# Function to import from JSON format
import_json() {
    local import_file="$1"
    
    # Try to determine if this is a single language or multiple languages
    local first_line=$(head -1 "$import_file")
    
    if [[ "$first_line" =~ \{\"[a-z]{2}\"\: ]]; then
        # This is a multi-language JSON file
        warning_log "Detected multi-language JSON file. Importing all languages..."
        
        # Extract each language section
        for lang in $(grep -o '"[a-z][a-z]":' "$import_file" | sed 's/"//g' | sed 's/://g'); do
            echo "Importing language: $lang"
            
            # Extract the language section to a temporary file
            local temp_json="/tmp/${lang}_import.json"
            # Use a combination of sed and awk to extract the language section
            sed -n "/\"$lang\":/,/^  \}/p" "$import_file" | sed '1s/^  "'"$lang"'": //' | sed '$s/  ,$//' > "$temp_json"
            
            # Import the language
            import_json_language "$lang" "$temp_json"
            rm -f "$temp_json"
        done
    else
        # This is a single language JSON file
        local lang=""
        
        # Try to determine language from filename
        if [[ "$import_file" =~ ([a-z]{2})\.json$ ]]; then
            lang="${BASH_REMATCH[1]}"
        else
            read -p "Please specify the language code for this JSON file (e.g., 'fr', 'de'): " lang
        fi
        
        if [ -z "$lang" ] || [ ${#lang} -ne 2 ]; then
            error_log "Invalid language code"
            return 1
        fi
        
        import_json_language "$lang" "$import_file"
    fi
}

# Helper function to import a single language from JSON
import_json_language() {
    local lang="$1"
    local json_file="$2"
    
    # Check if the language already exists
    local lang_file="${LANG_DIR}/${lang}/LC_MESSAGES/messages.sh"
    if [ ! -f "$lang_file" ]; then
        warning_log "Language file not found: $lang_file"
        read -p "Create a new language file? (y/N): " confirm
        if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
            create_language "$lang"
        else
            error_log "Import cancelled for language $lang"
            return 1
        fi
    fi
    
    # Create a temporary file for processing
    local temp_file="${lang_file}.new"
    cp "$lang_file" "$temp_file"
    
    # Parse each key-value pair in the JSON file
    grep -oP '"MSG_[^"]+"\s*:\s*"[^"]*"' "$json_file" | while read -r line; do
        local var_name=$(echo "$line" | grep -oP '"MSG_[^"]+"' | tr -d '"')
        local var_value=$(echo "$line" | grep -oP ':\s*"\K[^"]*')
        
        # Escape quotes in the value
        var_value=$(echo "$var_value" | sed 's/"/\\"/g')
        
        # Update the variable in the language file
        if grep -q "^$var_name=" "$temp_file"; then
            sed -i.bak "s|^$var_name=.*|$var_name=\"$var_value\"|" "$temp_file"
        else
            warning_log "Key $var_name not found in language file, skipping"
        fi
    done
    
    # Replace the original file
    mv "$temp_file" "$lang_file"
    rm -f "${lang_file}.bak"
    chmod +x "$lang_file"
    
    success_log "Imported translations from $json_file to $lang_file"
}

# Main script logic
case "$1" in
    list)
        list_languages
        ;;
    create)
        create_language "$2"
        ;;
    update)
        update_languages
        ;;
    status)
        show_status
        ;;
    check)
        check_translation_files
        ;;
    export)
        export_translations "$2"
        ;;
    import)
        import_translations "$2"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        if [ -z "$1" ]; then
            show_help
        else
            error_log "Unknown command: $1"
            echo "Use '$0 help' to see available commands"
            exit 1
        fi
        ;;
esac

exit 0 