#!/usr/bin/env bash

set -euo pipefail

script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repository_root="$(cd "${script_directory}/.." && pwd)"
default_runtime_directory="${repository_root}/dist/nubo-runtime"

if [[ -n "${1:-}" ]]; then
  output_path="$1"
else
  output_path="${default_runtime_directory}/bin/goapi"
fi

output_directory="$(dirname "${output_path}")"
mkdir -p "${output_directory}"
output_directory="$(cd "${output_directory}" && pwd)"
output_path="${output_directory}/$(basename "${output_path}")"
runtime_directory="$(dirname "${output_directory}")"
library_directory="${2:-${runtime_directory}/lib}"
license_directory="${3:-${runtime_directory}/licenses/sharp-libvips}"

mkdir -p "${library_directory}" "${license_directory}"
library_directory="$(cd "${library_directory}" && pwd)"
license_directory="$(cd "${license_directory}" && pwd)"

artifact_directory="$(mktemp -d)"
cleanup() {
  rm -rf "${artifact_directory}"
}
trap cleanup EXIT

docker buildx build \
  --platform linux/amd64 \
  --file "${repository_root}/build/ubuntu22/Dockerfile" \
  --target artifact \
  --output "type=local,dest=${artifact_directory}" \
  "${repository_root}"

install -m 0755 "${artifact_directory}/goapi-linux" "${output_path}"
cp -a "${artifact_directory}/lib/." "${library_directory}/"
cp -a "${artifact_directory}/licenses/sharp-libvips/." "${license_directory}/"

echo "Built Ubuntu 22.04-compatible GOAPI: ${output_path}"
echo "Bundled libvips compatibility and x86-64-v2 variants: ${library_directory}"
echo "Bundled libvips licenses: ${license_directory}"
if command -v file >/dev/null 2>&1; then
  file "${output_path}"
fi
