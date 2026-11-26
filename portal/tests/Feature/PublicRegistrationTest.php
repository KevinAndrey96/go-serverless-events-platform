<?php

namespace Tests\Feature;

use App\EventsApi\Attendee;
use App\EventsApi\Event;
use App\Mail\TicketMail;
use App\Models\Receipt;
use App\Support\TicketQr;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Http\Client\Request;
use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Mail;
use Illuminate\Support\Facades\Storage;
use PHPUnit\Framework\Attributes\DataProvider;
use Tests\ApiFixtures;
use Tests\TestCase;

final class PublicRegistrationTest extends TestCase
{
    use ApiFixtures;
    use RefreshDatabase;

    /**
     * @param  array<string, mixed>|null  $body
     */
    private function fakeApi(int $status = 201, ?array $body = null): void
    {
        Http::fake([
            'api.test/events' => Http::response(['events' => [$this->eventPayload()]]),
            'api.test/events/*/attendees' => Http::response($body ?? $this->attendeePayload(), $status),
        ]);
    }

    /**
     * @param  array<string, mixed>  $overrides
     * @return array<string, mixed>
     */
    private function form(array $overrides = []): array
    {
        return array_merge([
            'first_name' => ' Ana ',
            'last_name' => 'Pérez',
            'email' => 'ANA@example.com',
            'phone' => '+57 300 000 0000',
            'gender' => 'femenino',
            'privacy' => '1',
        ], $overrides);
    }

    public function test_lists_upcoming_events(): void
    {
        $this->fakeApi();

        $this->get('/')
            ->assertOk()
            ->assertSee('Foro de Ingeniería')
            ->assertSee('$ 30.000 COP')
            ->assertSee(route('registrations.create', self::EVENT_ID));
    }

    public function test_shows_a_friendly_message_when_the_api_is_down(): void
    {
        Http::fake(['api.test/*' => Http::response(['error' => 'internal'], 500)]);

        $this->get('/')->assertOk()->assertSee('no está disponible');
        Http::assertSentCount(3);
    }

    public function test_unknown_events_are_not_found(): void
    {
        $this->fakeApi();

        $this->get('/eventos/zzzzzzzzzzzzzzzz/inscripcion')->assertNotFound();
    }

    public function test_registers_stores_the_receipt_privately_and_emails_the_ticket(): void
    {
        $this->fakeApi();
        Mail::fake();
        Storage::fake('local');

        $response = $this->post(route('registrations.store', self::EVENT_ID), $this->form([
            'scholarship' => '1',
            'receipt' => UploadedFile::fake()->image('pago.png'),
        ]));

        $response->assertRedirect(route('tickets.show'));
        Http::assertSent(fn (Request $request) => $request->method() === 'POST'
            && $request->url() === 'http://api.test/events/'.self::EVENT_ID.'/attendees'
            && $request->data() === [
                'firstName' => 'Ana',
                'lastName' => 'Pérez',
                'email' => 'ana@example.com',
                'phone' => '+57 300 000 0000',
                'gender' => 'femenino',
                'scholarship' => true,
            ]
            && ! $request->hasHeader('Authorization'));

        $receipt = Receipt::query()->sole();
        $this->assertSame(self::ATTENDEE_ID, $receipt->attendee_id);
        Storage::disk('local')->assertExists($receipt->path);
        $this->assertStringStartsWith('receipts/'.self::EVENT_ID.'/', $receipt->path);

        Mail::assertSent(TicketMail::class, fn (TicketMail $mail) => $mail->hasTo('ana@example.com')
            && $mail->attendee->ticket === self::EVENT_ID.'.'.self::ATTENDEE_ID);

        $this->get(route('tickets.show'))
            ->assertOk()
            ->assertSee('Inscripción confirmada')
            ->assertSee(self::EVENT_ID.'.'.self::ATTENDEE_ID)
            ->assertSee('<svg', false);
    }

    public function test_ticket_mail_carries_the_code_and_a_png_qr(): void
    {
        $ticket = self::EVENT_ID.'.'.self::ATTENDEE_ID;
        $mail = new TicketMail(Event::fromArray($this->eventPayload()), Attendee::fromArray($this->attendeePayload()));

        $mail->assertSeeInHtml($ticket);
        $mail->assertSeeInHtml('comprobante de pago');
        $png = app(TicketQr::class)->png($ticket);
        $this->assertStringStartsWith("\x89PNG", $png);
        $mail->assertHasAttachedData($png, 'entrada-qr.png', ['mime' => 'image/png']);
    }

    public function test_the_ticket_page_needs_a_fresh_registration(): void
    {
        $this->get(route('tickets.show'))->assertRedirect(route('events.index'));
    }

    /**
     * @return array<string, array{string, string}>
     */
    public static function apiConflicts(): array
    {
        return [
            'full' => ['event_full', 'ya no tiene cupos'],
            'closed' => ['event_closed', 'no está recibiendo inscripciones'],
            'duplicate' => ['already_registered', 'ya está inscrito'],
        ];
    }

    #[DataProvider('apiConflicts')]
    public function test_shows_api_conflicts(string $code, string $message): void
    {
        $this->fakeApi(409, $this->apiError($code));
        Mail::fake();

        $this->from(route('tickets.show'))
            ->post(route('registrations.store', self::EVENT_ID), $this->form())
            ->assertRedirect(route('registrations.create', self::EVENT_ID))
            ->assertSessionHasErrors('registration');

        $this->assertStringContainsString($message, session('errors')->first('registration'));
        Mail::assertNothingSent();
        $this->assertSame(0, Receipt::query()->count());
    }

    public function test_maps_api_validation_errors_to_form_fields(): void
    {
        $this->fakeApi(422, $this->apiError('validation', 'email'));

        $this->post(route('registrations.store', self::EVENT_ID), $this->form())
            ->assertSessionHasErrors('email');
    }

    public function test_validates_the_form_before_calling_the_api(): void
    {
        $this->fakeApi();

        $this->post(route('registrations.store', self::EVENT_ID), $this->form([
            'email' => 'no-es-correo',
            'privacy' => null,
            'receipt' => UploadedFile::fake()->create('malware.exe', 10),
        ]))->assertSessionHasErrors(['email', 'privacy', 'receipt']);

        Http::assertNotSent(fn (Request $request) => $request->method() === 'POST');
    }

    public function test_registrations_are_rate_limited(): void
    {
        $this->fakeApi(409, $this->apiError('already_registered'));

        foreach (range(1, 10) as $attempt) {
            $this->post(route('registrations.store', self::EVENT_ID), $this->form())->assertRedirect();
        }

        $this->post(route('registrations.store', self::EVENT_ID), $this->form())->assertTooManyRequests();
    }

    public function test_sets_security_headers(): void
    {
        $this->fakeApi();

        $this->get('/')
            ->assertHeader('X-Frame-Options', 'DENY')
            ->assertHeader('X-Content-Type-Options', 'nosniff');
    }
}
