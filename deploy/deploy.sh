#!/usr/bin/env bash
set -euo pipefail
: "${TOURNAMENT_SSH_KEY:?Set a GitLab CI/CD File variable TOURNAMENT_SSH_KEY}"
: "${CI_COMMIT_SHA:?}"
deploy_host=${TOURNAMENT_DEPLOY_HOST:-172.16.0.164}
deploy_user=${TOURNAMENT_DEPLOY_USER:-su}
deploy_root=${TOURNAMENT_DEPLOY_ROOT:-/home/su/tournament}
if [[ ! $CI_COMMIT_SHA =~ ^[0-9a-f]{40}$ || ! $deploy_root =~ ^/[a-zA-Z0-9_/-]+$ ]]; then
    echo 'Expected a full Git commit SHA and an absolute deployment path without special characters' >&2
    exit 2
fi
test -f "$TOURNAMENT_SSH_KEY"
chmod 0400 "$TOURNAMENT_SSH_KEY"
ssh_command=(ssh -i "$TOURNAMENT_SSH_KEY"
    -o IdentitiesOnly=yes -o BatchMode=yes -o StrictHostKeyChecking=yes
    -o UserKnownHostsFile="${TOURNAMENT_SSH_KNOWN_HOSTS:-$PWD/deploy/ssh_known_hosts}"
    -o ConnectTimeout=15 -o ServerAliveInterval=15 -o ServerAliveCountMax=4)
remote="$deploy_user@$deploy_host"
# The archive contains only files tracked at this GitLab commit.
archive=$(mktemp)
trap 'rm -f "$archive"' EXIT
git archive --format=tar "$CI_COMMIT_SHA" > "$archive"
"${ssh_command[@]}" "$remote" "mkdir -p '$deploy_root/releases/$CI_COMMIT_SHA'"
"${ssh_command[@]}" "$remote" "tar -xf - -C '$deploy_root/releases/$CI_COMMIT_SHA'" < "$archive"
"${ssh_command[@]}" "$remote" "bash '$deploy_root/releases/$CI_COMMIT_SHA/deploy/release.sh' '$CI_COMMIT_SHA' '$deploy_root'"
