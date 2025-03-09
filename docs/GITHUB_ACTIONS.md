# GitHub Actions Workflow Documentation

This document provides detailed information about the GitHub Actions workflow used in this project to build and publish Docker images to the GitHub Container Registry (ghcr.io).

## Overview

The workflow is defined in `.github/workflows/docker-publish.yml` and performs the following tasks:

1. Builds a Docker image from the project
2. Pushes the image to GitHub Container Registry (ghcr.io)
3. Signs the image using Cosign for security
4. Automatically tags the latest build as "latest" for easy reference

## Workflow Triggers

The workflow is triggered by:

- **Schedule**: Runs daily at 15:32 UTC (`cron: "32 15 * * *"`)
- **Push to main branch**: Any commits pushed to the `main` branch
- **Push to develop branch**: Any commits pushed to the `develop` branch
- **Version tags**: Any tags matching the pattern `v*.*.*` (e.g., `v1.0.0`)
- **Pull requests**: Any pull requests targeting the `main` branch

## Workflow Steps

The workflow consists of the following main steps:

1. **Checkout repository**: Fetches the code from the repository
2. **Install Cosign**: Sets up Cosign for image signing (except on PRs)
3. **Set up Docker Buildx**: Configures Docker for multi-platform builds
4. **Log into registry**: Authenticates with GitHub Container Registry (except on PRs)
5. **Extract Docker metadata**: Prepares tags and labels for the image
6. **Build and push Docker image**: Builds the image and pushes it to the registry (except on PRs)
7. **Sign the published Docker image**: Signs the image using Cosign (except on PRs)

## Testing the Workflow

### Local Testing with `act`

You can test the workflow locally using [act](https://github.com/nektos/act):

```bash
# Install act
# macOS
brew install act

# Linux
curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run the workflow for a push event
act push

# Run the workflow for a specific event
act workflow_dispatch
```

Note: Some features like Cosign signing might not work correctly in local testing.

### Testing on GitHub

To test the workflow on GitHub:

1. **Push to main branch**:

   ```bash
   git add .
   git commit -m "Test GitHub Actions workflow"
   git push origin main
   ```

2. **Create and push a version tag**:

   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. **Manual trigger**:
   - Go to your repository on GitHub
   - Navigate to "Actions" tab
   - Select the "Docker" workflow
   - Click "Run workflow" dropdown
   - Select the branch and click "Run workflow"

## Customizing the Workflow

### Changing the Schedule

To change when the workflow runs automatically, modify the `cron` expression in the `schedule` section:

```yaml
on:
  schedule:
    - cron: "32 15 * * *" # Current: 15:32 UTC daily
```

Common cron examples:

- `0 0 * * *`: Daily at midnight UTC
- `0 */6 * * *`: Every 6 hours
- `0 0 * * 0`: Weekly on Sunday at midnight UTC

### Image Tagging Strategy

The workflow uses a comprehensive tagging strategy:

1. **Semantic Versioning Tags** (for version tags like `v1.2.3`):

   - Full version: `v1.2.3`
   - Minor version: `v1.2`
   - Major version: `v1`

2. **Branch and PR Tags**:

   - Branch name (e.g., `main`, `develop`)
   - PR number (e.g., `pr-42`)

3. **Special Tags**:
   - The `latest` tag is automatically applied to:
     - Builds from the `main` branch
     - Builds triggered by version tags (e.g., `v1.2.3`)
   - The `develop` tag is automatically applied to:
     - Builds from the `develop` branch

This ensures that users can always access:

- The most recent stable version using the `latest` tag
- The most recent development version using the `develop` tag

### Changing the Registry

The workflow is configured to push to GitHub Container Registry (ghcr.io). To use a different registry:

1. Modify the `REGISTRY` environment variable:

   ```yaml
   env:
     REGISTRY: docker.io # For Docker Hub
   ```

2. Update the authentication step with appropriate credentials.

### Multi-Platform Builds

The workflow is set up for multi-platform builds using Docker Buildx. To specify platforms, add a `platforms` parameter to the build-and-push step:

```yaml
- name: Build and push Docker image
  uses: docker/build-push-action@v5.0.0
  with:
    context: .
    push: ${{ github.event_name != 'pull_request' }}
    tags: ${{ steps.meta.outputs.tags }}
    labels: ${{ steps.meta.outputs.labels }}
    platforms: linux/amd64,linux/arm64
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

## Troubleshooting

### Common Issues

1. **Authentication Failures**:

   - Ensure your GitHub token has the necessary permissions
   - Check that the repository has packages write permissions

2. **Build Failures**:

   - Check the Dockerfile for errors
   - Ensure all required files are included in the repository

3. **Signing Issues**:
   - Verify Cosign is installed correctly
   - Check that the identity token is available

### Viewing Workflow Logs

To view detailed logs:

1. Go to your repository on GitHub
2. Navigate to the "Actions" tab
3. Click on the specific workflow run
4. Expand the job and step that failed to see detailed logs

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker GitHub Action](https://github.com/docker/build-push-action)
- [Cosign Documentation](https://github.com/sigstore/cosign)
- [GitHub Container Registry Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
