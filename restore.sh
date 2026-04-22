#!/usr/bin/env bash
# Restores all dotly-managed symlinks under a given directory back to regular files.
# Usage: ./dotly-restore-all.sh <directory>

set -euo pipefail

DOTLY="${HOME}/.local/share/dotly"
TARGET="${1:-}"

if [[ -z "$TARGET" ]]; then
    echo "usage: $0 <directory>" >&2
    exit 1
fi

if [[ ! -d "$TARGET" ]]; then
    echo "error: $TARGET is not a directory" >&2
    exit 1
fi

restored=0
skipped=0

while IFS= read -r -d '' link; do
    target=$(readlink "$link")

    # Only touch symlinks pointing into DOTLY.
    case "$target" in
        "$DOTLY"/*) ;;
        *)
            echo "  - skip (foreign): $link -> $target"
            skipped=$((skipped + 1))
            continue
            ;;
    esac

    # Target must be a real file.
    if [[ ! -f "$target" ]]; then
        echo "  ✗ target missing: $link -> $target" >&2
        skipped=$((skipped + 1))
        continue
    fi

    # Replace the symlink with the file.
    rm "$link"
    mv "$target" "$link"
    echo "  ✓ restored: $link"
    restored=$((restored + 1))

done < <(find "$TARGET" -type l -print0)

echo ""
echo "restored: $restored, skipped: $skipped"
