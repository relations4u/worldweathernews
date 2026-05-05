#!/usr/bin/env bash
set -euo pipefail

# Helper for tagging and pushing a new release.
# Pushes a signed annotated tag matching v[0-9]*; the release.yml workflow
# picks it up and runs the build/sign/release pipeline.

if ! command -v gh >/dev/null; then
    echo "Error: gh CLI not installed."
    exit 1
fi

LATEST_TAG="$(git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')"
LATEST_VERSION="${LATEST_TAG#v}"
echo "Latest tag: $LATEST_TAG"

IFS='.' read -r MAJOR MINOR PATCH <<<"$LATEST_VERSION"
PATCH="${PATCH%%-*}" # strip pre-release suffix

echo
echo "Select bump type:"
echo "  1) patch (v${MAJOR}.${MINOR}.$((PATCH + 1)))"
echo "  2) minor (v${MAJOR}.$((MINOR + 1)).0)"
echo "  3) major (v$((MAJOR + 1)).0.0)"
echo "  4) custom"
read -rp "> " choice

case "$choice" in
    1) NEW_VERSION="${MAJOR}.${MINOR}.$((PATCH + 1))" ;;
    2) NEW_VERSION="${MAJOR}.$((MINOR + 1)).0" ;;
    3) NEW_VERSION="$((MAJOR + 1)).0.0" ;;
    4) read -rp "Custom version (without 'v' prefix): " NEW_VERSION ;;
    *)
        echo "Invalid"
        exit 1
        ;;
esac

NEW_TAG="v${NEW_VERSION}"

echo
echo "About to create tag: $NEW_TAG"
echo "Branch: $(git branch --show-current)"
echo "Commit: $(git rev-parse --short HEAD)"
echo

if [ -n "$(git status --porcelain)" ]; then
    echo "Working tree not clean. Commit or stash changes first."
    exit 1
fi

if [ "$(git branch --show-current)" != "main" ]; then
    read -rp "Not on main branch. Continue anyway? (y/N) " confirm
    [[ "$confirm" =~ ^[Yy]$ ]] || exit 1
fi

read -rp "Create and push $NEW_TAG? (y/N) " confirm
[[ "$confirm" =~ ^[Yy]$ ]] || exit 1

git tag -s "$NEW_TAG" -m "Release $NEW_TAG"
git push origin "$NEW_TAG"

echo
echo "Tag $NEW_TAG pushed."
echo "Watch the release pipeline:"
echo "  https://github.com/$(gh repo view --json nameWithOwner -q .nameWithOwner)/actions"
