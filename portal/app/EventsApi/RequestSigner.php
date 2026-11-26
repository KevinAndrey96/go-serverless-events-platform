<?php

namespace App\EventsApi;

use Aws\Credentials\CredentialsInterface;
use Aws\Signature\SignatureV4;
use Closure;
use Psr\Http\Message\RequestInterface;

/**
 * Signs admin requests with SigV4 so API Gateway can authorize them with IAM.
 */
final class RequestSigner
{
    private ?CredentialsInterface $credentials = null;

    /**
     * @param  Closure(): CredentialsInterface  $resolveCredentials
     */
    public function __construct(
        private readonly SignatureV4 $signature,
        private readonly Closure $resolveCredentials,
    ) {}

    public function sign(RequestInterface $request): RequestInterface
    {
        if ($this->credentials === null || $this->credentials->isExpired()) {
            $this->credentials = ($this->resolveCredentials)();
        }

        return $this->signature->signRequest($request, $this->credentials);
    }
}
