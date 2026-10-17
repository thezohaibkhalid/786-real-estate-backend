#!/usr/bin/env python3
"""Render least-privilege app environments and Nginx from /opt/786/.env.

Values in the canonical file are literal KEY=value, without shell expansion.
Generated service files must never be edited by hand.
"""
import argparse
import os
import re
import subprocess
from pathlib import Path
from urllib.parse import urlsplit


def load_env(path):
    values = {}
    for number, line in enumerate(path.read_text().splitlines(), 1):
        if not line.strip() or line.lstrip().startswith('#'):
            continue
        key, separator, value = line.partition('=')
        if not separator or not re.fullmatch(r'[A-Z][A-Z0-9_]*', key):
            raise ValueError(f'Invalid environment key on line {number}')
        values[key] = value
    return values


def origin(value, key):
    parsed = urlsplit(value)
    if (parsed.scheme != 'https' or not parsed.hostname or parsed.username or parsed.password
            or parsed.path not in ('', '/') or parsed.query or parsed.fragment or parsed.port
            or not re.fullmatch(r'[a-z0-9][a-z0-9.-]*[a-z0-9]', parsed.hostname)):
        raise ValueError(f'{key} must be an HTTPS origin without a port or path')
    return 'https://' + parsed.hostname


def write_private(path, values):
    content = ''.join(f'{key}={value}\n' for key, value in values.items())
    if path.exists() and path.read_text() == content:
        return
    temporary = path.with_suffix('.tmp')
    temporary.write_text(content)
    temporary.chmod(0o600)
    temporary.replace(path)


def render(root, nginx_output=None, cert_root=Path('/etc/letsencrypt/live/786-real-estate')):
    os.umask(0o077)
    values = load_env(root / '.env')
    public = origin(values['PUBLIC_SITE_URL'], 'PUBLIC_SITE_URL')
    admin = origin(values['ADMIN_SITE_URL'], 'ADMIN_SITE_URL')
    if public == admin:
        raise ValueError('The public and admin apps require different hostnames')
    for key in ['DATABASE_URL', 'SESSION_SECRET', 'REVALIDATE_SECRET']:
        if not values.get(key):
            raise ValueError(f'{key} is required')
    if len(values['SESSION_SECRET']) < 32 or len(values['REVALIDATE_SECRET']) < 32:
        raise ValueError('Production secrets must contain at least 32 characters')
    env_dir = root / 'env'
    env_dir.mkdir(mode=0o700, exist_ok=True)
    backend = {key: value for key, value in values.items()
               if key.startswith(('EMAIL_', 'R2_')) or key in
               ['DATABASE_URL', 'SESSION_SECRET', 'DB_MAX_CONNS', 'TRUSTED_PROXIES']}
    backend.update(APP_ENV='production', HTTP_ADDR='0.0.0.0:8080', ADMIN_ORIGIN=admin)
    frontend = {'NODE_ENV': 'production', 'SITE_URL': public,
                'API_INTERNAL_BASE_URL': values.get('API_INTERNAL_BASE_URL', 'http://backend:8080/api'),
                'REVALIDATE_SECRET': values['REVALIDATE_SECRET']}
    admin_env = {'NODE_ENV': 'production', 'PUBLIC_SITE_URL': public,
                 'ADMIN_API_PROXY_URL': values.get('ADMIN_API_PROXY_URL', 'http://backend:8080/api/admin'),
                 'PUBLIC_SITE_INTERNAL_URL': values.get('PUBLIC_SITE_INTERNAL_URL', 'http://frontend:3000'),
                 'REVALIDATE_SECRET': values['REVALIDATE_SECRET']}
    for name, env in [('backend', backend), ('frontend', frontend), ('admin', admin_env)]:
        write_private(env_dir / (name + '.env'), env)
    if nginx_output:
        template = (root / 'deploy/nginx.conf').read_text()
        hosts = [urlsplit(public).hostname, urlsplit(admin).hostname]
        certificate = cert_root / 'fullchain.pem'
        key = cert_root / 'privkey.pem'
        tls = False
        if certificate.exists() and key.exists():
            text = subprocess.check_output(['openssl', 'x509', '-in', str(certificate), '-noout', '-ext', 'subjectAltName'], text=True)
            names = set(re.findall(r'DNS:([^,\s]+)', text))
            tls = all(host in names for host in hosts)
        listen = 'listen 80;\n    listen [::]:80;'
        if tls:
            listen = ('listen 443 ssl;\n    listen [::]:443 ssl;\n'
                      f'    ssl_certificate {certificate};\n    ssl_certificate_key {key};\n'
                      '    ssl_protocols TLSv1.2 TLSv1.3;')
        content = template.replace('@@LISTEN@@', listen).replace('@@PUBLIC_HOST@@', hosts[0]).replace('@@ADMIN_HOST@@', hosts[1])
        if tls:
            content += ('\nserver {\n    listen 80;\n    listen [::]:80;\n'
                        f'    server_name {" ".join(hosts)};\n'
                        '    location ^~ /.well-known/acme-challenge/ { root /var/www/letsencrypt; }\n'
                        '    location / { return 301 https://$host$request_uri; }\n}\n')
        nginx_output.write_text(content)
        nginx_output.chmod(0o644)
    return public, admin


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--root', type=Path, default=Path('/opt/786'))
    parser.add_argument('--nginx-output', type=Path)
    args = parser.parse_args()
    try:
        public, admin = render(args.root, args.nginx_output)
        print(f'Configuration rendered from .env: {public}, {admin}; secrets not logged.')
    except (ValueError, KeyError) as error:
        raise SystemExit(f'Invalid production .env: {error}')
