#!/bin/sh

# Remove the r2 CLI from where the current (/usr/local/bin) and older (/usr/bin) install scripts put
# it. Every release contains its R2 endpoint, so a file without it, such as radare2's r2, is left
# alone.
for path in /usr/local/bin/r2 /usr/bin/r2; do
  [ -e "$path" ] || continue
  if grep -qF "r2.cloudflarestorage.com" "$path" 2>/dev/null; then
    rm -f "$path" && echo "Removed $path"
  else
    echo "Skipped $path: not the r2 CLI for Cloudflare R2"
  fi
done
