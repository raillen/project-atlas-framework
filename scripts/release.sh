#!/bin/sh
set -eu

VERSION="${VERSION:-0.4.1}"
OUTPUT="${OUTPUT:-dist}"

mkdir -p "$OUTPUT"

targets="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

for target in $targets; do
  os="${target%%/*}"
  arch="${target##*/}"
  name="atlas-${os}-${arch}"
  if [ "$os" = "windows" ]; then
    name="${name}.exe"
  fi
  echo "building $name"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "$OUTPUT/$name" ./cmd/atlas
done

(cd "$OUTPUT" && sha256sum atlas-* > checksums.txt)

cat > "$OUTPUT/release.json" <<EOF
{
  "version": "$VERSION",
  "artifacts": [
    "atlas-linux-amd64",
    "atlas-linux-arm64",
    "atlas-darwin-amd64",
    "atlas-darwin-arm64",
    "atlas-windows-amd64.exe",
    "atlas-windows-arm64.exe"
  ]
}
EOF

echo "release artifacts written to $OUTPUT"
