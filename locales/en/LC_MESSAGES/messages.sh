#!/bin/bash
#
# Language: en
#

# Define messages for en
# Variables must start with MSG_ to be recognized by the system

# General messages
MSG_HELLO="Hello"
MSG_WELCOME="Welcome to Export Trakt 4 Letterboxd"
MSG_GOODBYE="Goodbye"
MSG_ERROR="Error"
MSG_WARNING="Warning"
MSG_INFO="Information"
MSG_SUCCESS="Success"
MSG_FAILED="Failed"
MSG_DONE="Done"
MSG_ABORT="Abort"
MSG_CONTINUE="Continue"
MSG_YES="Yes"
MSG_NO="No"
MSG_CONFIRM="Confirm"
MSG_CANCEL="Cancel"
MSG_EXIT="Exit"
MSG_HELP="Help"
MSG_INVALID_OPTION="Invalid option"
MSG_PROCESSING="Processing"
MSG_PLEASE_WAIT="Please wait"

# Script messages
MSG_SCRIPT_STARTING="Starting Export Trakt 4 Letterboxd script"
MSG_SCRIPT_FINISHED="Export Trakt 4 Letterboxd script finished"
MSG_SCRIPT_INTERRUPTED="Script interrupted by user"
MSG_SCRIPT_ERROR="An error occurred while running the script"
MSG_SCRIPT_EXECUTION_START="Script execution started"
MSG_SCRIPT_EXECUTION_END="Script execution ended"
MSG_SCRIPT_OPTION="Script option"
MSG_NONE="none"
MSG_STARTING="Starting"
MSG_RUNNING_IN="running on"
MSG_LANGUAGE_SET="Language set to"
MSG_AUTO_DETECTED="auto-detected"
MSG_RUNNING_DOCKER="Running in Docker container"
MSG_SCRIPT_COMPLETE="Script execution completed successfully"

# Trakt API messages
MSG_TRAKT_AUTH_REQUIRED="Trakt authentication required"
MSG_TRAKT_AUTH_SUCCESS="Trakt authentication successful"
MSG_TRAKT_AUTH_FAILED="Trakt authentication failed"
MSG_TRAKT_API_ERROR="Error connecting to Trakt API"
MSG_TRAKT_API_RATE_LIMIT="Trakt API rate limit reached, waiting..."
MSG_API_REQUEST="API request"
MSG_API_RESPONSE="API response"
MSG_API_ERROR="API error"
MSG_API_RETRY="Retry"
MSG_API_LIMIT="API limit reached"
MSG_API_WAIT="Waiting before next request"
MSG_API_AUTH_REQUIRED="Authentication required"
MSG_API_AUTH_SUCCESS="Authentication successful"
MSG_API_AUTH_FAILURE="Authentication failed"

# Export messages
MSG_EXPORT_STARTING="Starting export process"
MSG_EXPORT_FINISHED="Export process completed"
MSG_EXPORT_FAILED="Export process failed"
MSG_EXPORT_NO_DATA="No data to export"
MSG_EXPORT_FILE_CREATED="Export file created: %s"
MSG_EXPORT_START="Starting export"
MSG_EXPORT_COMPLETE="Export completed"
MSG_EXPORT_PROCESSING="Processing export data"
MSG_EXPORT_FORMATTING="Formatting export data"
MSG_EXPORT_GENERATING="Generating export file"
MSG_EXPORT_SAVING="Saving export file"
MSG_EXPORT_SUMMARY="Export summary"

# User messages
MSG_USER_INPUT_REQUIRED="Please provide input"
MSG_USER_CONFIRM="Do you want to continue? (y/N)"
MSG_USER_INVALID_INPUT="Invalid input, please try again"
MSG_USER_INPUT="User input"
MSG_USER_SELECTION="User selection"
MSG_USER_CONFIRMATION="User confirmation"
MSG_USER_PROMPT="User prompt"
MSG_USER="User"

# Configuration messages
MSG_CONFIG_LOADED="Configuration loaded"
MSG_CONFIG_SAVED="Configuration saved"
MSG_CONFIG_ERROR="Error in configuration file"
MSG_CONFIG_NOT_FOUND="Configuration file not found"
MSG_CONFIG_CREATED="Configuration file created"
MSG_CONFIG_LOADING="Loading configuration"
MSG_CONFIG_SAVING="Saving configuration"
MSG_CONFIG_MISSING="Configuration missing"
MSG_CONFIG_UPDATED="Configuration updated"
MSG_CONFIG_DEFAULT="Default configuration"

# File operation messages
MSG_FILE_NOT_FOUND="File not found: %s"
MSG_FILE_CREATED="File created: %s"
MSG_FILE_DELETED="File deleted: %s"
MSG_FILE_UPDATED="File updated: %s"
MSG_FILE_PERMISSION_DENIED="Permission denied for file: %s"
MSG_FILE_READ_ERROR="File read error"
MSG_FILE_WRITE_ERROR="File write error"
MSG_DIRECTORY_CREATED="Directory created"
MSG_DIRECTORY_NOT_FOUND="Directory not found"
MSG_FILE_EXISTS="File exists"
MSG_FILE_EXISTS_NOT="File not found"
MSG_FILE_HAS_CONTENT="File has content"
MSG_FILE_IS_READABLE="File is readable"
MSG_FILE_IS_WRITABLE="File is writable"

# Translation messages
MSG_ERROR_MISSING_LANG_FILE="Error: Language file not found. Using English defaults."
MSG_TRANSLATION_LOADED="Translation loaded"
MSG_TRANSLATION_MISSING="Translation missing"
MSG_TRANSLATION_ERROR="Translation error"
MSG_TRANSLATION_UPDATED="Translation updated"

# System and directory messages
MSG_BACKUP_DIRECTORY="Backup directory"
MSG_BACKUP_DIRECTORY_EXISTS="Backup directory exists"
MSG_BACKUP_DIRECTORY_NOT_WRITABLE="WARNING: Backup directory is not writable. Check permissions."
MSG_BACKUP_DIRECTORY_WRITABLE="Backup directory is writable"
MSG_CHECKING_DEPENDENCIES="Checking required dependencies"
MSG_COPY_DIRECTORY="Copy directory"
MSG_CREATED_BACKUP_DIRECTORY="Created backup directory"
MSG_DIRECTORY_EXISTS="Directory exists"
MSG_DIRECTORY_PERMISSIONS="Directory permissions"
MSG_ENVIRONMENT_INFO="Environment information"
MSG_EXISTING_CSV_CHECK="Existing CSV file check"
MSG_LOG_DIRECTORY="Log directory"
MSG_MISSING_DEPENDENCIES="Some required dependencies are missing. Please install them before continuing."
MSG_NO_OPTION="No option provided, using default"
MSG_OS_TYPE="OS Type"
MSG_RETRIEVING_INFO="Retrieving information"
MSG_SCRIPT_DIRECTORY="Script directory"
MSG_WORKING_DIRECTORY="Working directory"

# API and token messages
MSG_API_KEY_CHECK="API key check"
MSG_API_KEY_FOUND="API key found"
MSG_API_KEY_NOT_FOUND="API key not found"
MSG_API_SECRET_CHECK="API secret check"
MSG_API_SECRET_FOUND="API secret found"
MSG_API_SECRET_NOT_FOUND="API secret not found"
MSG_ACCESS_TOKEN_CHECK="Access token check"
MSG_ACCESS_TOKEN_FOUND="Access token found"
MSG_ACCESS_TOKEN_NOT_FOUND="Access token not found"
MSG_REFRESH_TOKEN_CHECK="Refresh token check"
MSG_REFRESH_TOKEN_FOUND="Refresh token found"
MSG_REFRESH_TOKEN_NOT_FOUND="Refresh token not found"

# Deprecated keys (kept for backward compatibility)
# These will be removed in future versions
MSG_all_dependencies_installed="All required dependencies are installed." 