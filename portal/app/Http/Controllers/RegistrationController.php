<?php

namespace App\Http\Controllers;

use App\EventsApi\Attendee;
use App\EventsApi\Event;
use App\EventsApi\EventsApi;
use App\EventsApi\EventsApiException;
use App\Http\Requests\RegistrationRequest;
use App\Mail\TicketMail;
use App\Models\Receipt;
use App\Support\TicketQr;
use Illuminate\Contracts\View\View;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\UploadedFile;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Mail;
use Illuminate\Validation\ValidationException;
use Symfony\Component\HttpKernel\Exception\HttpException;
use Throwable;

final class RegistrationController extends Controller
{
    public function __construct(private readonly EventsApi $api) {}

    public function create(string $eventId): View
    {
        return view('public.register', ['event' => $this->findOpenEvent($eventId)]);
    }

    public function store(RegistrationRequest $request, string $eventId): RedirectResponse
    {
        $event = $this->findOpenEvent($eventId);

        try {
            $attendee = $this->api->register($event->id, $request->registration());
        } catch (EventsApiException $e) {
            if ($e->isValidation() && $e->field !== null) {
                throw ValidationException::withMessages([
                    RegistrationRequest::API_FIELDS[$e->field] ?? 'email' => 'Revisa este campo.',
                ]);
            }
            if ($e->status >= 500) {
                report($e);
            }

            return redirect()->route('registrations.create', $event->id)
                ->withInput($request->except('receipt'))
                ->withErrors(['registration' => $e->userMessage()]);
        }

        $receipt = $request->file('receipt');
        if ($receipt instanceof UploadedFile) {
            $this->storeReceipt($attendee, $receipt);
        }
        $this->sendTicket($event, $attendee);

        return redirect()->route('tickets.show')->with('ticket', [
            'event' => $event->name,
            'startsAt' => $event->startsAt->toIso8601String(),
            'name' => $attendee->fullName(),
            'email' => $attendee->email,
            'ticket' => $attendee->ticket,
        ]);
    }

    public function ticket(TicketQr $qr): View|RedirectResponse
    {
        $ticket = session('ticket');
        if (! is_array($ticket)) {
            return redirect()->route('events.index');
        }

        return view('public.ticket', ['ticket' => $ticket, 'qr' => $qr->svg((string) $ticket['ticket'])]);
    }

    private function findOpenEvent(string $eventId): Event
    {
        try {
            $events = $this->api->upcomingEvents();
        } catch (EventsApiException $e) {
            report($e);

            throw new HttpException(503, $e->userMessage(), $e);
        }

        foreach ($events as $event) {
            if ($event->id === $eventId) {
                return $event;
            }
        }

        abort(404, 'El evento no existe o ya no recibe inscripciones.');
    }

    private function storeReceipt(Attendee $attendee, UploadedFile $file): void
    {
        $extension = $file->extension() ?: 'bin';
        $path = $file->storeAs("receipts/{$attendee->eventId}", "{$attendee->id}.{$extension}", 'local');
        if ($path === false) {
            Log::error('receipt storage', ['event_id' => $attendee->eventId, 'attendee_id' => $attendee->id]);

            return;
        }

        Receipt::query()->create([
            'event_id' => $attendee->eventId,
            'attendee_id' => $attendee->id,
            'path' => $path,
            'mime_type' => (string) $file->getMimeType(),
        ]);
    }

    /**
     * The registration already exists, so a mail failure must not hide the ticket from the person.
     */
    private function sendTicket(Event $event, Attendee $attendee): void
    {
        try {
            Mail::to($attendee->email)->send(new TicketMail($event, $attendee));
        } catch (Throwable $e) {
            Log::warning('ticket email', ['attendee_id' => $attendee->id, 'error' => $e->getMessage()]);
        }
    }
}
