# Export Trakt 4 Letterboxd

[![GitHub release](https://img.shields.io/github/v/release/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/releases)
[![GitHub stars](https://img.shields.io/github/stars/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/stargazers)
[![GitHub issues](https://img.shields.io/github/issues/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues)
[![GitHub license](https://img.shields.io/github/license/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/blob/main/LICENSE)
[![Docker Build](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/docker-publish.yml)
[![Docker Package](https://img.shields.io/badge/Docker-ghcr.io-blue?logo=docker)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkgs/container/export_trakt_4_letterboxd)
[![GitHub package size](https://img.shields.io/github/repo-size/JohanDevl/Export_Trakt_4_Letterboxd?logo=docker&label=Image%20Size)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkgs/container/export_trakt_4_letterboxd)
[![Trakt.tv](https://img.shields.io/badge/Trakt.tv-ED1C24?logo=trakt&logoColor=white)](https://trakt.tv)
[![Letterboxd](https://img.shields.io/badge/Letterboxd-00D735?logo=letterboxd&logoColor=white)](https://letterboxd.com)

This project allows you to export your Trakt.tv data to a format compatible with Letterboxd.

## Quick Start

### Prerequisites

- A Trakt.tv account
- A Trakt.tv application (Client ID and Client Secret)
- jq and curl (for local installation)

### Local Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Run the installation script:

   ```bash
   ./install.sh
   ```

3. Configure Trakt authentication:

   ```bash
   ./setup_trakt.sh
   ```

4. Export your data:
   ```bash
   ./Export_Trakt_4_Letterboxd.sh [option]
   ```
   Options: `normal` (default), `initial`, or `complete`

### Docker Installation

Using Docker Compose:

```bash
# Clone the repository
git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd

# Start the container
docker compose up -d

# Configure Trakt authentication
docker compose exec trakt-export ./setup_trakt.sh

# Run the export script
docker compose exec trakt-export ./Export_Trakt_4_Letterboxd.sh
```

Or pull the pre-built image:

```bash
docker pull ghcr.io/johandevl/export_trakt_4_letterboxd:latest
```

## Features

- Export rated movies and TV shows
- Export watch history with dates and ratings
- Export watchlist items
- Automatic detection of rewatched movies
- Supports various export modes (normal, initial, complete)
- Modular code structure for better maintainability
- Automated exports with cron
- Docker support

## Project Structure

The codebase has been modularized for better maintenance and readability:

```
Export_Trakt_4_Letterboxd/
├── lib/                     # Library modules
│   ├── config.sh            # Configuration management
│   ├── utils.sh             # Utility functions and debugging
│   ├── trakt_api.sh         # API interaction functions
│   ├── data_processing.sh   # Data transformation functions
│   └── main.sh              # Main orchestration module
├── config/                  # Configuration files
├── logs/                    # Log output
├── backup/                  # Backup of API responses
├── TEMP/                    # Temporary processing files
├── copy/                    # Output CSV files
├── Export_Trakt_4_Letterboxd.sh # Main script (simplified)
├── setup_trakt.sh           # Authentication setup
└── install.sh               # Installation script
```

## Documentation

For more detailed information, please refer to the documentation in the `docs` folder:

- [Configuration and Basic Usage](docs/CONFIGURATION.md)
- [Docker Usage Guide](docs/DOCKER_USAGE.md)
- [Docker Testing](docs/DOCKER_TESTING.md)
- [GitHub Actions](docs/GITHUB_ACTIONS.md)
- [Automatic Version Tagging](docs/AUTO_TAGGING.md)

## Troubleshooting

If you encounter issues:

1. Check that your Trakt.tv profile is public
2. Verify your authentication configuration
3. Run `./setup_trakt.sh` again to refresh your tokens
4. Check logs in the `logs` directory for detailed error information

## Acknowledgements

This project is based on the original work by [u2pitchjami](https://github.com/u2pitchjami/Export_Trakt_4_Letterboxd). I would like to express my sincere gratitude to u2pitchjami for creating the initial version of this tool, which has been an invaluable foundation for this project.

The original repository can be found at: https://github.com/u2pitchjami/Export_Trakt_4_Letterboxd

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

The original work by u2pitchjami is also licensed under the MIT License. This fork maintains the same license to respect the original author's intentions.

## Authors

👤 **JohanDevl**

- Twitter: [@0xUta](https://twitter.com/0xUta)
- Github: [@JohanDevl](https://github.com/JohanDevl)
- LinkedIn: [@johan-devlaminck](https://linkedin.com/in/johan-devlaminck)
