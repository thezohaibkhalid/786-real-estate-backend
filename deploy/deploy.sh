#!/usr/bin/env bash
set -euo pipefail
[[ $EUID == 0 ]] || { echo 'Run as root' >&2; exit 1; }
service=${1:?Missing service}; sha=${2:?Missing commit SHA}; archive=${3:?Missing archive}
case "$service" in frontend|admin|backend) ;; *) exit 64 ;; esac
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || exit 64
[[ -f "$archive" && ! -L "$archive" ]] || exit 64
[[ "$(realpath "$archive")" == /var/lib/786-deploy/incoming/"$service-$sha".*.tar.gz ]] || exit 64
exec 9>/run/lock/786-deploy.lock
flock -w 900 9
cd /opt/786
python3 deploy/configure.py
compose=(docker compose --env-file state/images.env -f deploy/compose.production.yaml)
image="786/$service:$sha"
variable="$(printf '%s' "$service" | tr '[:lower:]' '[:upper:]')_IMAGE"
previous=$(sed -n "s/^${variable}=//p" state/images.env)
set_image() {
  python3 - "$variable" "$1" <<'PY'
import os,sys
from pathlib import Path
path=Path('state/images.env')
values=dict(line.split('=',1) for line in path.read_text().splitlines() if line and not line.startswith('#'))
values[sys.argv[1]]=sys.argv[2]
tmp=path.with_suffix('.tmp')
tmp.write_text(''.join(f'{key}={value}\n' for key,value in values.items()))
os.chmod(tmp,0o600)
os.replace(tmp,path)
PY
}
gzip -dc "$archive" | docker image load
docker image inspect "$image" >/dev/null
# Migrations are transactional and checksum-checked. A failed migration stops
# before replacing the old backend. Database changes require forward fixes.
if [[ "$service" == backend ]]; then
  docker run --rm --network realestate786_default --env-file env/backend.env --memory=192m --memory-swap=384m "$image" /app/migrate
fi
set_image "$image"
healthy=true
"${compose[@]}" up -d --no-deps --wait --wait-timeout 180 "$service" || healthy=false
if [[ "$healthy" == true ]]; then
  case "$service" in
    frontend) url=http://127.0.0.1:3000/ ;;
    admin) url=http://127.0.0.1:3001/api/health ;;
    backend) url=http://127.0.0.1:8080/health/ready ;;
  esac
  curl --fail --silent --show-error --retry 3 --retry-delay 3 --max-time 45 "$url" >/dev/null || healthy=false
fi
if [[ "$healthy" != true ]]; then
  "${compose[@]}" logs --tail=30 "$service" >&2 || true
  set_image "$previous"
  if docker image inspect "$previous" >/dev/null 2>&1; then
    "${compose[@]}" up -d --no-deps --wait --wait-timeout 180 "$service"
    echo 'Application health failed; previous image restored.' >&2
  else
    "${compose[@]}" rm --stop --force "$service"
    echo 'Initial deployment failed; no previous image exists.' >&2
  fi
  exit 1
fi
printf '%s %s %s\n' "$(date -u +%FT%TZ)" "$service" "$image" >> state/deployments.log
printf 'Deployed %s at %s\n' "$service" "$sha"
# Keep the current and immediately previous image for each service.
for candidate in $(docker image ls "786/$service" --format '{{.Repository}}:{{.Tag}}'); do
  if [[ "$candidate" != "$image" && "$candidate" != "$previous" ]]; then
    docker image rm "$candidate" >/dev/null 2>&1 || true
  fi
done
