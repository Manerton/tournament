#!/usr/bin/env bash
set -euo pipefail

revision=${1:?Usage: release.sh COMMIT_SHA [DEPLOY_ROOT]}
deploy_root=${2:-/home/su/tournament}

if [[ ! $revision =~ ^[0-9a-f]{40}$ ]]; then
    echo 'Expected a full Git commit SHA' >&2
    exit 2
fi

release_dir="$deploy_root/releases/$revision"
shared_dir="$deploy_root/shared"

cd "$release_dir"

# Check the existing server configuration without printing secrets.
test -f "$shared_dir/.env"
test -f "$shared_dir/docker.yaml"
chmod 0700 "$shared_dir"
chmod 0600 "$shared_dir/.env"
# The container runs as UID 10001; shared/ stays private on the host.
chmod 0644 "$shared_dir/docker.yaml"

# Keep shared mounts stable and retain images for each release.
export APP_ROOT="$deploy_root"
export RELEASE_REVISION="$revision"

compose=(docker compose -p tournament --env-file "$shared_dir/.env" -f docker-compose.prod.yml)

echo '===> Validating Docker Compose configuration...'
"${compose[@]}" config --quiet

echo "===> Building images for release $revision..."
"${compose[@]}" build --pull backend frontend

previous_revision=$(cat "$deploy_root/current-revision" 2>/dev/null || true)

check_application() {
    # Check both the static frontend and a public API that queries PostgreSQL.
    # --retry-all-errors: right after container start docker-proxy can accept and
    # immediately reset connections (curl 56) before nginx/backend are ready.
    curl --fail --silent --show-error --retry 12 --retry-all-errors \
        --retry-delay 2 --max-time 5 -o /dev/null http://127.0.0.1:8081/ || return 1
    curl --fail --silent --show-error --retry 12 --retry-all-errors \
        --retry-delay 2 --max-time 5 -o /dev/null \
        http://127.0.0.1:8081/backend/api/locations/districts || return 1
}

echo '===> Starting services and checking application availability...'
if ! "${compose[@]}" up -d --no-build --wait --wait-timeout 180 || ! check_application; then
    echo '!!! Release failed; inspecting service status.' >&2
    "${compose[@]}" ps || true
    echo '!!! Service logs (backend, frontend, migrator):' >&2
    "${compose[@]}" logs --tail 100 backend frontend migrator || true

    # Restore application images; applied database migrations remain in place.
    if [[ $previous_revision =~ ^[0-9a-f]{40}$ && $previous_revision != "$revision" ]]; then
        prev_dir="$deploy_root/releases/$previous_revision"
        if [ -d "$prev_dir" ]; then
            echo "===> Restoring previous release: $previous_revision"
            cd "$prev_dir"
            RELEASE_REVISION="$previous_revision" docker compose -p tournament \
                --env-file "$shared_dir/.env" -f docker-compose.prod.yml \
                up -d --no-build --wait --wait-timeout 120
            check_application
        fi
    fi
    exit 1
fi

echo '===> Recording the successful release...'
ln -sfn "$release_dir" "$deploy_root/current"
printf '%s\n' "$revision" > "$deploy_root/current-revision"

# Retain release sources and tagged images for rollback.
echo "=== Release $revision deployed successfully! ==="
