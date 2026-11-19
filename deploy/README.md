# Tournament CI/CD

Merge requests and the default branch run Bash/ShellCheck checks, Go vet/tests,
compilation of both Go binaries, and a Deno frontend build. Successful pipelines
on the **protected default branch** deploy automatically to `172.16.0.164`.
Deployment jobs are serialized with the `tournament-deploy` resource group.

## GitLab setup

1. Protect the default branch in **Settings > Repository > Branch rules**.
2. In **Settings > CI/CD > Runners**, enable an online Linux runner with the
   Docker executor and the `tournament` tag used by `default.tags` in
   `.gitlab-ci.yml`. The project runner is `for tournament deploy` (#5), managed
   on `172.16.1.105`. Its service account needs access to the Docker socket.
   It must reach GitLab, Docker Hub, dependency registries, and
   `172.16.0.164:22`. Docker-in-Docker and privileged mode are unnecessary:
   production images are built by Docker on the deployment server.
3. Add the following variables in **Settings > CI/CD > Variables**.
4. In **Settings > CI/CD > General pipelines**, enable **Prevent outdated
   deployment jobs** to prevent an older queued pipeline from replacing a
   newer release.

The tournament runner's existing user configuration is
`/home/su/.gitlab-runner/config.toml`. Run it through
`gitlab-runner-tournament.service`, which grants the service the supplementary
`docker` group and starts it after reboot. The separate system runner service
and `/etc/gitlab-runner/config.toml` remain managed independently. Do not start
an additional manual `gitlab-runner run` process for the same user configuration.

| Variable | Type | Required / default |
| --- | --- | --- |
| `TOURNAMENT_SSH_KEY` | File | Required; protected SSH private key without a passphrase |
| `TOURNAMENT_SSH_KNOWN_HOSTS` | File | Optional; defaults to `deploy/ssh_known_hosts` |
| `TOURNAMENT_DEPLOY_HOST` | Variable | `172.16.0.164` |
| `TOURNAMENT_DEPLOY_USER` | Variable | `su` |
| `TOURNAMENT_DEPLOY_ROOT` | Variable | `/home/su/tournament` |

Mark the key **Protected** and scope it to `production` (or `*`). Multiline SSH
keys need **Visible** visibility because GitLab masking does not accept
whitespace; store them only in CI/CD settings. End the key value with a newline.
Install its public key in the deployment user's SSH `authorized_keys` on the
server. Verify the committed server host key through a trusted channel before
the first deployment; provide `TOURNAMENT_SSH_KNOWN_HOSTS` when changing hosts.
The pipeline keeps strict host-key checking enabled.

## Server prerequisites

The deployment account needs Bash, Git-independent SSH access, curl, Docker,
Docker Compose v2 with `up --wait`, and permission to use Docker. The server
must reach Docker Hub and the Go/Deno dependency registries during builds.

The existing persistent configuration must be present:

- `/home/su/tournament/shared/.env` with the existing database settings;
- `/home/su/tournament/shared/docker.yaml` with backend configuration.

These files stay on the server and must not be committed, printed, or replaced
during CI/CD setup. The configuration directory uses mode `0700`; `.env` uses
`0600`. The read-only backend bind mount uses `0644` so the container's UID 10001
can read it; the private parent directory restricts access on the host.
The production Compose project is named `tournament`, with persistent
database volume `tournament_pgdata`; verify the existing database volume before
the first deployment if the application was started with another project name.

The frontend binds to `127.0.0.1:8081`. Host nginx should proxy the application
domain there, as shown in `nginx-tournament.conf`.

## Deployment and rollback

The job archives the exact GitLab commit and transfers it over SSH into
`releases/<commit-sha>`. The server validates Compose configuration, builds
`tournament-backend:<commit-sha>` and `tournament-frontend:<commit-sha>`, runs
migrations, and starts the services. It checks both `/` and the public database
API `/backend/api/locations/districts` through the frontend proxy.

Only a successful deployment updates `current` and `current-revision`. If
startup or either HTTP check fails, the script attempts to restore the images
from the previous successful release and leaves the CI job failed. Database
migrations are **not reverted**; migrations must remain compatible with the
previous application version. The first deployment has no previous release to
restore.

Release directories and tagged images are retained for rollback. Schedule
cleanup separately, keeping the current and previous successful releases and
their images; never prune the database volume.

For a manual application rollback on the server:

```bash
bash /home/su/tournament/releases/<previous-sha>/deploy/release.sh <previous-sha>
```

This command rebuilds that release before starting it and records it as current
after successful HTTP checks. Automatic rollback uses the already built images.

References: [GitLab SSH keys](https://docs.gitlab.com/ci/jobs/ssh_keys/),
[deployment safety](https://docs.gitlab.com/ci/environments/deployment_safety/),
[Deno images](https://docs.deno.com/runtime/reference/docker/).
## Dependency and compiler caches

GitLab CI caches downloaded Go modules and Deno packages separately from the
Go compiler cache. Dependency archive keys include the dependency manifests;
compiler keys include the toolchain and Linux/amd64 target. Cache archives use
Zstandard with fast compression. Keep GitLab's separate caches for protected
and non-protected branches enabled. Secrets and runtime data are not cached.

Production Docker builds use BuildKit cache mounts for Go modules, Go compilation
and Deno downloads. Mount IDs are scoped to the project, toolchain and target
platform, and writes are locked. The Deno mount is present during installation
and compilation. Go modules are verified before compilation.

The first build fills an empty cache; later builds reuse it. A missing or pruned
cache must remain safe: tools download the required versions again. Frozen
Deno installation, release builds, health checks and rollback remain required.
The deploy script still checks for updated base images with `--pull`.
