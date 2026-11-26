<?php

namespace App\EventsApi;

use Carbon\CarbonInterface;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\Factory;
use Illuminate\Http\Client\PendingRequest;
use Illuminate\Http\Client\RequestException;
use Illuminate\Http\Client\Response;
use Throwable;

final class HttpEventsApi implements EventsApi
{
    private const RETRIES = 3;

    public function __construct(
        private readonly Factory $http,
        private readonly string $baseUrl,
        private readonly int $timeout,
        private readonly ?RequestSigner $signer = null,
    ) {}

    public function upcomingEvents(): array
    {
        return $this->events($this->get('/events'));
    }

    public function register(string $eventId, array $registration): Attendee
    {
        return Attendee::fromArray($this->send('POST', '/events/'.rawurlencode($eventId).'/attendees', $registration));
    }

    public function allEvents(): array
    {
        return $this->events($this->get('/admin/events', admin: true));
    }

    public function createEvent(string $name, CarbonInterface $startsAt, int $capacity, int $priceCop, ?string $bank, ?string $account): Event
    {
        return Event::fromArray($this->send('POST', '/admin/events', [
            'name' => $name,
            'startsAt' => $startsAt->toIso8601String(),
            'capacity' => $capacity,
            'priceCop' => $priceCop,
            'bank' => $bank ?? '',
            'account' => $account ?? '',
        ], admin: true));
    }

    public function changeEventStatus(string $eventId, string $status): Event
    {
        return Event::fromArray($this->send(
            'PATCH', '/admin/events/'.rawurlencode($eventId).'/status', ['status' => $status], admin: true,
        ));
    }

    public function attendees(string $eventId): array
    {
        $body = $this->get('/admin/events/'.rawurlencode($eventId).'/attendees', admin: true);

        return [
            'event' => Event::fromArray($body['event']),
            'attendees' => array_values(array_map(Attendee::fromArray(...), $body['attendees'] ?? [])),
        ];
    }

    public function changePaymentStatus(string $eventId, string $attendeeId, string $paymentStatus): Attendee
    {
        return Attendee::fromArray($this->send(
            'PATCH',
            '/admin/events/'.rawurlencode($eventId).'/attendees/'.rawurlencode($attendeeId).'/payment',
            ['paymentStatus' => $paymentStatus],
            admin: true,
        ));
    }

    public function checkIn(string $ticket): Attendee
    {
        return Attendee::fromArray($this->send('POST', '/admin/check-ins', ['ticket' => $ticket], admin: true));
    }

    /**
     * @param  array<string, mixed>  $body
     * @return list<Event>
     */
    private function events(array $body): array
    {
        return array_values(array_map(Event::fromArray(...), $body['events'] ?? []));
    }

    /**
     * Reads are idempotent, so connection errors and 5xx answers are retried with backoff.
     *
     * @return array<string, mixed>
     */
    private function get(string $path, bool $admin = false): array
    {
        $request = $this->request($admin)->retry(
            self::RETRIES,
            fn (int $attempt): int => 100 * 2 ** $attempt,
            fn (Throwable $e): bool => $e instanceof ConnectionException
                || ($e instanceof RequestException && $e->response->serverError()),
            throw: false,
        );

        return $this->decode(fn () => $request->get($path));
    }

    /**
     * Writes are sent once: a timed-out registration may have succeeded.
     *
     * @param  array<string, mixed>  $body
     * @return array<string, mixed>
     */
    private function send(string $method, string $path, array $body, bool $admin = false): array
    {
        return $this->decode(fn () => $this->request($admin)->send($method, $path, ['json' => $body]));
    }

    private function request(bool $admin): PendingRequest
    {
        $request = $this->http->baseUrl($this->baseUrl)
            ->acceptJson()
            ->asJson()
            ->connectTimeout(3)
            ->timeout($this->timeout);

        if ($admin && $this->signer !== null) {
            $request = $request->withRequestMiddleware($this->signer->sign(...));
        }

        return $request;
    }

    /**
     * @param  callable(): Response  $call
     * @return array<string, mixed>
     */
    private function decode(callable $call): array
    {
        try {
            $response = $call();
        } catch (ConnectionException $e) {
            throw EventsApiException::unavailable($e);
        }

        if ($response->failed()) {
            throw EventsApiException::fromResponse($response);
        }

        $body = $response->json();

        return is_array($body) ? $body : [];
    }
}
