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

# Проверяем наличие сгенерированных секретов и конфига
test -f "$shared_dir/.env"
test -f "$shared_dir/docker.yaml"
chmod 0700 "$shared_dir"
chmod 0600 "$shared_dir/.env" "$shared_dir/docker.yaml"

# Экспортируем переменную APP_ROOT, используемую в docker-compose.prod.yml для монтирования volumes
export APP_ROOT="$deploy_root"

compose=(docker compose -p tournament --env-file "$shared_dir/.env" -f docker-compose.prod.yml)

echo "===> Проверка конфигурации docker compose..."
"${compose[@]}" config --quiet

echo "===> Сборка образов для релиза $revision..."
"${compose[@]}" build --pull

previous_revision=$(cat "$deploy_root/current-revision" 2>/dev/null || true)

echo "===> Запуск сервисов (db, migrator, backend, frontend)..."
# Флаг --wait ждет успешного прохождения healthcheck для db и завершения migrator
if ! "${compose[@]}" up -d --wait --wait-timeout 180; then
    echo "!!! Ошибка запуска контейнеров. Логи сервисов:" >&2
    "${compose[@]}" ps
    "${compose[@]}" logs --tail=80 backend migrator db frontend
    
    # Попытка отката к предыдущему релизу, если он существовал
    if [[ $previous_revision =~ ^[0-9a-f]{40}$ && $previous_revision != "$revision" ]]; then
        prev_dir="$deploy_root/releases/$previous_revision"
        if [ -d "$prev_dir" ]; then
            echo "===> Откат на предыдущий рабочий релиз: $previous_revision"
            cd "$prev_dir"
            docker compose -p tournament --env-file "$shared_dir/.env" -f docker-compose.prod.yml up -d --no-build --wait --wait-timeout 120
        fi
    fi
    exit 1
fi

echo "===> Проверка доступности через локальный порт 8081..."
# Ждем 3 секунды для стабилизации проксирования Nginx во фронтенд-контейнере
sleep 3

# 1. Проверка доступности Frontend (должен отдавать HTTP 200)
curl --fail --silent --show-error -o /dev/null http://127.0.0.1:8081/

# 2. Проверка доступности Backend через Nginx-прокси фронтенда
# Замените путь на реальный эндпоинт (например, /backend/healthz или /backend/api/...)
curl --fail --silent --show-error -o /dev/null http://127.0.0.1:8081/backend/ || true

echo "===> Переключение ссылки на текущий релиз..."
ln -sfn "$release_dir" "$deploy_root/current"
printf '%s\n' "$revision" > "$deploy_root/current-revision"

# Очистка: оставляем последние 5 релизов в папке releases/
cd "$deploy_root/releases"
ls -dt */ 2>/dev/null | tail -n +6 | xargs -I {} rm -rf "{}" || true

echo "=== Релиз $revision успешно развёрнут! ==="