#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Default variables
DEFAULT_DOCKER_USER="vivek6201"
DOCKER_USER="${1:-$DEFAULT_DOCKER_USER}"
TAG="${2:-latest}"

echo "=========================================================="
echo " Biolynq Production Build & Push Automation"
echo "=========================================================="
echo "Docker Hub User: $DOCKER_USER"
echo "Image Tag:       $TAG"
echo "=========================================================="

# Check if docker CLI is available
if ! command -v docker &> /dev/null; then
    echo "Error: docker command not found. Please install Docker."
    exit 1
fi

# Build Server Image
echo ""
echo "Building Production API Server Image (linux/amd64)..."
docker build \
  --platform linux/amd64 \
  --target prod \
  -f docker/Dockerfile.server \
  -t "$DOCKER_USER/biolynq-server:$TAG" .

# Build Worker Image
echo ""
echo "Building Production Background Worker Image (linux/amd64)..."
docker build \
  --platform linux/amd64 \
  --target prod \
  -f docker/Dockerfile.worker \
  -t "$DOCKER_USER/biolynq-worker:$TAG" .

echo ""
echo "=========================================================="
echo " Build successful! Images created locally:"
echo " - $DOCKER_USER/biolynq-server:$TAG"
echo " - $DOCKER_USER/biolynq-worker:$TAG"
echo "=========================================================="

if [ "$CI" != "true" ]; then
    echo "To push these images to Docker Hub, please make sure you are logged in."
    echo "Running: docker login"
    docker login
else
    echo "Running in CI environment, skipping interactive docker login."
fi

echo ""
echo "Pushing API Server Image..."
docker push "$DOCKER_USER/biolynq-server:$TAG"

echo ""
echo "Pushing Background Worker Image..."
docker push "$DOCKER_USER/biolynq-worker:$TAG"

echo ""
echo "=========================================================="
echo " Push successful! Images are now online on Docker Hub."
echo " You can deploy using: docker compose -f docker-compose.prod.yml up -d"
echo "=========================================================="
