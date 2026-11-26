<?php

namespace Tests\Unit;

use App\EventsApi\EventsApi;
use App\EventsApi\EventsApiException;
use App\EventsApi\HttpEventsApi;
use App\EventsApi\RequestSigner;
use Aws\Credentials\Credentials;
use Aws\Signature\SignatureV4;
use Illuminate\Http\Client\ConnectionException;
use Illuminate\Http\Client\Factory;
use Illuminate\Http\Client\Request;
use Illuminate\Support\Facades\Http;
use Tests\ApiFixtures;
use Tests\TestCase;

final class HttpEventsApiTest extends TestCase
{
    use ApiFixtures;

    private function api(bool $sign): EventsApi
    {
        $signer = new RequestSigner(
            new SignatureV4('execute-api', 'us-east-1'),
            fn () => new Credentials('AKIDEXAMPLE', 'secret-example'),
        );

        return new HttpEventsApi(app(Factory::class), 'https://api.test', 5, $sign ? $signer : null);
    }

    public function test_signs_only_admin_requests(): void
    {
        Http::fake([
            'api.test/events' => Http::response(['events' => [$this->eventPayload()]]),
            'api.test/admin/events' => Http::response(['events' => [$this->eventPayload()]]),
        ]);
        $api = $this->api(sign: true);

        $this->assertCount(1, $api->upcomingEvents());
        $this->assertSame('Foro de Ingeniería', $api->allEvents()[0]->name);

        Http::assertSent(fn (Request $request) => str_ends_with($request->url(), '/events') && ! str_contains($request->url(), 'admin')
            && ! $request->hasHeader('Authorization'));
        Http::assertSent(fn (Request $request) => str_ends_with($request->url(), '/admin/events')
            && str_starts_with($request->header('Authorization')[0] ?? '', 'AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/')
            && str_contains($request->header('Authorization')[0], '/us-east-1/execute-api/aws4_request')
            && $request->hasHeader('X-Amz-Date'));
    }

    public function test_does_not_sign_without_a_signer(): void
    {
        Http::fake(['api.test/admin/events' => Http::response(['events' => []])]);

        $this->assertSame([], $this->api(sign: false)->allEvents());
        Http::assertSent(fn (Request $request) => ! $request->hasHeader('Authorization'));
    }

    public function test_retries_reads_but_not_writes(): void
    {
        Http::fake([
            'api.test/events' => Http::sequence()->push([], 503)->push(['events' => []]),
            'api.test/events/*/attendees' => Http::response(['error' => 'internal'], 503),
        ]);
        $api = $this->api(sign: false);

        $this->assertSame([], $api->upcomingEvents());
        Http::assertSentCount(2);

        try {
            $api->register(self::EVENT_ID, ['firstName' => 'Ana', 'lastName' => 'P', 'email' => 'a@example.com', 'phone' => null, 'gender' => null, 'scholarship' => false]);
            $this->fail('expected an exception');
        } catch (EventsApiException $e) {
            $this->assertSame(503, $e->status);
        }
        Http::assertSentCount(3);
    }

    public function test_translates_api_errors(): void
    {
        Http::fake(['api.test/admin/check-ins' => Http::response($this->apiError('validation', 'ticket'), 422)]);

        try {
            $this->api(sign: false)->checkIn('x');
            $this->fail('expected an exception');
        } catch (EventsApiException $e) {
            $this->assertTrue($e->isValidation());
            $this->assertSame('ticket', $e->field);
            $this->assertSame('Revisa los datos enviados.', $e->userMessage());
        }
    }

    public function test_connection_errors_become_unavailable(): void
    {
        Http::fake(fn () => throw new ConnectionException('refused'));

        try {
            $this->api(sign: false)->checkIn('x');
            $this->fail('expected an exception');
        } catch (EventsApiException $e) {
            $this->assertSame('unavailable', $e->errorCode);
            $this->assertStringContainsString('no está disponible', $e->userMessage());
        }
    }

    public function test_decodes_attendees(): void
    {
        Http::fake(['api.test/admin/events/*/attendees' => Http::response([
            'event' => $this->eventPayload(),
            'attendees' => [$this->attendeePayload(['paymentStatus' => 'paid', 'checkedInAt' => '2027-01-10T14:00:00Z', 'attendance' => 'attended'])],
        ])]);

        ['event' => $event, 'attendees' => $attendees] = $this->api(sign: false)->attendees(self::EVENT_ID);

        $this->assertSame(8, $event->seatsLeft);
        $this->assertTrue($attendees[0]->isPaid());
        $this->assertTrue($attendees[0]->hasAttended());
        $this->assertSame('09:00', $attendees[0]->checkedInAt?->format('H:i'));
    }
}
