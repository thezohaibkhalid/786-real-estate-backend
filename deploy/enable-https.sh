#!/usr/bin/env bash
set -euo pipefail
[[ $EUID == 0 ]] || { echo 'Run as root' >&2; exit 1; }
cd /opt/786
# Validate and render HTTP challenge routing for the configured names first.
python3 deploy/configure.py --nginx-output /etc/nginx/sites-available/786-real-estate
nginx -t
systemctl reload nginx
mapfile -t hosts < <(python3 - <<'PY'
import sys
from pathlib import Path
from urllib.parse import urlsplit
sys.path.insert(0,'deploy')
from configure import load_env,origin
values=load_env(Path('.env'))
for key in ['PUBLIC_SITE_URL','ADMIN_SITE_URL']:
 print(urlsplit(origin(values[key],key)).hostname)
PY
)
[[ ${#hosts[@]} == 2 ]] || exit 1
certbot certonly --webroot -w /var/www/letsencrypt --cert-name 786-real-estate \
  -d "${hosts[0]}" -d "${hosts[1]}" --renew-with-new-domains --keep-until-expiring \
  --non-interactive --agree-tos --register-unsafely-without-email
/opt/786/deploy/apply-env.sh
