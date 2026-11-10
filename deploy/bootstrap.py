#!/usr/bin/env python3
"""Первичная настройка постоянной конфигурации и паролей для tournament."""

import argparse
import json
import os
from pathlib import Path
import secrets
from urllib.parse import urlsplit


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        '--root',
        default='/home/su/tournament',
        help='Абсолютный путь к рабочей директории приложения на сервере',
    )
    parser.add_argument(
        '--url',
        default='https://tournament.schooltech.ru',
        help='Внешний HTTPS URL приложения (без путей)',
    )
    args = parser.parse_args()

    root = Path(args.root).resolve()
    url = args.url.rstrip('/')
    parsed = urlsplit(url)

    if parsed.scheme not in ('http', 'https') or not parsed.hostname or parsed.path:
        parser.error('--url должен быть формата http(s)://domain.com без путей')

    if any(c.isspace() for c in str(root)):
        parser.error('--root не должен содержать пробелов')

    # Ограничиваем права доступа (только для владельца)
    os.umask(0o077)

    shared = root / 'shared'
    shared.mkdir(parents=True, exist_ok=True, mode=0o700)

    env_path = shared / '.env'
    config_path = shared / 'docker.yaml'

    # Защита от перезаписи: если конфигурация уже существует, выходим
    if env_path.exists() or config_path.exists():
        print(f'Конфигурация уже существует в {shared}. Секреты не изменены.')
        return

    # Генерация криптографически стойкого пароля для PostgreSQL
    db_pass = secrets.token_hex(24)
    db_user = 'tournament_user'
    db_name = 'tournament_db'

    # Структура точно под config.go: DB_INFO, HTTP_SERVER_INFO, FRONTEND_INFO, env
    backend_config = {
        'env': 'prod',
        'DB_INFO': {
            'host': 'db',
            'user': db_user,
            'pass': db_pass,
            'name': db_name,
            'ssl_model': 'disable',
        },
        'HTTP_SERVER_INFO': {
            'host': '0.0.0.0',
            'port': '8080',
        },
        'FRONTEND_INFO': url,
    }

    # Записываем YAML-конфигурацию (JSON валиден для cleanenv)
    with config_path.open('x', encoding='utf-8') as f:
        json.dump(backend_config, f, ensure_ascii=False, indent=2)
        f.write('\n')

    # Записываем .env для docker compose (db и migrator)
    with env_path.open('x', encoding='utf-8') as f:
        f.write(
            f'DB_USER={db_user}\n'
            f'DB_PASS={db_pass}\n'
            f'DB_NAME={db_name}\n'
            f'APP_ROOT={root}\n'
        )

    print(f'Успешно создана постоянная конфигурация в: {shared}')
    print('Секреты безопасно сохранены и не выведены на экран.')


if __name__ == '__main__':
    main()