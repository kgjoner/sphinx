#! /bin/bash

set -e

IMAGE_NAME="sphinx"

echo "Starting integration script for image: $IMAGE_NAME"

#########################################################################################################
# SETUP
# -------------------------------------------------------------------------------------------------------
# 1. Parse flags
#########################################################################################################

RELEASE_KIND="stable"
DOCKER_REGISTRY=""
PLATFORM=""
for arg in "$@"; do
  case $arg in
    --registry=*)
      DOCKER_REGISTRY="${arg#*=}/"
      shift
      ;;
    --registry)
      DOCKER_REGISTRY="$2/"
      shift 2
      ;;
    --platform=*)
      PLATFORM="${arg#*=}"
      shift
      ;;
    --platform)
      PLATFORM="$2"
      shift 2
      ;;
    --rc)
      RELEASE_KIND="rc"
      shift
      ;;
    --canary)
      RELEASE_KIND="canary"
      shift
      ;;
    --nightly)
      RELEASE_KIND="nightly"
      shift
      ;;
    --stable)
      RELEASE_KIND="stable"
      shift
      ;;
    *)
      echo "Unknown argument: $arg"
      ;;
  esac
done

#########################################################################################################
# BUILD IMAGE
# -------------------------------------------------------------------------------------------------------
# 1. Get tag
# 2. Define build args
# 3. Build and push Docker image
#########################################################################################################

# 1. Get tag
# Fetch latest tag for release kind
if [ "$RELEASE_KIND" == "stable" ]; then
  LATEST_TAG=$(git tag | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1)
elif [ "$RELEASE_KIND" == "rc" ]; then
  LATEST_TAG=$(git tag | grep -E '^[v0-9]+\.[0-9]+\.[0-9]+-rc\.[0-9]+$' | sort -V | tail -1)
elif [ "$RELEASE_KIND" == "canary" ]; then
  LATEST_TAG=$(git tag | grep -E '^[v0-9]+\.[0-9]+\.[0-9]+-canary\.[0-9]+$' | sort -V | tail -1)
elif [ "$RELEASE_KIND" == "nightly" ]; then
  LATEST_TAG=$(git tag | grep -E '^[v0-9]+\.[0-9]+\.[0-9]+-nightly\.[0-9]+$' | sort -V | tail -1)
else
  LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
fi

if [ -z "$LATEST_TAG" ]; then
  echo "Error: No tag found for release kind '$RELEASE_KIND'. Exiting."
  exit 1
fi

# Extract major version (v1.2.3 → v1)
MAJOR_VERSION=$(echo "$LATEST_TAG" | sed -E 's/^(v[0-9]+)\..*$/\1/')

# 2. Define build args
DOCKER_TAGS="-t ${DOCKER_REGISTRY}$IMAGE_NAME:$LATEST_TAG"
if [ "$RELEASE_KIND" == "stable" ]; then
  DOCKER_TAGS="$DOCKER_TAGS -t ${DOCKER_REGISTRY}$IMAGE_NAME:$MAJOR_VERSION -t ${DOCKER_REGISTRY}$IMAGE_NAME:latest"
elif [ "$RELEASE_KIND" == "rc" ]; then
  DOCKER_TAGS="$DOCKER_TAGS -t ${DOCKER_REGISTRY}$IMAGE_NAME:rc"
elif [ "$RELEASE_KIND" == "canary" ]; then
  DOCKER_TAGS="$DOCKER_TAGS -t ${DOCKER_REGISTRY}$IMAGE_NAME:canary"
elif [ "$RELEASE_KIND" == "nightly" ]; then
  DOCKER_TAGS="$DOCKER_TAGS -t ${DOCKER_REGISTRY}$IMAGE_NAME:nightly"
fi

# 3. Build and push Docker image
eval "$(ssh-agent)"
ssh-add ~/.ssh/id_ed25519

echo "Building and pushing Docker image with version: $LATEST_TAG and kind: $RELEASE_KIND"

PLATFORM_ARG=""
if [ -n "$PLATFORM" ]; then
  PLATFORM_ARG="--platform $PLATFORM"
fi

PUSH_ARG=""
if [ -n "$DOCKER_REGISTRY" ]; then
  PUSH_ARG="--push"
fi

docker build -f Dockerfile.release ../ \
  --build-arg APP_VERSION="$LATEST_TAG" \
  $DOCKER_TAGS \
  --ssh default=$SSH_AUTH_SOCK \
  $PLATFORM_ARG \
  $PUSH_ARG

#########################################################################################################
# PUSH HELM CHART
# -------------------------------------------------------------------------------------------------------
# 1. Prepare Helm chart
# 2. Package it
# 3. Push to OCI registry
# 4. Clean up
#########################################################################################################

if [ -z "$DOCKER_REGISTRY" ]; then
  echo "No Docker registry specified, skipping Helm chart push."
  exit 0
fi

echo "Packaging Helm chart..."

# 1. Prepare Helm chart
# Replace app version in Helm chart
sed -i -E "s/appVersion:.+/appVersion: \"$LATEST_TAG\"/g" helm/Chart.yaml

# Copy migration files
echo "Copying migration files"
mkdir -p helm/migrations
cp internal/repositories/base/migrations/*.sql helm/migrations/

# 2. Package it
# Create temporary directory for chart packaging
TEMP_DIR=$(mktemp -d)

helm package helm/ --destination "$TEMP_DIR"

# 3. Push to OCI registry
CHART_FILE=$(ls "$TEMP_DIR"/*.tgz | head -n1)

if [[ -f "$CHART_FILE" ]]; then
  echo "Pushing Helm chart to OCI registry..."
  helm push "$CHART_FILE" oci://${DOCKER_REGISTRY}charts
  echo "✅ Helm chart pushed successfully!"
else
  echo "❌ Chart packaging failed - no .tgz file found"
  exit 1
fi

# 4. Clean up
rm -rf "$TEMP_DIR"

echo "🚀 Release completed!"
echo "Docker Images:"
echo "  ${DOCKER_REGISTRY}$IMAGE_NAME:$LATEST_TAG"
echo "  ${DOCKER_REGISTRY}$IMAGE_NAME:$MAJOR_VERSION"
echo "  ${DOCKER_REGISTRY}$IMAGE_NAME:latest"
echo "Helm Chart:"
echo "  oci://${DOCKER_REGISTRY}charts/$IMAGE_NAME"