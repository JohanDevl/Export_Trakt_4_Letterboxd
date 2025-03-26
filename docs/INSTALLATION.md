# Installation Guide

This guide provides detailed instructions for installing and configuring Export Trakt for Letterboxd on different platforms.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation Methods](#installation-methods)
  - [Using Pre-built Binaries](#using-pre-built-binaries)
  - [Building from Source](#building-from-source)
  - [Docker Installation](#docker-installation)
- [Configuration](#configuration)
- [Trakt.tv API Setup](#traktv-api-setup)
- [Troubleshooting](#troubleshooting)

## Prerequisites

Before installing Export Trakt for Letterboxd, ensure you have:

1. A Trakt.tv account
2. A Trakt.tv API application (Client ID and Client Secret)
3. Sufficient storage space for movie data and logs

## Installation Methods

### Using Pre-built Binaries

The easiest way to install Export Trakt for Letterboxd is to download the pre-built binary for your platform from the [releases page](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/releases).

1. Download the appropriate binary for your platform:

   - `export-trakt-linux-amd64.tar.gz` for Linux (64-bit)
   - `export-trakt-linux-arm64.tar.gz` for Linux ARM (64-bit)
   - `export-trakt-darwin-amd64.tar.gz` for macOS (Intel)
   - `export-trakt-darwin-arm64.tar.gz` for macOS (Apple Silicon)
   - `export-trakt-windows-amd64.zip` for Windows (64-bit)

2. Extract the archive:

   ```bash
   # Linux/macOS
   tar -xzf export-trakt-[platform].tar.gz

   # Windows
   # Use Windows Explorer or a tool like 7-Zip to extract the ZIP file
   ```

3. Move the binary to a location in your PATH (optional, for easier access):

   ```bash
   # Linux/macOS
   sudo mv export-trakt /usr/local/bin/

   # Windows
   # Move to a directory in your PATH or create a shortcut
   ```

4. Verify the installation:
   ```bash
   export-trakt --version
   ```

### Building from Source

To build from source, you'll need:

- Go 1.22 or later
- Git

Follow these steps:

1. Clone the repository:

   ```bash
   git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Build the application:

   ```bash
   go build -o export-trakt ./cmd/export_trakt
   ```

3. (Optional) Install the binary:

   ```bash
   # Linux/macOS
   sudo mv export-trakt /usr/local/bin/

   # Windows
   # Move to a directory in your PATH
   ```

### Docker Installation

Using Docker is recommended for cross-platform compatibility and isolated environments.

1. Pull the image from Docker Hub:

   ```bash
   docker pull johandevl/export-trakt-4-letterboxd:latest
   ```

2. Run the container:

   ```bash
   docker run -it --name trakt-export \
     -v $(pwd)/config:/app/config \
     -v $(pwd)/logs:/app/logs \
     -v $(pwd)/exports:/app/exports \
     johandevl/export-trakt-4-letterboxd:latest
   ```

3. Or use Docker Compose:

   Create a `docker-compose.yml` file:

   ```yaml
   version: "3"
   services:
     export-trakt:
       image: johandevl/export-trakt-4-letterboxd:latest
       container_name: trakt-export
       volumes:
         - ./config:/app/config
         - ./logs:/app/logs
         - ./exports:/app/exports
       restart: unless-stopped
   ```

   Then run:

   ```bash
   docker-compose up -d
   ```

## Configuration

Configuration is managed through a TOML file located at `config/config.toml`.

1. Create the configuration directory:

   ```bash
   mkdir -p config
   ```

2. Create a basic configuration file:

   ```bash
   cat > config/config.toml << EOF
   [trakt]
   client_id = "YOUR_CLIENT_ID"
   client_secret = "YOUR_CLIENT_SECRET"
   redirect_uri = "urn:ietf:wg:oauth:2.0:oob"

   [export]
   output_dir = "./exports"

   [logging]
   level = "info"
   file = "./logs/export.log"

   [i18n]
   language = "en"
   EOF
   ```

3. Replace `YOUR_CLIENT_ID` and `YOUR_CLIENT_SECRET` with your Trakt.tv API credentials.

## Trakt.tv API Setup

To use Export Trakt for Letterboxd, you need to register a Trakt.tv API application:

1. Go to [Trakt.tv API Applications](https://trakt.tv/oauth/applications)
2. Sign in to your Trakt.tv account
3. Click "New Application"
4. Fill in the application details:
   - Name: Export Trakt for Letterboxd
   - Redirect URI: `urn:ietf:wg:oauth:2.0:oob`
   - Description: Tool to export Trakt.tv data for Letterboxd import
5. Click "Save App"
6. Note the Client ID and Client Secret

The first time you run the application, it will guide you through the authentication process:

1. Run the application:

   ```bash
   export-trakt
   ```

2. The application will display a URL to visit
3. Visit the URL and authorize the application
4. Copy the authorization code
5. Paste the code into the application when prompted

## Troubleshooting

### Common Issues

1. **Authentication Failed**

   - Verify your Client ID and Client Secret are correct
   - Try re-authenticating by deleting the token file and restarting

2. **No Data Exported**

   - Ensure your Trakt.tv profile is public
   - Check that you have watched movies in your profile
   - Review the log file for detailed error messages

3. **Permission Denied Errors**
   - Check folder permissions for config, logs, and exports directories
   - If using Docker, verify that volume mounts are correctly configured

### Logs

Log files are stored in the `logs` directory. Check these files for detailed information about any issues:

```bash
cat logs/export.log
```

### Getting Help

If you continue to experience issues:

1. Check the [GitHub Issues](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues) to see if your problem has been reported
2. Create a new issue with details about your problem and the error messages
3. Include relevant sections from your log files (with sensitive information redacted)

## Updating

To update to the latest version:

1. For pre-built binaries, download the latest release
2. For Docker, pull the latest image:
   ```bash
   docker pull johandevl/export-trakt-4-letterboxd:latest
   ```
3. For source builds, pull the latest code and rebuild:
   ```bash
   git pull
   go build -o export-trakt ./cmd/export_trakt
   ```

Your existing configuration and tokens will be preserved across updates.
