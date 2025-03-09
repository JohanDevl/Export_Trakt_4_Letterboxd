![Export_Trakt_4_Letterboxd](https://socialify.git.ci/u2pitchjami/Export_Trakt_4_Letterboxd/image?description=1&descriptionEditable=The%20purpose%20of%20this%20script%20is%20to%20export%20Trakt%20movies%20watchlist%20to%20csv%20file%20for%20manual%20Letterboxd%20import&font=Jost&language=1&logo=https%3A%2F%2Fgreen-berenice-35.tiiny.site%2Fimage2vector-3.svg&name=1&owner=1&pattern=Charlie%20Brown&stargazers=1&theme=Dark)

# Export Trakt 4 Letterboxd

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

2. Configure Trakt authentication:

   ```bash
   ./setup_trakt.sh
   ```

3. Export your data:
   ```bash
   ./Export_Trakt_4_Letterboxd.sh [option]
   ```
   Options: `normal` (default), `initial`, or `complet`

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
- Export watch history
- Export watchlist
- Automated exports with cron
- Docker support

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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Authors

👤 **u2pitchjami**

- Twitter: [@u2pitchjami](https://twitter.com/u2pitchjami)
- Github: [@u2pitchjami](https://github.com/u2pitchjami)
- LinkedIn: [@thierry-beugnet-a7761672](https://linkedin.com/in/thierry-beugnet-a7761672)
