#!/usr/bin/env bash
set -euo pipefail
if [[ $EUID -ne 0 ]]; then
    echo 'Run with sudo: bash deploy/install-proxy.sh' >&2
    exit 1
fi
source_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
target=/etc/nginx/sites-available/tournament.schooltech.ru.conf
enabled=/etc/nginx/sites-enabled/tournament.schooltech.ru.conf
if [[ -e $target ]]; then
    echo "$target already exists; inspect it instead of overwriting it." >&2
    exit 1
fi
nginx -t
install -m 0644 "$source_dir/nginx-tournament.conf" "$target"
ln -s "$target" "$enabled"
if ! nginx -t; then
    unlink "$enabled"
    unlink "$target"
    exit 1
fi
systemctl reload nginx
echo 'HTTP proxy installed. After DNS resolves to this server, enable HTTPS:'
echo 'sudo certbot --nginx -d tournament.schooltech.ru --redirect'