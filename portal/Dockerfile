# Organizer and registration portal: Laravel 13 on PHP 8.4 with Apache.
FROM node:24-alpine AS assets
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY vite.config.js ./
COPY resources ./resources
RUN npm run build

FROM php:8.4-apache AS base
RUN apt-get update \
    && apt-get install -y --no-install-recommends libpng-dev libzip-dev unzip \
    && docker-php-ext-install gd pdo_mysql zip opcache \
    && rm -rf /var/lib/apt/lists/* \
    && a2enmod rewrite headers \
    && sed -ri -e 's!/var/www/html!/var/www/html/public!g' -e 's!<VirtualHost \*:80>!<VirtualHost *:8080>!' \
        /etc/apache2/sites-available/000-default.conf \
    && sed -ri 's!^Listen 80$!Listen 8080!' /etc/apache2/ports.conf \
    && printf 'ServerTokens Prod\nServerSignature Off\nServerName localhost\n' >> /etc/apache2/conf-available/security.conf
COPY --from=composer:2 /usr/bin/composer /usr/bin/composer
WORKDIR /var/www/html

# Test image: dev dependencies included, used by CI and `docker compose run portal-tests`.
FROM base AS test
COPY composer.json composer.lock ./
RUN composer install --no-interaction --no-progress --no-scripts --prefer-dist
COPY . .
COPY --from=assets /app/public/build ./public/build
RUN composer dump-autoload --optimize
CMD ["vendor/bin/phpunit"]

FROM base AS runtime
RUN mv "$PHP_INI_DIR/php.ini-production" "$PHP_INI_DIR/php.ini" \
    && printf 'upload_max_filesize=6M\npost_max_size=8M\nexpose_php=Off\n' > "$PHP_INI_DIR/conf.d/portal.ini"
COPY composer.json composer.lock ./
RUN composer install --no-dev --no-interaction --no-progress --no-scripts --prefer-dist
COPY . .
COPY --from=assets /app/public/build ./public/build
RUN composer dump-autoload --optimize --no-dev \
    && chown -R www-data:www-data storage bootstrap/cache
COPY docker/entrypoint.sh /usr/local/bin/portal-entrypoint
USER www-data
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD php -r 'exit(@file_get_contents("http://127.0.0.1:8080/up") === false ? 1 : 0);'
ENTRYPOINT ["portal-entrypoint"]
CMD ["apache2-foreground"]
