#!/usr/bin/env bash

set -euo pipefail

script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repository_root="$(cd "${script_directory}/.." && pwd)"
repository_parent="$(cd "${repository_root}/.." && pwd)"
default_nubo_directory="${repository_parent}/nubo"
legacy_nubo_directory="${repository_parent}/nubo.git"

if [[ -n "${1:-}" ]]; then
  output_path="$1"
elif [[ -d "${default_nubo_directory}" ]]; then
  output_path="${default_nubo_directory}/goapi-linux"
elif [[ -d "${legacy_nubo_directory}" ]]; then
  output_path="${legacy_nubo_directory}/goapi-linux"
else
  output_path="${repository_root}/dist/goapi-linux"
fi

output_directory="$(dirname "${output_path}")"
mkdir -p "${output_directory}"
output_directory="$(cd "${output_directory}" && pwd)"
output_path="${output_directory}/$(basename "${output_path}")"

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

echo "Built Ubuntu 22.04-compatible GOAPI: ${output_path}"
if command -v file >/dev/null 2>&1; then
  file "${output_path}"
fi
