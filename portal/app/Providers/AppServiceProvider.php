<?php

namespace App\Providers;

use App\EventsApi\EventsApi;
use App\EventsApi\HttpEventsApi;
use App\EventsApi\RequestSigner;
use App\Models\User;
use Aws\Credentials\CredentialProvider;
use Aws\Credentials\CredentialsInterface;
use Aws\Signature\SignatureV4;
use Illuminate\Cache\RateLimiting\Limit;
use Illuminate\Http\Client\Factory;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Facades\RateLimiter;
use Illuminate\Support\ServiceProvider;

class AppServiceProvider extends ServiceProvider
{
    public function register(): void
    {
        $this->app->singleton(RequestSigner::class, function (): RequestSigner {
            $provider = CredentialProvider::memoize(CredentialProvider::defaultProvider());

            return new RequestSigner(
                new SignatureV4('execute-api', (string) config('events.region')),
                fn (): CredentialsInterface => $provider()->wait(),
            );
        });

        $this->app->singleton(EventsApi::class, fn (): EventsApi => new HttpEventsApi(
            $this->app->make(Factory::class),
            rtrim((string) config('events.api_url'), '/'),
            (int) config('events.timeout'),
            config('events.sign_requests') ? $this->app->make(RequestSigner::class) : null,
        ));
    }

    public function boot(): void
    {
        Gate::define('admin', fn (User $user): bool => $user->isAdmin());

        RateLimiter::for('registrations', fn (Request $request) => Limit::perMinute(10)->by((string) $request->ip()));
        RateLimiter::for('check-ins', fn (Request $request) => Limit::perMinute(120)->by((string) $request->user()?->getAuthIdentifier()));
    }
}
