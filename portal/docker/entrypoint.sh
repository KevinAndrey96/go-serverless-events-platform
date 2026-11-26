#!/bin/sh
set -eu

if [ -z "${APP_KEY:-}" ]; then
    if [ "${APP_ENV:-production}" != "local" ]; then
        echo "APP_KEY is required outside local environments" >&2
        exit 1
    fi
    APP_KEY="$(php artisan key:generate --show --no-ansi)"
    export APP_KEY
    echo "Generated a temporary APP_KEY; sessions end when the container restarts" >&2
fi

php artisan config:cache
php artisan route:cache
php artisan view:cache

if [ "${RUN_MIGRATIONS:-false}" = "true" ]; then
    php artisan migrate --force
    php artisan db:seed --force
fi

exec "$@"
