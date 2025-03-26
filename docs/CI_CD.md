# CI/CD Pipeline Documentation

This project uses GitHub Actions for continuous integration and continuous deployment. The CI/CD pipeline automates testing, building, and deployment of the application.

## Pipeline Components

### 1. Go Tests Workflow

The tests workflow runs all unit and integration tests for the Go application.

**File**: `.github/workflows/go-tests.yml`

**Triggered by**:

- Push to main, develop, and feature/\* branches
- Pull requests to main and develop branches

**Steps**:

1. Check out the code
2. Set up Go environment
3. Install dependencies
4. Run all tests with coverage tracking
5. Generate a coverage report
6. Upload the coverage report as an artifact
7. Verify the code coverage meets minimum threshold (70%)

**Usage**:

```bash
# To run tests locally with coverage reporting
go test -v ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### 2. Go Build and Docker Publish Workflow

This workflow builds the Go application and creates a Docker image.

**File**: `.github/workflows/go-build.yml`

**Triggered by**:

- Push to main and develop branches
- Push of version tags (v\*)
- Pull requests to main and develop branches

**Jobs**:

1. **Build**: Compiles the Go application

   - Produces executable binary
   - Uploads the binary as an artifact

2. **Docker** (runs only on push, not PR):
   - Builds a Docker image using Dockerfile.go
   - Tags the image according to branch/tag
   - Pushes to GitHub Container Registry

**Docker Image Tags**:

- `latest` - Always points to the most recent build from main
- `vX.Y.Z` - For version releases
- `vX.Y` - Major.Minor version
- `develop` - For builds from the develop branch
- `sha-XXXXXXX` - Git commit SHA

## Docker Images

### Go Application (Dockerfile.go)

The Go application uses a minimal Alpine-based image for runtime.

- Base Image: `alpine:3.18`
- Image URL: `ghcr.io/johandevl/export_trakt_4_letterboxd`

**Volumes**:

- `/app/config` - Configuration files
- `/app/logs` - Log files
- `/app/exports` - Export output files

**Customization**:
The application can be configured by mounting a custom config file:

```bash
docker run -v /path/to/config.toml:/app/config/config.toml ghcr.io/johandevl/export_trakt_4_letterboxd
```

## Quality Standards

The CI pipeline enforces the following quality standards:

1. **Test Coverage**: Minimum 70% code coverage required
2. **Passing Tests**: All tests must pass
3. **Build Verification**: Application must build successfully

## Troubleshooting

### Common Issues

1. **Failed Tests**:

   - Check the test logs to identify which tests failed
   - Run tests locally to debug the issues

2. **Coverage Below Threshold**:

   - Add more tests to cover untested code
   - Run coverage report locally to identify uncovered areas

3. **Docker Build Failures**:
   - Verify the Dockerfile.go is valid
   - Check if all required files are available in the repository

### Viewing Artifacts

1. Go to the GitHub Actions tab in the repository
2. Select the workflow run
3. Scroll down to the "Artifacts" section
4. Download the artifact (coverage report or binary)
