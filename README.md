![Export_Trakt_4_Letterboxd](https://socialify.git.ci/u2pitchjami/Export_Trakt_4_Letterboxd/image?description=1&descriptionEditable=The%20purpose%20of%20this%20script%20is%20to%20export%20Trakt%20movies%20watchlist%20to%20csv%20file%20for%20manual%20Letterboxd%20import&font=Jost&language=1&logo=https%3A%2F%2Fgreen-berenice-35.tiiny.site%2Fimage2vector-3.svg&name=1&owner=1&pattern=Charlie%20Brown&stargazers=1&theme=Dark)

# Export Trakt 4 Letterboxd

This project allows you to export your Trakt.tv data to a format compatible with Letterboxd.

## Prerequisites

- A Trakt.tv account
- A Trakt.tv application (see below)
- jq (for JSON processing)
- curl (for API requests)

## Configuration

### 1. Create a Trakt.tv application

1. Log in to your Trakt.tv account
2. Go to https://trakt.tv/oauth/applications
3. Click on "New Application"
4. Fill in the information:
   - Name: Export Trakt 4 Letterboxd
   - Redirect URL: urn:ietf:wg:oauth:2.0:oob
   - Description: (optional)
5. Save the application
6. Note your Client ID and Client Secret

### 2. Set up the configuration file

Copy the example configuration file to create your own:

```bash
cp .config.cfg.example .config.cfg
```

You can edit the configuration file manually if you prefer, but it's recommended to use the setup script in the next step.

### 3. Authentication configuration

Run the configuration script:

```bash
./setup_trakt.sh
```

This script will guide you through the following steps:

1. Enter your Client ID and Client Secret
2. Obtain an authorization code
3. Generate access tokens

## Usage

### Export your data

```bash
./Export_Trakt_4_Letterboxd.sh [option]
```

Available options:

- `normal` (default): Exports rated movies, rated episodes, movie and TV show history, and watchlist
- `initial`: Exports only rated and watched movies
- `complet`: Exports all available data

### Result

The script generates a `letterboxd_import.csv` file that you can import on Letterboxd at the following address: https://letterboxd.com/import/

## Docker Usage

You can also run this application in a Docker container.

### Prerequisites for Docker

- Docker installed on your system
- Docker Compose (optional, but recommended)

### Using Docker Compose (recommended)

1. Clone the repository:

   ```bash
   git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Build and start the container:

   ```bash
   docker compose up -d
   ```

3. Enter the container:

   ```bash
   docker compose exec trakt-export bash
   ```

4. Configure Trakt authentication:

   ```bash
   ./setup_trakt.sh
   ```

5. Run the export script:
   ```bash
   ./Export_Trakt_4_Letterboxd.sh [option]
   ```

### Using Docker directly

1. Build the Docker image:

   ```bash
   docker build -t trakt-export .
   ```

2. Run the container:

   ```bash
   docker run -it --name trakt-export \
     -v $(pwd)/config:/app/config \
     -v $(pwd)/logs:/app/logs \
     -v $(pwd)/copy:/app/copy \
     -v $(pwd)/brain_ops:/app/brain_ops \
     -v $(pwd)/backup:/app/backup \
     trakt-export
   ```

3. Configure Trakt authentication:

   ```bash
   ./setup_trakt.sh
   ```

4. Run the export script:
   ```bash
   ./Export_Trakt_4_Letterboxd.sh [option]
   ```

### Docker Volumes

The Docker container uses the following volumes to persist data:

- `/app/config`: Contains the configuration file
- `/app/logs`: Contains log files
- `/app/copy`: Contains the exported Letterboxd CSV file
- `/app/brain_ops`: Contains additional export data
- `/app/backup`: Contains Trakt API backup data

### Automated Exports with Cron

You can configure the Docker container to automatically run the export script on a schedule using cron. To enable this feature, set the following environment variables:

- `CRON_SCHEDULE`: The cron schedule expression (e.g., `0 3 * * *` for daily at 3:00 AM)
- `EXPORT_OPTION`: The export option to use (`normal`, `initial`, or `complet`)

#### Example with Docker Compose:

```yaml
version: "3"

services:
  trakt-export:
    build: .
    container_name: trakt-export
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./copy:/app/copy
      - ./brain_ops:/app/brain_ops
      - ./backup:/app/backup
    environment:
      - TZ=Europe/Paris
      - CRON_SCHEDULE=0 3 * * * # Run daily at 3:00 AM
      - EXPORT_OPTION=normal # Use the normal export option
    stdin_open: true
    tty: true
```

#### Example with Docker Run:

```bash
docker run -it --name trakt-export \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/copy:/app/copy \
  -v $(pwd)/brain_ops:/app/brain_ops \
  -v $(pwd)/backup:/app/backup \
  -e CRON_SCHEDULE="0 3 * * *" \
  -e EXPORT_OPTION="normal" \
  trakt-export
```

#### Cron Job Logging and Monitoring

The cron job provides comprehensive logging to help you monitor the export process:

1. **Container Logs**:

   - User-friendly messages with emojis appear in the container logs
   - Start and completion notifications with timestamps
   - Progress indicators and success confirmations

   View these logs with:

   ```bash
   docker logs trakt-export
   ```

2. **Detailed Export Logs**:

   - Complete export details are saved to `/app/logs/cron_export.log`
   - Includes API responses, processing steps, and any warnings or errors
   - Timestamped entries for easier troubleshooting

   View these logs with:

   ```bash
   docker exec trakt-export cat /app/logs/cron_export.log
   ```

The cron job is configured to provide clear visual feedback about the export process, making it easy to confirm that your exports are running successfully.

### Docker Implementation Notes

The Docker implementation includes several optimizations:

1. **Modified `sed` commands**: The `sed` commands in the scripts have been adapted to work in Alpine Linux by removing the empty string argument (`''`) which is specific to macOS/BSD versions of `sed`.

2. **Configuration file handling**: The Docker setup uses a dedicated configuration directory (`/app/config`) with proper symlinks to ensure scripts can find and modify the configuration file.

3. **Permissions management**: The Docker entrypoint script ensures all files and directories have the correct permissions for read/write operations.

4. **Path handling**: All scripts have been updated to use absolute paths with the `SCRIPT_DIR` variable to ensure consistent file access regardless of the current working directory.

If you encounter any issues with the Docker implementation, please check the logs and ensure your configuration file is properly set up.

## Troubleshooting

### No data is exported

If the script runs without error but no data is exported:

1. Check that your Trakt.tv profile is public
2. Verify that you have correctly configured authentication
3. Run the configuration script again: `./setup_trakt.sh`

### Authentication errors

If you encounter authentication errors:

1. Check that your Client ID and Client Secret are correct
2. Get a new access token by running `./setup_trakt.sh`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

The MIT License is a permissive license that is short and to the point. It lets people do anything they want with your code as long as they provide attribution back to you and don't hold you liable.

## GitHub Actions and Container Registry

This project uses GitHub Actions to automatically build and publish Docker images to the GitHub Container Registry (ghcr.io). The workflow is defined in the `.github/workflows/docker-publish.yml` file.

For detailed documentation on the GitHub Actions workflow, see [GITHUB_ACTIONS.md](docs/GITHUB_ACTIONS.md).

### Automated Docker Testing

This project includes an automated testing workflow for Docker images. The workflow is triggered when a Pull Request is opened against the `main` or `develop` branches and performs a comprehensive series of tests to ensure that the Docker image is functional and error-free.

The tests include:

- Image building verification
- Structure and dependency checks
- Configuration handling tests
- Docker Compose validation
- Volume mounting tests

For detailed documentation on the Docker testing workflow, see [DOCKER_TESTING.md](docs/DOCKER_TESTING.md).

### Automatic Version Tagging

This project implements automatic version tagging when Pull Requests are merged into the main branch. The system:

- Automatically increments version numbers
- Creates Git tags for each merged PR
- Generates GitHub Releases with notes from PR descriptions
- Triggers the Docker image build workflow

For more information on how this works, see [AUTO_TAGGING.md](docs/AUTO_TAGGING.md).

### Testing the GitHub Actions Workflow

To test the GitHub Actions workflow locally or trigger it on GitHub:

#### Local Testing with `act`

You can test GitHub Actions workflows locally using [act](https://github.com/nektos/act), a tool that runs your GitHub Actions locally:

1. Install `act`:

   ```bash
   # macOS
   brew install act

   # Linux
   curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
   ```

2. Run the workflow locally:
   ```bash
   act push
   ```

#### Triggering the Workflow on GitHub

The workflow is configured to run automatically on:

- Schedule: Every day at 15:32 UTC
- Push to the `main` branch
- Push of version tags (e.g., `v1.0.0`)
- Pull requests to the `main` branch

To manually trigger the workflow:

1. Push to the main branch:

   ```bash
   git commit -m "Your commit message"
   git push origin main
   ```

2. Create and push a version tag:

   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. Or manually trigger the workflow from the GitHub Actions tab in your repository.

### Using the Published Docker Image

Once the workflow has run successfully, you can pull the published image from GitHub Container Registry. The image is automatically tagged with version numbers and special tags for easy reference:

```bash
# Pull the latest stable version (from main branch)
docker pull ghcr.io/johandevl/export_trakt_4_letterboxd:latest

# Pull the latest development version (from develop branch)
docker pull ghcr.io/johandevl/export_trakt_4_letterboxd:develop

# Or pull a specific version
docker pull ghcr.io/johandevl/export_trakt_4_letterboxd:v1.0.0
```

You can specify the image in your docker-compose.yml:

```yaml
version: "3"

services:
  trakt-export:
    # For stable production use:
    image: ghcr.io/johandevl/export_trakt_4_letterboxd:latest

    # For testing the latest development version:
    # image: ghcr.io/johandevl/export_trakt_4_letterboxd:develop

    container_name: trakt-export
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./copy:/app/copy
      - ./brain_ops:/app/brain_ops
      - ./backup:/app/backup
    environment:
      - TZ=Europe/Paris
      - CRON_SCHEDULE=0 3 * * * # Run daily at 3:00 AM
      - EXPORT_OPTION=normal # Use the normal export option
```

The tagging strategy ensures that:

- `latest` always points to the most recent stable build (main branch)
- `develop` always points to the most recent development build (develop branch)
- Specific versions are available for production stability
- You can also use major or minor version tags (e.g., `v1` or `v1.0`)

## Authors

👤 **u2pitchjami**

- Twitter: [@u2pitchjami](https://twitter.com/u2pitchjami)
- Github: [@u2pitchjami](https://github.com/u2pitchjami)
- LinkedIn: [@thierry-beugnet-a7761672](https://linkedin.com/in/thierry-beugnet-a7761672)

## Documentation

thanks to :

https://gist.github.com/kijart/4974b7b61bcec092dc3de3433e6e00e2

https://gist.github.com/darekkay/ff1c5aadf31588f11078
