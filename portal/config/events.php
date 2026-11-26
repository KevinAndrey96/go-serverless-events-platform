<?php

return [

    /*
    | Base URL of the events API: the API Gateway stage in AWS or the Go dev server locally.
    */
    'api_url' => env('EVENTS_API_URL', 'http://localhost:8080'),

    'timeout' => (int) env('EVENTS_API_TIMEOUT', 10),

    /*
    | Admin routes are protected with IAM. When signing is enabled the portal signs them with
    | SigV4 using the default AWS credential chain (environment, profile or task role).
    */
    'sign_requests' => (bool) env('EVENTS_API_SIGN', false),

    'region' => env('EVENTS_API_REGION', env('AWS_DEFAULT_REGION', 'us-east-1')),

    /*
    | Organizers enter and read dates in this zone; the API stores instants with their offset.
    */
    'timezone' => env('EVENTS_TIMEZONE', 'America/Bogota'),

    'receipt_max_kb' => 5120,

    /*
    | First administrator, created by `php artisan db:seed`.
    */
    'admin' => [
        'email' => env('ADMIN_EMAIL'),
        'password' => env('ADMIN_PASSWORD'),
    ],

];
