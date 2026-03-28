#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist="${root}/dist"
mkdir -p "${dist}"

targets=(
  "windows/amd64/.exe"
  "windows/arm64/.exe"
  "linux/amd64/"
  "linux/arm64/"
  "darwin/amd64/"
  "darwin/arm64/"
)

for target in "${targets[@]}"; do
  IFS="/" read -r goos goarch ext <<< "${target}"
  export GOOS="${goos}"
  export GOARCH="${goarch}"
  export CGO_ENABLED=0
  out="${dist}/vterminal_${goos}_${goarch}${ext}"
  echo "Building ${out}"
  go build -trimpath -ldflags "-s -w" -o "${out}" ./cmd/vterminal
done