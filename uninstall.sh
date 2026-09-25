#!/bin/sh

# Remove the r2 CLI from where the current (/usr/local/bin) and older (/usr/bin) install scripts put
# it. Every release contains its R2 endpoint, so a file without it, such as radare2's r2, is left
# alone. Exits non-zero if a copy couldn't be removed, e.g. when not run as root.
status=0
for path in /usr/local/bin/r2 /usr/bin/r2; do
  [ -e "$path" ] || continue
  if grep -qF "r2.cloudflarestorage.com" "$path" 2>/dev/null; then
    if rm -f "$path"; then
      echo "Removed $path"
    else
      echo "Couldn't remove $path; run this script with sudo." >&2
      status=1
    fi
  else
    echo "Skipped $path: not the r2 CLI for Cloudflare R2"
  fi
done
exit $status
