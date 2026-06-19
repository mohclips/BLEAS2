#!/usr/bin/env bash
# One-shot installer for the duckdb CLI binary. No sudo required — drops
# it next to this script. run.sh will pick it up automatically.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="$HERE/duckdb"

if [[ -x "$TARGET" ]]; then
    echo "duckdb already installed at $TARGET"
    "$TARGET" --version
    exit 0
fi

case "$(uname -s)-$(uname -m)" in
    Linux-x86_64)   ASSET=duckdb_cli-linux-amd64.zip ;;
    Linux-aarch64)  ASSET=duckdb_cli-linux-arm64.zip ;;
    Darwin-x86_64)  ASSET=duckdb_cli-osx-universal.zip ;;
    Darwin-arm64)   ASSET=duckdb_cli-osx-universal.zip ;;
    *) echo "unsupported platform: $(uname -s)-$(uname -m)" >&2; exit 1 ;;
esac

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "Downloading $ASSET..."
curl -fsSL "https://github.com/duckdb/duckdb/releases/latest/download/$ASSET" -o "$TMP/duckdb.zip"
unzip -q "$TMP/duckdb.zip" -d "$TMP"
mv "$TMP/duckdb" "$TARGET"
chmod +x "$TARGET"

echo "Installed: $TARGET"
"$TARGET" --version
