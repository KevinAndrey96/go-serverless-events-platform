<?php

namespace Tests\Feature;

use App\Models\Receipt;
use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Http\Client\Request;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Storage;
use PHPUnit\Framework\Attributes\DataProvider;
use Tests\ApiFixtures;
use Tests\TestCase;

final class AdminTest extends TestCase
{
    use ApiFixtures;
    use RefreshDatabase;

    /**
     * @return array<string, array{string, string}>
     */
    public static function protectedPages(): array
    {
        return [
            'events' => ['get', '/admin/eventos'],
            'attendees' => ['get', '/admin/eventos/abcdefghijklmnop/asistentes'],
            'check-in' => ['get', '/admin/ingreso'],
            'create event' => ['post', '/admin/eventos'],
            'staff' => ['get', '/admin/equipo'],
        ];
    }

    #[DataProvider('protectedPages')]
    public function test_guests_are_sent_to_the_login(string $method, string $uri): void
    {
        $this->call(strtoupper($method), $uri)->assertRedirect(route('login'));
    }

    /**
     * @return array<string, array{string, string, array<string, mixed>}>
     */
    public static function adminOnlyActions(): array
    {
        return [
            'event form' => ['get', '/admin/eventos/crear', []],
            'create event' => ['post', '/admin/eventos', []],
            'cancel event' => ['patch', '/admin/eventos/abcdefghijklmnop/estado', ['status' => 'cancelled']],
            'confirm payment' => ['patch', '/admin/eventos/abcdefghijklmnop/asistentes/AAAAAAAAAAAAAAAAAAAAAA/pago', ['payment_status' => 'paid']],
            'staff' => ['get', '/admin/equipo', []],
            'create staff' => ['post', '/admin/equipo', []],
        ];
    }

    /**
     * @param  array<string, mixed>  $data
     */
    #[DataProvider('adminOnlyActions')]
    public function test_staff_cannot_run_admin_actions(string $method, string $uri, array $data): void
    {
        Http::fake();

        $this->actingAs(User::factory()->create())->call(strtoupper($method), $uri, $data)->assertForbidden();
        Http::assertNothingSent();
    }

    public function test_lists_every_event(): void
    {
        Http::fake(['api.test/admin/events' => Http::response(['events' => [
            $this->eventPayload(),
            $this->eventPayload(['id' => 'qrstuvwxyz234567', 'name' => 'Taller cancelado', 'status' => 'cancelled']),
        ]])]);

        $this->actingAs(User::factory()->create())
            ->get(route('admin.events.index'))
            ->assertOk()
            ->assertSee('Foro de Ingeniería')
            ->assertSee('2 / 10')
            ->assertSee('Cancelado')
            ->assertDontSee('Nuevo evento');
    }

    public function test_admins_create_events_in_the_organizer_time_zone(): void
    {
        Http::fake(['api.test/admin/events' => Http::response($this->eventPayload(), 201)]);

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('admin.events.store'), [
                'name' => 'Foro de Ingeniería',
                'date' => '2027-03-15',
                'time' => '18:30',
                'capacity' => 120,
                'price' => 30000,
                'bank' => 'Bancolombia',
                'account' => '123-456',
            ])
            ->assertRedirect(route('admin.events.index'))
            ->assertSessionHas('status');

        Http::assertSent(fn (Request $request) => $request->method() === 'POST'
            && $request['startsAt'] === '2027-03-15T18:30:00-05:00'
            && $request['capacity'] === 120
            && $request['priceCop'] === 30000);
    }

    public function test_paid_events_need_bank_details(): void
    {
        Http::fake();

        $this->actingAs(User::factory()->admin()->create())
            ->post(route('admin.events.store'), ['name' => 'Foro', 'date' => '2027-03-15', 'time' => '18:30', 'capacity' => 10, 'price' => 5000])
            ->assertSessionHasErrors(['bank', 'account']);
        Http::assertNothingSent();
    }

    public function test_admins_cancel_events(): void
    {
        Http::fake(['api.test/admin/events/*/status' => Http::response($this->eventPayload(['status' => 'cancelled']))]);

        $this->actingAs(User::factory()->admin()->create())
            ->patch(route('admin.events.status', self::EVENT_ID), ['status' => 'cancelled'])
            ->assertRedirect(route('admin.events.index'))
            ->assertSessionHas('status', 'Evento «Foro de Ingeniería» cancelado.');

        Http::assertSent(fn (Request $request) => $request->method() === 'PATCH' && $request['status'] === 'cancelled');
    }

    public function test_lists_attendees_with_their_receipts(): void
    {
        Http::fake(['api.test/admin/events/*/attendees' => Http::response([
            'event' => $this->eventPayload(),
            'attendees' => [
                $this->attendeePayload(),
                $this->attendeePayload(['id' => 'BBBBBBBBBBBBBBBBBBBBBB', 'firstName' => 'Luis', 'lastName' => 'Arias', 'paymentStatus' => 'paid', 'attendance' => 'attended', 'checkedInAt' => now()->toIso8601String()]),
            ],
        ])]);
        Receipt::query()->create(['event_id' => self::EVENT_ID, 'attendee_id' => self::ATTENDEE_ID, 'path' => 'receipts/x.png', 'mime_type' => 'image/png']);

        $this->actingAs(User::factory()->create())
            ->get(route('admin.attendees.index', self::EVENT_ID))
            ->assertOk()
            ->assertSeeInOrder(['Luis Arias', 'Ana Pérez'])
            ->assertSee(route('admin.attendees.receipt', [self::EVENT_ID, self::ATTENDEE_ID]))
            ->assertDontSee('Confirmar pago');
    }

    public function test_unknown_events_have_no_attendee_page(): void
    {
        Http::fake(['api.test/*' => Http::response($this->apiError('event_not_found'), 404)]);

        $this->actingAs(User::factory()->create())->get(route('admin.attendees.index', self::EVENT_ID))->assertNotFound();
    }

    public function test_admins_confirm_payments(): void
    {
        Http::fake(['api.test/admin/events/*/attendees/*/payment' => Http::response($this->attendeePayload(['paymentStatus' => 'paid']))]);

        $this->actingAs(User::factory()->admin()->create())
            ->patch(route('admin.attendees.payment', [self::EVENT_ID, self::ATTENDEE_ID]), ['payment_status' => 'paid'])
            ->assertRedirect(route('admin.attendees.index', self::EVENT_ID))
            ->assertSessionHas('status', 'Pago de Ana Pérez confirmado.');
    }

    public function test_receipts_are_served_only_to_signed_in_staff(): void
    {
        Storage::fake('local');
        Storage::disk('local')->put('receipts/'.self::EVENT_ID.'/'.self::ATTENDEE_ID.'.png', 'png-bytes');
        Receipt::query()->create([
            'event_id' => self::EVENT_ID,
            'attendee_id' => self::ATTENDEE_ID,
            'path' => 'receipts/'.self::EVENT_ID.'/'.self::ATTENDEE_ID.'.png',
            'mime_type' => 'image/png',
        ]);
        $url = route('admin.attendees.receipt', [self::EVENT_ID, self::ATTENDEE_ID]);

        $this->get($url)->assertRedirect(route('login'));
        $this->actingAs(User::factory()->create())->get($url)
            ->assertOk()
            ->assertHeader('Content-Type', 'image/png')
            ->assertHeader('X-Content-Type-Options', 'nosniff');
        $this->get(route('admin.attendees.receipt', ['zzzzzzzzzzzzzzzz', self::ATTENDEE_ID]))->assertNotFound();
    }

    public function test_check_in_admits_paid_tickets(): void
    {
        Http::fake(['api.test/admin/check-ins' => Http::response($this->attendeePayload(['paymentStatus' => 'paid', 'attendance' => 'attended']))]);

        $this->actingAs(User::factory()->create())
            ->postJson(route('admin.check-in.store'), ['ticket' => self::EVENT_ID.'.'.self::ATTENDEE_ID])
            ->assertOk()
            ->assertJson(['admitted' => true, 'message' => 'Bienvenido, Ana Pérez.']);

        Http::assertSent(fn (Request $request) => $request['ticket'] === self::EVENT_ID.'.'.self::ATTENDEE_ID);
    }

    /**
     * @return array<string, array{string, int, int, string}>
     */
    public static function refusedTickets(): array
    {
        return [
            'used' => ['already_checked_in', 409, 422, 'ya fue usado'],
            'unpaid' => ['payment_pending', 409, 422, 'aún no está confirmado'],
            'unknown' => ['invalid_ticket', 404, 422, 'no corresponde'],
            'api down' => ['internal', 500, 503, 'no está disponible'],
        ];
    }

    #[DataProvider('refusedTickets')]
    public function test_check_in_explains_refusals(string $code, int $apiStatus, int $status, string $message): void
    {
        Http::fake(['api.test/admin/check-ins' => Http::response($this->apiError($code), $apiStatus)]);

        $response = $this->actingAs(User::factory()->create())
            ->postJson(route('admin.check-in.store'), ['ticket' => 'x.y'])
            ->assertStatus($status)
            ->assertJson(['admitted' => false]);

        $this->assertStringContainsString($message, (string) $response->json('message'));
    }

    public function test_check_in_page_loads_the_scanner(): void
    {
        $this->actingAs(User::factory()->create())
            ->get(route('admin.check-in.show'))
            ->assertOk()
            ->assertSee('ticket-form', false);
    }
}
