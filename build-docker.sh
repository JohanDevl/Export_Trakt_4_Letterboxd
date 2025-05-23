#!/bin/bash
# Script for building multi-architecture Docker images for Export_Trakt_4_Letterboxd

set -e

# Default settings
REGISTRY="docker.io"
REPOSITORY="johandevl/export-trakt-4-letterboxd"
DEFAULT_TAG="latest"
PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7"

# Usage info
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo "Build and push multi-architecture Docker images for Export_Trakt_4_Letterboxd"
    echo ""
    echo "Options:"
    echo "  -h, --help                 Show this help message"
    echo "  -t, --tag TAG              Specify Docker image tag (default: $DEFAULT_TAG)"
    echo "  -v, --version VERSION      Specify application version (default: derived from tag)"
    echo "  -p, --platforms PLATFORMS  Specify platforms to build for (default: $PLATFORMS)"
    echo "  -n, --no-push              Build but don't push images"
    echo "  -l, --local                Build for local platform only"
    echo "  --dry-run                  Show commands without executing"
    echo ""
    echo "Example:"
    echo "  $0 --tag v1.0.0 --version 1.0.0"
}

# Parse arguments
TAG=$DEFAULT_TAG
VERSION=""
NO_PUSH=false
LOCAL_ONLY=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -h|--help)
            show_help
            exit 0
            ;;
        -t|--tag)
            TAG="$2"
            shift
            shift
            ;;
        -v|--version)
            VERSION="$2"
            shift
            shift
            ;;
        -p|--platforms)
            PLATFORMS="$2"
            shift
            shift
            ;;
        -n|--no-push)
            NO_PUSH=true
            shift
            ;;
        -l|--local)
            LOCAL_ONLY=true
            PLATFORMS=""
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# If version not specified, derive from tag
if [[ -z "$VERSION" ]]; then
    # Remove 'v' prefix if present
    VERSION=$(echo "$TAG" | sed 's/^v//')
    if [[ "$VERSION" == "latest" ]]; then
        VERSION="dev"
    fi
fi

# Get current date in ISO 8601 format
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Get Git commit hash
VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "========================================================"
echo "Building Docker image for Export_Trakt_4_Letterboxd"
echo "========================================================"
echo "Image tag:    $REGISTRY/$REPOSITORY:$TAG"
echo "Version:      $VERSION"
echo "Build date:   $BUILD_DATE"
echo "Git commit:   $VCS_REF"
echo "Platforms:    ${PLATFORMS:-local platform only}"
echo "Push images:  $(if $NO_PUSH; then echo "No"; else echo "Yes"; fi)"
echo "========================================================"

# Check if Docker buildx is available
if ! docker buildx version &>/dev/null; then
    echo "Error: Docker buildx is not available. Please install it first."
    exit 1
fi

# Create a new builder instance if not in local-only mode
if [[ "$LOCAL_ONLY" == "false" ]]; then
    BUILDER_NAME="export-trakt-builder"
    
    # Check if builder exists, create if not
    if ! docker buildx inspect "$BUILDER_NAME" &>/dev/null; then
        echo "Creating new buildx builder: $BUILDER_NAME"
        if [[ "$DRY_RUN" == "false" ]]; then
            docker buildx create --name "$BUILDER_NAME" --use
        else
            echo "[DRY RUN] docker buildx create --name \"$BUILDER_NAME\" --use"
        fi
    else
        echo "Using existing buildx builder: $BUILDER_NAME"
        if [[ "$DRY_RUN" == "false" ]]; then
            docker buildx use "$BUILDER_NAME"
        else
            echo "[DRY RUN] docker buildx use \"$BUILDER_NAME\""
        fi
    fi
fi

# Build command components
BUILD_ARGS=(
    --build-arg "APP_VERSION=$VERSION"
    --build-arg "BUILD_DATE=$BUILD_DATE"
    --build-arg "VCS_REF=$VCS_REF"
)

TAG_ARGS=(
    -t "$REGISTRY/$REPOSITORY:$TAG"
)

# Add latest tag if this is a version tag
if [[ "$TAG" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    TAG_ARGS+=(-t "$REGISTRY/$REPOSITORY:latest")
fi

# Set platform args if not in local-only mode
PLATFORM_ARGS=()
if [[ "$LOCAL_ONLY" == "false" && -n "$PLATFORMS" ]]; then
    PLATFORM_ARGS=(--platform "$PLATFORMS")
fi

# Set output type based on push/no-push
OUTPUT_ARGS=()
if [[ "$NO_PUSH" == "true" ]]; then
    if [[ "$LOCAL_ONLY" == "true" ]]; then
        OUTPUT_ARGS=(--load)
    else
        OUTPUT_ARGS=(--output "type=image,push=false")
    fi
else
    OUTPUT_ARGS=(--push)
fi

# Build the image
echo "Building Docker image..."
BUILD_CMD=(docker buildx build "${BUILD_ARGS[@]}" "${TAG_ARGS[@]}" "${PLATFORM_ARGS[@]}" "${OUTPUT_ARGS[@]}" .)

if [[ "$DRY_RUN" == "true" ]]; then
    echo "[DRY RUN] ${BUILD_CMD[*]}"
else
    "${BUILD_CMD[@]}"
    
    echo "========================================================"
    if [[ "$NO_PUSH" == "true" ]]; then
        echo "Build completed. Images were not pushed."
    else
        echo "Build completed. Images were pushed to registry."
    fi
    echo "========================================================"
fi

exit 0 