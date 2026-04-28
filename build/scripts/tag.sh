#!/bin/bash

set -euo pipefail

log() {
  echo "[tag] $*"
}

on_error() {
  local exit_code=$?
  log "ERROR at line $1 (exit code: $exit_code)"
}

trap 'on_error $LINENO' ERR

# Function to print usage
usage() {
  echo "Usage: $0 [--dry-run] [--stable] [--rc] [--canary] [--nightly]"
  exit 1
}

# Parse command-line arguments
DRY_RUN=false
RELEASE_TYPE="stable"
for arg in "$@"; do
  case $arg in
    --dry-run)
      DRY_RUN=true
      ;;
    --stable)
      RELEASE_TYPE="stable"
      ;;
    --rc)
      RELEASE_TYPE="rc"
      ;;
    --canary)
      RELEASE_TYPE="canary"
      ;;
    --nightly)
      RELEASE_TYPE="nightly"
      ;;
    *)
      usage
      ;;
  esac
done

log "Release type: $RELEASE_TYPE (dry-run: $DRY_RUN)"

# 1. Find last stable tag
LAST_STABLE=$(git tag | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1)
if [ -z "$LAST_STABLE" ]; then
  LAST_STABLE="v0.0.0"
fi
log "Last stable tag: $LAST_STABLE"

# 2. Extract MAJOR, MINOR, PATCH from last stable
IFS='.' read -r MAJOR MINOR PATCH <<< "${LAST_STABLE#v}"

# 3. Get commits since last stable
if [ "$LAST_STABLE" == "v0.0.0" ]; then
  COMMITS_SINCE_STABLE=$(git log --oneline --pretty=format:"- %s (%h)" --no-merges HEAD)
else
  COMMITS_SINCE_STABLE=$(git log --oneline --pretty=format:"- %s (%h)" --no-merges "$LAST_STABLE"..HEAD)
fi

# 4. Compute new base version
if echo "$COMMITS_SINCE_STABLE" | grep -qE "^- .+(\(.+\))?!:"; then
  MAJOR=$((MAJOR+1))
  MINOR=0
  PATCH=0
  INCREMENT="Major"
elif echo "$COMMITS_SINCE_STABLE" | grep -qE 'feat(\(.+\))?:'; then
  MINOR=$((MINOR+1))
  PATCH=0
  INCREMENT="Minor"
elif echo "$COMMITS_SINCE_STABLE" | grep -qE 'fix(\(.+\))?:'; then
  PATCH=$((PATCH+1))
  INCREMENT="Patch"
else
  INCREMENT="None"
fi

NEW_BASE_TAG="v$MAJOR.$MINOR.$PATCH"

# 5. For pre-releases, find last pre-release for computed base version
if [[ "$RELEASE_TYPE" =~ ^(rc|canary|nightly)$ ]]; then
  LAST_PRERELEASE=$(git tag | grep -E "^$NEW_BASE_TAG-$RELEASE_TYPE\\." | sort -V | tail -1 || true)
  if [ -n "$LAST_PRERELEASE" ]; then
    # Continue pre-release series
    if [ "$RELEASE_TYPE" == "rc" ]; then
      PRERELEASE_NUM=$(echo "$LAST_PRERELEASE" | sed 's/.*-rc\.\([0-9]*\)$/\1/')
      PRERELEASE_NUM=$((PRERELEASE_NUM+1))
      NEW_TAG="$NEW_BASE_TAG-rc.$PRERELEASE_NUM"
    elif [ "$RELEASE_TYPE" == "canary" ]; then
      PRERELEASE_NUM=$(echo "$LAST_PRERELEASE" | sed 's/.*-canary\.\([0-9]*\)$/\1/')
      PRERELEASE_NUM=$((PRERELEASE_NUM+1))
      NEW_TAG="$NEW_BASE_TAG-canary.$PRERELEASE_NUM"
    elif [ "$RELEASE_TYPE" == "nightly" ]; then
      DATE=$(date +%Y%m%d)
      NEW_TAG="$NEW_BASE_TAG-nightly.$DATE"
    fi
  else
    # Start new pre-release series
    if [ "$RELEASE_TYPE" == "rc" ]; then
      NEW_TAG="$NEW_BASE_TAG-rc.1"
    elif [ "$RELEASE_TYPE" == "canary" ]; then
      NEW_TAG="$NEW_BASE_TAG-canary.1"
    elif [ "$RELEASE_TYPE" == "nightly" ]; then
      DATE=$(date +%Y%m%d)
      NEW_TAG="$NEW_BASE_TAG-nightly.$DATE"
    fi
  fi
else
  # Stable release: always bump base version
  NEW_TAG="$NEW_BASE_TAG"
fi

# 6. Find previous tag for commit aggregation
if [[ "$RELEASE_TYPE" =~ ^(rc|canary|nightly)$ ]]; then
  LAST_PRERELEASE=$(git tag | grep -E "^$NEW_BASE_TAG-$RELEASE_TYPE\\." | sort -V | tail -1 || true)
  if [ -n "$LAST_PRERELEASE" ]; then
    PREV_TAG="$LAST_PRERELEASE"
  else
    PREV_TAG="$LAST_STABLE"
  fi
else
  PREV_TAG="$LAST_STABLE"
fi

if [ -z "$PREV_TAG" ]; then
  COMMITS_AGG=$(git log --oneline --pretty=format:"- %s (%h)" --no-merges HEAD)
else
  COMMITS_AGG=$(git log --oneline --pretty=format:"- %s (%h)" --no-merges "$PREV_TAG"..HEAD)
fi

if [ -z "$COMMITS_AGG" ]; then
  log "No commits to aggregate. Skipping tag creation."
  exit 0
fi


COMMITS=$(echo "$COMMITS_AGG" | grep -E "^- (feat|fix|.+(\(.+\))?!:)" || true)
if [ -z "$COMMITS" ]; then
  log "No feat/fix commits found. Skipping tag creation."
  exit 0
fi

ANNOTATED_MESSAGE="$INCREMENT $NEW_TAG

$COMMITS"

log "Tag: $NEW_TAG"
log "Message: \"$ANNOTATED_MESSAGE\""

if $DRY_RUN; then
  log "Dry run: tag creation and push skipped."
  exit 0
fi

git tag -a "$NEW_TAG" -m "$ANNOTATED_MESSAGE"

# Push the new tag to the remote
log "Pushing tag $NEW_TAG"
git push origin "$NEW_TAG"

# # For github action
# echo "new_tag=$(echo $NEW_TAG)" >> $GITHUB_OUTPUT
