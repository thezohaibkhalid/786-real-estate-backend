# Production deployment

Three repositories deploy independently to Ubuntu at `130.210.33.71`:

- Website: `https://786.zohaibkhalid.me` → Nginx → frontend container.
- Admin: `https://786-admin.zohaibkhalid.me` → Nginx → admin container.
- Public API: website `/api/` → backend container. A separate API hostname is unnecessary.
- Admin API: admin `/api/admin/` → backend container. Public-site admin API access is blocked.

Push or merge into **production** in the corresponding repository. `main` pushes and pull requests run validation without deployment. Manual workflow runs deploy only when the selected branch is `production`.

## Release process

GitHub-hosted Ubuntu runners run lint, tests, TypeScript/Go checks and production builds. Backend CI provisions a disposable PostgreSQL database and exercises real migration and persistence tests. Docker builds create Linux amd64 images tagged by the exact Git commit. No application build runs on the 1 GiB production VM.

The workflow streams the compressed image over SSH to a forced-command deployment account. No container registry login, GitHub token, source checkout or git pull is needed to deploy. The SSH host key is pinned. Each repository has its own key, restricted to its own service; `deploy786` has no unrestricted SSH key or Docker-group membership.

The VM serializes releases using `flock`, loads the image, runs transactional backend migrations before replacing the API, and checks container health plus an HTTP smoke request. A failed application health check restores the preceding image when available. Database migrations are forward-only: image rollback does not reverse schema changes, so migrations must remain compatible with the preceding application release.

Only Nginx binds publicly to ports 80/443. Application ports 3000, 3001 and 8080 bind to loopback. Containers run as non-root, drop Linux capabilities, restart after reboot, have memory/process limits, and rotate logs. A 4 GiB swap file accommodates memory spikes; sustained traffic and concurrent large image uploads still require monitoring on this 1 GiB VM.

## GitHub configuration

Each repo requires encrypted Actions secrets:

- `PRODUCTION_HOST`: VM IP/hostname.
- `PRODUCTION_SSH_KEY`: that service's dedicated Actions-to-VM private key.
- `PRODUCTION_SSH_HOST_KEY`: a known_hosts line for the VM's trusted Ed25519 host key.

No application URLs or credentials are compiled into images or stored in GitHub variables. The admin “View Site” route reads `PUBLIC_SITE_URL` at runtime. Do not put database, R2, SMTP or cache-refresh credentials in Actions variables or Docker build arguments.

The production source of truth is root-owned `/opt/786/.env`, mode 600. `deploy/configure.py` generates `/opt/786/env/{backend,frontend,admin}.env`, also mode 600, with only the values each app needs. Never edit generated files. After changing `.env`, run `sudo /opt/786/deploy/apply-env.sh`; this regenerates Nginx and app configuration and recreates deployed containers as needed. Compose uses raw env-file parsing so `$` in credentials is preserved. `.dockerignore` excludes local secrets and SSH keys from build contexts. The Docker bridge gateway is the API's only `TRUSTED_PROXIES` entry; Nginx overwrites `X-Real-IP`. Cloudflare source ranges may be trusted by Nginx for the client-IP header, while arbitrary incoming forwarding headers remain untrusted.

## Provisioning and installed files

Run `sudo bash deploy/bootstrap.sh` once on a fresh Ubuntu 24.04 amd64 server. Install:

- `compose.production.yaml` at `/opt/786/deploy/compose.production.yaml`.
- `deploy.sh` at `/usr/local/sbin/786-deploy`, mode 755.
- `receive-deploy.sh` at `/usr/local/lib/786/receive-deploy`, mode 755.
- `nginx-proxy.conf` at `/etc/nginx/snippets/786-proxy.conf`.
- The Nginx template `nginx.conf`, `configure.py`, `apply-env.sh` and `enable-https.sh` under `/opt/786/deploy/`. The renderer creates `/etc/nginx/sites-available/786-real-estate`, symlinked into sites-enabled.

Populate `/opt/786/.env` from `production.env.example` and the existing backend credentials before deployment. Use the existing hosted PostgreSQL database and R2 bucket; this stack does not run PostgreSQL on the VM. Add the two DNS A records, then run `sudo /opt/786/deploy/enable-https.sh` to issue certificates and enable HTTPS redirects. Certbot's renewal timer and an actual dry run should be verified.

Bootstrap creates separate Ubuntu-to-GitHub read-only keys at `~ubuntu/.ssh/786_github_{frontend,admin,backend}`. Add each `.pub` as a read-only deploy key on its matching repository. Their SSH aliases are `github-786-frontend`, `github-786-admin`, and `github-786-backend`. Example:

```sh
git clone git@github-786-frontend:thezohaibkhalid/786-real-estate-frontend.git
```

The private keys stay on Ubuntu. Read-only GitHub keys are separate from the forced-command Actions-to-VM keys. Register GitHub's official host keys before connecting.

## Operations

```sh
sudo docker compose --env-file /opt/786/state/images.env -f /opt/786/deploy/compose.production.yaml ps
sudo docker compose --env-file /opt/786/state/images.env -f /opt/786/deploy/compose.production.yaml logs --tail=100 backend
sudo cat /opt/786/state/deployments.log
sudo certbot renew --dry-run
free -h
```

Each service keeps its current and immediately preceding image. Tagged images are not globally pruned across applications. Application containers restart automatically; Nginx and Docker are enabled at boot. Backend migrations use checksums and stop the release if an already-applied migration was edited.

To change domains later: edit only `PUBLIC_SITE_URL` and `ADMIN_SITE_URL` in `/opt/786/.env`, point both new DNS records at this VM, and run `sudo /opt/786/deploy/enable-https.sh`. The command derives hostnames from `.env`, obtains certificates, regenerates Nginx, and applies the app environments. No application rebuild or GitHub variable change is needed. Backend `ADMIN_ORIGIN` and frontend `SITE_URL` are derived automatically. The admin's `PUBLIC_SITE_INTERNAL_URL=http://frontend:3000` deliberately stays internal for authenticated cache refresh. Both apps receive the same canonical `REVALIDATE_SECRET`.

Cloudflare can cache static assets. Do not force homepage HTML caching: the application currently invalidates its Next.js data cache, and does not purge a separate Cloudflare HTML cache. Never cache admin/auth/API mutation responses. Preserve Next.js `_rsc` query parameters and origin Cache-Control headers.
