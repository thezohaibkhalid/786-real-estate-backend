#!/usr/bin/env bash
# Forced SSH command: the key fixes the service; the command is only a Git SHA.
set -euo pipefail
service=${1:?Missing service}
case "$service" in frontend|admin|backend) ;; *) exit 64 ;; esac
sha=${SSH_ORIGINAL_COMMAND:-}
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || { echo 'Expected a 40-character commit SHA' >&2; exit 64; }
umask 077
archive=$(mktemp "/var/lib/786-deploy/incoming/${service}-${sha}.XXXXXX.json")
trap 'rm -f "$archive"' EXIT
# Only a small, short-lived registry credential crosses SSH; images use HTTPS.
python3 -c 'import sys
size=0
with open(sys.argv[1],"wb") as out:
 while chunk := sys.stdin.buffer.read(1024*1024):
  size+=len(chunk)
  if size > 65536: raise SystemExit("Deployment payload too large")
  out.write(chunk)' "$archive"
sudo -n /usr/local/sbin/786-deploy "$service" "$sha" "$archive"
