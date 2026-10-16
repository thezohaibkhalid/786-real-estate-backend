#!/usr/bin/env bash
set -euo pipefail
[[ $EUID == 0 ]] || exit 1
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y ca-certificates curl gnupg nginx certbot python3-certbot-nginx git jq
if ! command -v docker >/dev/null; then
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  chmod a+r /etc/apt/keyrings/docker.asc
  printf 'Types: deb\nURIs: https://download.docker.com/linux/ubuntu\nSuites: noble\nComponents: stable\nArchitectures: amd64\nSigned-By: /etc/apt/keyrings/docker.asc\n' > /etc/apt/sources.list.d/docker.sources
  apt-get update
  apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
fi
systemctl enable --now docker nginx
# Four GiB of swap absorbs deployment and image-processing peaks on this 1 GiB VM.
if ! swapon --show=NAME --noheadings | grep -qx /swapfile; then
  if [[ ! -e /swapfile ]]; then
    fallocate -l 4G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
  fi
  swapon /swapfile
fi
grep -q '^/swapfile ' /etc/fstab || printf '/swapfile none swap sw 0 0\n' >> /etc/fstab
printf 'vm.swappiness=10\n' > /etc/sysctl.d/90-786-memory.conf
sysctl -p /etc/sysctl.d/90-786-memory.conf
install -d -m 0755 /opt/786/deploy /usr/local/lib/786
install -d -m 0700 /opt/786/env /opt/786/state
if ! id deploy786 >/dev/null 2>&1; then
  useradd --create-home --shell /bin/bash deploy786
fi
install -d -o root -g root -m 0755 /home/deploy786/.ssh
install -d -o deploy786 -g deploy786 -m 0700 /var/lib/786-deploy/incoming
install -d -o root -g root -m 0755 /var/lib/786-deploy
# Restrict pipeline users to the validated deploy entry point. No Docker group.
printf 'deploy786 ALL=(root) NOPASSWD: /usr/local/sbin/786-deploy *\n' > /etc/sudoers.d/786-deploy
chmod 0440 /etc/sudoers.d/786-deploy
visudo -cf /etc/sudoers.d/786-deploy
sudo -u ubuntu mkdir -p /home/ubuntu/.ssh
: > /home/deploy786/.ssh/authorized_keys
for service in frontend admin backend; do
  for kind in actions github; do
    key="/home/ubuntu/.ssh/786_${kind}_${service}"
    if [[ ! -f "$key" ]]; then
      sudo -u ubuntu ssh-keygen -q -t ed25519 -N '' -C "786-${kind}-${service}@130.210.33.71" -f "$key"
    fi
  done
  printf 'restrict,command="/usr/local/lib/786/receive-deploy %s" ' "$service" >> /home/deploy786/.ssh/authorized_keys
  cat "/home/ubuntu/.ssh/786_actions_${service}.pub" >> /home/deploy786/.ssh/authorized_keys
  if ! grep -q "^Host github-786-${service}$" /home/ubuntu/.ssh/config 2>/dev/null; then
    printf '\nHost github-786-%s\n  HostName github.com\n  User git\n  IdentityFile ~/.ssh/786_github_%s\n  IdentitiesOnly yes\n' "$service" "$service" >> /home/ubuntu/.ssh/config
  fi
done
chown ubuntu:ubuntu /home/ubuntu/.ssh/config
chmod 600 /home/ubuntu/.ssh/config
chmod 0644 /home/deploy786/.ssh/authorized_keys
if [[ ! -f /etc/docker/daemon.json ]]; then
  printf '{"log-driver":"json-file","log-opts":{"max-size":"10m","max-file":"3"}}\n' > /etc/docker/daemon.json
  systemctl restart docker
fi
docker network inspect realestate786_default >/dev/null 2>&1 || docker network create realestate786_default
if [[ ! -f /opt/786/state/images.env ]]; then
  printf 'FRONTEND_IMAGE=786/frontend:unconfigured\nADMIN_IMAGE=786/admin:unconfigured\nBACKEND_IMAGE=786/backend:unconfigured\n' > /opt/786/state/images.env
  chmod 0600 /opt/786/state/images.env
fi
systemctl enable docker nginx
free -h
docker compose version
