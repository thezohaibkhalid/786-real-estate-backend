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

Admin also requires repository variable `PUBLIC_SITE_URL`, currently `https://786.zohaibkhalid.me`. It is compiled into the public “View Site” link. Do not put runtime database, R2, SMTP or cache-refresh credentials in Actions variables or Docker build arguments.

Runtime secrets exist only in root-owned `/opt/786/env/{backend,frontend,admin}.env`, mode 600. Compose uses raw env-file parsing so `$` in credentials is preserved. `.dockerignore` excludes local secrets and SSH keys from build contexts. The Docker bridge gateway is the API's only `TRUSTED_PROXIES` entry; Nginx overwrites `X-Real-IP`. Cloudflare source ranges may be trusted by Nginx for the client-IP header, while arbitrary incoming forwarding headers remain untrusted.

## Provisioning and installed files

Run `sudo bash deploy/bootstrap.sh` once on a fresh Ubuntu 24.04 amd64 server. Install:

- `compose.production.yaml` at `/opt/786/deploy/compose.production.yaml`.
- `deploy.sh` at `/usr/local/sbin/786-deploy`, mode 755.
- `receive-deploy.sh` at `/usr/local/lib/786/receive-deploy`, mode 755.
- `nginx-proxy.conf` at `/etc/nginx/snippets/786-proxy.conf`.
- `nginx.conf` at `/etc/nginx/sites-available/786-real-estate`, symlinked into sites-enabled.

Populate runtime env files before deployment. Use the existing hosted PostgreSQL database and R2 bucket; this stack does not run PostgreSQL on the VM. Add the two DNS A records, then use Certbot's Nginx integration to issue certificates and enable HTTPS redirects. Certbot's renewal timer and an actual dry run should be verified.

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

To change domains later: update Nginx server names and certificates; update `SITE_URL` in frontend.env and `ADMIN_ORIGIN` in backend.env; update GitHub admin variable `PUBLIC_SITE_URL`; redeploy frontend/admin/backend as appropriate. The admin's `PUBLIC_SITE_URL=http://frontend:3000` is deliberately internal for authenticated cache revalidation and does not need the public hostname. Both apps share the runtime `REVALIDATE_SECRET`.

Cloudflare can cache static assets. Do not force homepage HTML caching: the application currently invalidates its Next.js data cache, and does not purge a separate Cloudflare HTML cache. Never cache admin/auth/API mutation responses. Preserve Next.js `_rsc` query parameters and origin Cache-Control headers.
