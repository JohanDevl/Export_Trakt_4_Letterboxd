# Migration Guide: From Bash to Go

This guide helps users migrate from the original Bash implementation to the new Go implementation of Export_Trakt_4_Letterboxd.

## Key Differences

| Feature                  | Bash Version            | Go Version                                          |
| ------------------------ | ----------------------- | --------------------------------------------------- |
| **Performance**          | Good for small datasets | Significantly faster, especially for large datasets |
| **Error Handling**       | Basic                   | Comprehensive with detailed logging                 |
| **Configuration**        | .env file               | TOML-based configuration                            |
| **Internationalization** | None                    | Full support for multiple languages                 |
| **Extensibility**        | Modular scripts         | Modular packages with clear interfaces              |
| **Testing**              | Bats test suite         | Go native testing with high coverage                |
| **Dependencies**         | jq, curl, bash          | Standalone binary                                   |

## Migration Steps

### 1. Installation

#### Bash Version

```bash
git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd
./install.sh
```

#### Go Version

```bash
# Option 1: Download pre-built binary
curl -L https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/releases/latest/download/export_trakt_linux_amd64 -o export_trakt
chmod +x export_trakt

# Option 2: Build from source
git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd
go build -o export_trakt ./cmd/export_trakt
```

### 2. Configuration

#### Bash Version

Configuration was stored in a `.env` file with these key parameters:

```
CLIENT_ID=your_trakt_client_id
CLIENT_SECRET=your_trakt_client_secret
ACCESS_TOKEN=your_access_token
REFRESH_TOKEN=your_refresh_token
OUTPUT_DIR=./copy
```

#### Go Version

Configuration is stored in a TOML file (`config/config.toml`):

```toml
# Trakt.tv API Configuration
[trakt]
client_id = "your_trakt_client_id"
client_secret = "your_trakt_client_secret"
access_token = "your_access_token"
api_base_url = "https://api.trakt.tv"

# Letterboxd Export Configuration
[letterboxd]
export_dir = "exports"

# Export Settings
[export]
format = "csv"
date_format = "2006-01-02"

# Logging Configuration
[logging]
level = "info"
file = "logs/export.log"

# Internationalization Settings
[i18n]
default_language = "en"
language = "en"
locales_dir = "locales"
```

### 3. Running the Application

#### Bash Version

```bash
./Export_Trakt_4_Letterboxd.sh [option]
# Options: normal (default), initial, complete
```

#### Go Version

```bash
./export_trakt --config config/config.toml
```

### 4. Docker Usage

#### Bash Version

```bash
docker run -it --name trakt-export \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/copy:/app/copy \
  -v $(pwd)/backup:/app/backup \
  johandevl/export-trakt-4-letterboxd:latest
```

#### Go Version

```bash
docker run -it --name trakt-export \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/exports:/app/exports \
  ghcr.io/johandevl/export_trakt_4_letterboxd:latest
```

## Breaking Changes

### 1. Configuration Format

The Go version uses TOML format instead of an environment file. You'll need to create a new config file following the format above.

### 2. Command-line Options

The Go version uses flags instead of positional arguments. Instead of different execution modes, all options are defined in the config file.

### 3. Output Directory Structure

The Go version uses a more organized output directory structure, with exports going to the configured export directory by default.

### 4. Docker Image Name

The Go version uses a different Docker image name and is hosted on GitHub Container Registry instead of Docker Hub.

## New Features in Go Version

### 1. Improved Logging

The Go version includes a more comprehensive logging system with different log levels (debug, info, warn, error).

```bash
# View logs to diagnose issues
cat logs/export.log
```

### 2. Internationalization

The Go version supports multiple languages. To change the language, update the config file:

```toml
[i18n]
default_language = "en"
language = "fr" # Change to your preferred language
locales_dir = "locales"
```

### 3. Error Recovery

The Go version includes better error handling and retry mechanisms for API calls, making it more resilient.

### 4. Performance Improvements

The Go version is significantly faster and uses less memory, especially for large export operations.

## Migration FAQ

### Q: Can I use my existing Trakt.tv authentication?

**A:** Yes, you can copy your `CLIENT_ID`, `CLIENT_SECRET`, and `ACCESS_TOKEN` from the `.env` file to the new TOML configuration file.

### Q: Are the export files compatible?

**A:** Yes, both versions produce Letterboxd-compatible CSV files. The Go version may include additional metadata fields.

### Q: Do I need to install Go to use the Go version?

**A:** No, you can download the pre-built binary or use the Docker image, which doesn't require a Go installation.

### Q: Can I run both versions side by side?

**A:** Yes, they can coexist in different directories without conflict.

### Q: How do I migrate my customizations?

**A:** The Go version is designed to be more configurable through the TOML file. If you had custom scripts or modifications in the Bash version, review the API documentation to see how to achieve the same in the Go version.

## Getting Help

If you encounter issues during migration:

1. Check the detailed logs in the `logs` directory
2. Review the configuration file for errors
3. Open an issue on GitHub with details about your problem
4. Check the [API Documentation](API.md) for detailed information about the Go implementation
