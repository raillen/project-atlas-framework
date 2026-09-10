#!/bin/sh
set -eu

REPOSITORY="${ATLAS_REPOSITORY:-raillen/project-atlas-framework}"
VERSION="${ATLAS_VERSION:-v0.4.2}"
INSTALL_DIR="${ATLAS_INSTALL_DIR:-${HOME}/.local/bin}"
ATLAS_HOME_VALUE="${ATLAS_HOME:-${HOME}/.atlas}"
BASE_URL="https://github.com/${REPOSITORY}/releases/download/${VERSION}"

case "$(uname -s)" in
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
  *) printf '%s\n' "Unsupported operating system. Use a release asset or build from source." >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) printf '%s\n' "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ASSET="atlas-${OS}-${ARCH}" ;;
esac

if ! command -v curl >/dev/null 2>&1; then
  printf '%s\n' "curl is required." >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  CHECKSUM_TOOL="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  CHECKSUM_TOOL="shasum -a 256"
else
  printf '%s\n' "sha256sum or shasum is required." >&2
  exit 1
fi

TEMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TEMP_DIR"
}
trap cleanup EXIT INT TERM

curl --fail --silent --show-error --location "${BASE_URL}/checksums.txt" -o "${TEMP_DIR}/checksums.txt"
curl --fail --silent --show-error --location "${BASE_URL}/${ASSET}" -o "${TEMP_DIR}/${ASSET}"

EXPECTED="$(awk -v asset="$ASSET" '$2 == asset {print $1}' "${TEMP_DIR}/checksums.txt")"
if [ -z "$EXPECTED" ]; then
  printf '%s\n' "No checksum found for ${ASSET}." >&2
  exit 1
fi

if [ "$CHECKSUM_TOOL" = "sha256sum" ]; then
  ACTUAL="$(sha256sum "${TEMP_DIR}/${ASSET}" | awk '{print $1}')"
else
  ACTUAL="$(shasum -a 256 "${TEMP_DIR}/${ASSET}" | awk '{print $1}')"
fi
if [ "$EXPECTED" != "$ACTUAL" ]; then
  printf '%s\n' "Checksum verification failed." >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
chmod 0755 "${TEMP_DIR}/${ASSET}"
TARGET="${INSTALL_DIR}/atlas"
STAGED="${TARGET}.tmp.$$"
cp "${TEMP_DIR}/${ASSET}" "$STAGED"
chmod 0755 "$STAGED"
mv "$STAGED" "$TARGET"

ATLAS_HOME="$ATLAS_HOME_VALUE" "$TARGET" setup >/dev/null
printf '%s\n' "Atlas ${VERSION} installed at ${TARGET}."

# Configure system PATH idempotently
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*)
    # Already in PATH
    ;;
  *)
    UPDATED_FILES=""
    if [ -f "${HOME}/.bashrc" ] && ! grep -qF "$INSTALL_DIR" "${HOME}/.bashrc"; then
      printf '\n# Added by Project Atlas\nexport PATH="%s:$PATH"\n' "$INSTALL_DIR" >> "${HOME}/.bashrc"
      UPDATED_FILES="${UPDATED_FILES} ~/.bashrc"
    fi

    if [ -f "${HOME}/.zshrc" ] && ! grep -qF "$INSTALL_DIR" "${HOME}/.zshrc"; then
      printf '\n# Added by Project Atlas\nexport PATH="%s:$PATH"\n' "$INSTALL_DIR" >> "${HOME}/.zshrc"
      UPDATED_FILES="${UPDATED_FILES} ~/.zshrc"
    fi

    if [ -f "${HOME}/.config/fish/config.fish" ] && ! grep -qF "$INSTALL_DIR" "${HOME}/.config/fish/config.fish"; then
      printf '\n# Added by Project Atlas\nfish_add_path "%s"\n' "$INSTALL_DIR" >> "${HOME}/.config/fish/config.fish"
      UPDATED_FILES="${UPDATED_FILES} ~/.config/fish/config.fish"
    fi

    if [ -f "${HOME}/.profile" ] && ! grep -qF "$INSTALL_DIR" "${HOME}/.profile"; then
      printf '\n# Added by Project Atlas\nexport PATH="%s:$PATH"\n' "$INSTALL_DIR" >> "${HOME}/.profile"
      UPDATED_FILES="${UPDATED_FILES} ~/.profile"
    fi

    if [ -n "$UPDATED_FILES" ]; then
      printf '%s\n' "Configured PATH in:${UPDATED_FILES}"
      printf '%s\n' "To start using atlas in this terminal session, run:"
      printf '%s\n' "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
    ;;
esac

printf '%s\n' "Run: atlas version"
