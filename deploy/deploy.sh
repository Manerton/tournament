#!/usr/bin/env sh
set -eu
: "${TOURNAMENT_SSH_KEY:?Set a GitLab CI/CD File variable TOURNAMENT_SSH_KEY}"
: "${CI_COMMIT_SHA:?}"
deploy_host=${TOURNAMENT_DEPLOY_HOST:-172.16.0.164}
deploy_user=${TOURNAMENT_DEPLOY_USER:-su}
deploy_root=/home/su/tournament
test -f "$TOURNAMENT_SSH_KEY"
umask 077
mkdir -p "$HOME/.ssh"
cp "$TOURNAMENT_SSH_KEY" "$HOME/.ssh/tournament_deploy"
cp deploy/ssh_known_hosts "$HOME/.ssh/known_hosts"
export GIT_SSH_COMMAND="ssh -i $HOME/.ssh/tournament_deploy -o IdentitiesOnly=yes -o BatchMode=yes -o StrictHostKeyChecking=yes -o ConnectTimeout=15"
remote="$deploy_user@$deploy_host"
# The archive contains only files tracked at this GitLab commit.
git archive --format=tar HEAD > /tmp/tournament-release.tar
$GIT_SSH_COMMAND "$remote" "mkdir -p '$deploy_root/releases/$CI_COMMIT_SHA'"
$GIT_SSH_COMMAND "$remote" "tar -xf - -C '$deploy_root/releases/$CI_COMMIT_SHA'" < /tmp/tournament-release.tar
$GIT_SSH_COMMAND "$remote" "bash '$deploy_root/releases/$CI_COMMIT_SHA/deploy/release.sh' '$CI_COMMIT_SHA' '$deploy_root'"