#!/usr/bin/env bash
# Start the local container runtime when needed, then use the official Linux build.
set -euo pipefail

script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ "$(uname -s)" != Darwin ]]; then
  echo "This helper is for macOS. On Linux use ./scripts/build-ubuntu22.sh." >&2
  exit 1
fi
if ! command -v docker >/dev/null 2>&1; then
  echo "Docker CLI is missing. Install Docker and Buildx, and a runtime such as Colima; see README.md." >&2
  exit 1
fi
if ! docker info >/dev/null 2>&1; then
  if ! command -v colima >/dev/null 2>&1; then
    echo "Docker is not running. Start Docker Desktop or install and start Colima; see README.md." >&2
    exit 1
  fi
  echo "Docker is unavailable; starting Colima using its saved configuration..."
  colima start
  if ! docker info >/dev/null 2>&1; then
    echo "Docker is still unavailable. Check DOCKER_HOST and the selected Docker context (docker context show)." >&2
    exit 1
  fi
fi
if ! docker buildx version >/dev/null 2>&1; then
  echo "Docker Buildx is missing or unavailable. See the macOS build instructions in README.md." >&2
  exit 1
fi

exec "${script_directory}/build-ubuntu22.sh" "$@"
