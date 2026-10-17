#!/usr/bin/env bash
set -euo pipefail
[[ $EUID == 0 ]] || { echo 'Run as root' >&2; exit 1; }
exec 9>/run/lock/786-deploy.lock
flock -w 900 9
cd /opt/786
python3 deploy/configure.py --nginx-output /etc/nginx/sites-available/786-real-estate
nginx -t
systemctl reload nginx
compose=(docker compose --env-file state/images.env -f deploy/compose.production.yaml)
services=()
for service in backend frontend admin; do
  variable="$(printf '%s' "$service" | tr '[:lower:]' '[:upper:]')_IMAGE"
  image=$(sed -n "s/^${variable}=//p" state/images.env)
  if docker image inspect "$image" >/dev/null 2>&1; then services+=("$service"); fi
done
if [[ ${#services[@]} -gt 0 ]]; then
  "${compose[@]}" up -d --no-deps --wait --wait-timeout 180 "${services[@]}"
fi
