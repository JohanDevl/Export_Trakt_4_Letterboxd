# Internationalization Guide

This document explains how the internationalization (i18n) system works in the Export Trakt 4 Letterboxd project and how to contribute to translation.

## Overview

The internationalization system allows the script to be displayed in different languages based on user preferences. The process is automated and selects the language based on:

1. The user's explicit configuration in the `.config.cfg` file
2. The operating system language if no configuration is specified
3. English as the default language if the system language is not supported

## Supported Languages

Currently, the following languages are supported:

- English (en) - Default language
- French (fr)
- Spanish (es)
- German (de)
- Italian (it)

## Configuration

To explicitly set the language, edit the `config/.config.cfg` file and set the `LANGUAGE` variable:

```bash
# Language for user interface (en, fr, es, de, it)
# Leave empty for automatic detection from the system
LANGUAGE="en"
```

To use the system language, simply leave this value empty:

```bash
LANGUAGE=""
```

## File Structure

The internationalization system is organized according to a standard structure:

```
Export_Trakt_4_Letterboxd/
├── lib/
│   ├── i18n.sh             # Main internationalization module
│   └── ...
├── locales/                # Directory containing translations
│   ├── en/                 # English
│   │   └── LC_MESSAGES/
│   │       └── messages.sh # English messages file
│   ├── fr/                 # French
│   │   └── LC_MESSAGES/
│   │       └── messages.sh
│   ├── es/                 # Spanish
│   │   └── LC_MESSAGES/
│   │       └── messages.sh
│   ├── it/                 # Italian
│   │   └── LC_MESSAGES/
│   │       └── messages.sh
│   └── de/                 # German
│       └── LC_MESSAGES/
│           └── messages.sh
├── manage_translations.sh  # Translation management utility
└── ...
```

## How It Works

1. At startup, the script initializes the i18n module
2. The module loads the language specified in the configuration file or detects the system language
3. It then loads the corresponding messages from the appropriate language file
4. When displaying text to the user, the script uses the `_()` function to get the translated text

## Translation Management Utility

A translation utility `manage_translations.sh` is provided to help manage language files. It allows you to:

- List available languages
- Create a template for a new language
- Update language files with new strings
- Display translation status for all languages

### Using the Utility

```bash
# Display help
./manage_translations.sh help

# List available languages
./manage_translations.sh list

# Create a template for a new language (ex: Italian)
./manage_translations.sh create it

# Update all language files with new/missing strings
./manage_translations.sh update

# Display translation status for all languages
./manage_translations.sh status
```

## Translation Contribution Guide

If you want to contribute to translating the application into a new language or improving an existing translation, follow these steps:

1. **For a new language:**

   - Run `./manage_translations.sh create xx` (where `xx` is the 2-letter language code)
   - Edit the generated file in `locales/xx/LC_MESSAGES/messages.sh`

2. **To update an existing translation:**
   - Run `./manage_translations.sh update` to add missing strings
   - Look for comments with `# TODO: Translate this` and translate those strings

### Translation Tips

- Keep special characters like `%s`, `%d`, etc. as they are used for variable insertion
- Respect case and punctuation when relevant
- Make sure the translated text has a similar meaning to the original text
- Test your translation by setting `LANGUAGE="xx"` in the configuration file

## Message File Format

Each message file is a bash script that declares message variables:

```bash
#!/bin/bash
#
# Language: en
#

# Define messages
# Variables must start with MSG_ to be recognized by the system

# General messages
MSG_HELLO="Hello"
MSG_WELCOME="Welcome to Export Trakt 4 Letterboxd"
# More translations...
```

## Adding New Translatable Strings

If you're developing new features that require adding new translatable strings:

1. First add the string to the English file (`locales/en/LC_MESSAGES/messages.sh`)
2. Use the `_()` function to reference the string in your code
3. Run `./manage_translations.sh update` to update all language files

## Debugging

If you encounter issues with translations:

1. Check that the language file exists and is properly formatted
2. Verify that the message key exists in the message file
3. If a translation is missing, the system will use the default English text

## Locale-Specific Date and Time Formats

In addition to text translation, the system also supports different date formats based on language. This allows dates to be displayed in a format familiar to users from each region.
