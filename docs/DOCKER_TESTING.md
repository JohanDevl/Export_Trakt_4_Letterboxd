# Docker Testing Workflow

This document explains the automated testing workflow for Docker images in this project. The workflow is designed to verify that Docker images are functional and error-free before they are merged into the main or develop branches.

## Overview

The Docker testing workflow is defined in `.github/workflows/docker-test.yml` and is automatically triggered when a Pull Request is opened against the `main` or `develop` branches. This ensures that all Docker-related changes are thoroughly tested before being integrated.

## What Gets Tested

The workflow performs a comprehensive series of tests on the Docker image:

1. **Image Building**: Verifies that the Docker image can be built successfully from the Dockerfile.

2. **Structure Verification**: Checks that all required files, directories, and permissions are correctly set up:

   - Essential scripts (`Export_Trakt_4_Letterboxd.sh`, `setup_trakt.sh`, `docker-entrypoint.sh`)
   - Required directories (`config`, `logs`, `copy`, `brain_ops`, `backup`, `TEMP`)
   - Proper executable permissions

3. **Dependency Verification**: Ensures all required tools are installed:

   - `jq` for JSON processing
   - `curl` for API requests
   - `sed` for text manipulation

4. **Configuration Handling**: Tests the configuration file handling:

   - Presence of the example configuration file
   - Ability to create a working configuration file

5. **Cron Setup**: Verifies that the cron job setup functionality works correctly.

6. **Docker Compose**: Tests the Docker Compose configuration:

   - Validates the `docker-compose.yml` file
   - Ensures the container can be started and stopped with Docker Compose

7. **Volume Mounting**: Tests that volumes can be correctly mounted and accessed:
   - Creates test directories for all required volumes
   - Mounts these volumes to a test container
   - Verifies that the container can access the mounted volumes

## Test Steps

The workflow consists of the following steps:

1. **Checkout Repository**: Fetches the code from the repository.

2. **Set up Docker Buildx**: Configures Docker for building the image.

3. **Build Docker Image**: Builds the Docker image with the tag `trakt-export:test`.

4. **Verify Docker Image**: Runs a series of tests to verify the structure and dependencies of the image.

5. **Test Docker Compose**: Validates and tests the Docker Compose configuration.

6. **Test Docker Image with Mock Data**: Creates a test environment with mock data and verifies that the container can access and use this data.

7. **Summary**: Provides a summary of the test results.

## Test Output

The workflow provides detailed output for each test step, including:

- 🔍 Descriptive messages indicating what is being tested
- ✅ Success indicators for passed tests
- Detailed error messages for failed tests

If any test fails, the workflow will exit with a non-zero status code, causing the GitHub Actions check to fail. This prevents merging Pull Requests with broken Docker functionality.

## Running Tests Locally

You can run similar tests locally to verify your Docker image before creating a Pull Request:

```bash
# Build the Docker image
docker build -t trakt-export:test .

# Verify the image structure
docker run --rm trakt-export:test ls -la /app

# Test with Docker Compose
docker compose config
docker compose up -d
docker compose ps
docker compose down
```

## Troubleshooting

If the Docker testing workflow fails, check the GitHub Actions logs for detailed error messages. Common issues include:

1. **Missing Dependencies**: Ensure all required tools are installed in the Dockerfile.
2. **Permission Issues**: Check that scripts have the correct executable permissions.
3. **Configuration Problems**: Verify that the configuration file handling is working correctly.
4. **Volume Mounting Issues**: Ensure that the volume paths are correctly defined.

## Extending the Tests

To add more tests to the workflow, edit the `.github/workflows/docker-test.yml` file and add new steps or commands to the existing steps. Make sure to include descriptive messages and clear success/failure indicators.
