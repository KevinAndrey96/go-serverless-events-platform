<?php

namespace App\Http\Controllers\Admin;

use App\EventsApi\EventsApi;
use App\EventsApi\EventsApiException;
use App\Http\Controllers\Controller;
use App\Models\Receipt;
use Illuminate\Contracts\View\View;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Storage;
use Illuminate\Validation\Rule;
use Symfony\Component\HttpFoundation\StreamedResponse;

final class AttendeeController extends Controller
{
    public function __construct(private readonly EventsApi $api) {}

    public function index(string $eventId): View
    {
        try {
            ['event' => $event, 'attendees' => $attendees] = $this->api->attendees($eventId);
        } catch (EventsApiException $e) {
            abort($e->status === 404 ? 404 : 503, $e->userMessage());
        }

        usort($attendees, fn ($a, $b) => strcasecmp($a->lastName.$a->firstName, $b->lastName.$b->firstName));

        return view('admin.attendees.index', [
            'event' => $event,
            'attendees' => $attendees,
            'receipts' => Receipt::query()->where('event_id', $eventId)->pluck('attendee_id')->flip(),
        ]);
    }

    public function updatePayment(Request $request, string $eventId, string $attendeeId): RedirectResponse
    {
        $status = $request->validate(['payment_status' => ['required', Rule::in(['pending', 'paid'])]])['payment_status'];

        try {
            $attendee = $this->api->changePaymentStatus($eventId, $attendeeId, $status);
        } catch (EventsApiException $e) {
            return redirect()->route('admin.attendees.index', $eventId)->withErrors(['attendee' => $e->userMessage()]);
        }

        $label = $attendee->isPaid() ? 'confirmado' : 'marcado como pendiente';

        return redirect()->route('admin.attendees.index', $eventId)
            ->with('status', "Pago de {$attendee->fullName()} {$label}.");
    }

    public function receipt(string $eventId, string $attendeeId): StreamedResponse
    {
        $receipt = Receipt::query()
            ->where('event_id', $eventId)
            ->where('attendee_id', $attendeeId)
            ->firstOrFail();

        return Storage::disk('local')->response($receipt->path, null, [
            'Content-Type' => $receipt->mime_type,
            'Content-Disposition' => 'inline',
            'Cache-Control' => 'private, no-store',
            'X-Content-Type-Options' => 'nosniff',
        ]);
    }
}
