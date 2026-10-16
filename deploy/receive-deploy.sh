#!/usr/bin/env bash
# Forced SSH command: the key fixes the service; the command is only a Git SHA.
set -euo pipefail
service=${1:?Missing service}
case "$service" in frontend|admin|backend) ;; *) exit 64 ;; esac
sha=${SSH_ORIGINAL_COMMAND:-}
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || { echo 'Expected a 40-character commit SHA' >&2; exit 64; }
umask 077
archive=$(mktemp "/var/lib/786-deploy/incoming/${service}-${sha}.XXXXXX.tar.gz")
trap 'rm -f "$archive"' EXIT
# Bound the stream to 1.5 GiB; a failed/oversized transfer cannot deploy.
python3 -c 'import sys
size=0
with open(sys.argv[1],"wb") as out:
 while chunk := sys.stdin.buffer.read(1024*1024):
  size+=len(chunk)
  if size > 1536*1024*1024: raise SystemExit("Deployment archive too large")
  out.write(chunk)' "$archive"
sudo -n /usr/local/sbin/786-deploy "$service" "$sha" "$archive"
