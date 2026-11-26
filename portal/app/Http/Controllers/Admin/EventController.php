<?php

namespace App\Http\Controllers\Admin;

use App\EventsApi\EventsApi;
use App\EventsApi\EventsApiException;
use App\Http\Controllers\Controller;
use App\Http\Requests\StoreEventRequest;
use Illuminate\Contracts\View\View;
use Illuminate\Http\RedirectResponse;
use Illuminate\Http\Request;
use Illuminate\Validation\Rule;
use Illuminate\Validation\ValidationException;

final class EventController extends Controller
{
    public function __construct(private readonly EventsApi $api) {}

    public function index(): View
    {
        try {
            $events = $this->api->allEvents();
        } catch (EventsApiException $e) {
            report($e);

            return view('admin.events.index', ['events' => [], 'error' => $e->userMessage()]);
        }

        return view('admin.events.index', ['events' => $events, 'error' => null]);
    }

    public function create(): View
    {
        return view('admin.events.create');
    }

    public function store(StoreEventRequest $request): RedirectResponse
    {
        try {
            $event = $this->api->createEvent(
                name: $request->string('name')->trim()->value(),
                startsAt: $request->startsAt(),
                capacity: $request->integer('capacity'),
                priceCop: $request->integer('price'),
                bank: $request->filled('bank') ? $request->string('bank')->trim()->value() : null,
                account: $request->filled('account') ? $request->string('account')->trim()->value() : null,
            );
        } catch (EventsApiException $e) {
            if ($e->isValidation() && $e->field !== null) {
                throw ValidationException::withMessages([
                    StoreEventRequest::API_FIELDS[$e->field] ?? 'name' => 'Revisa este campo: '.$e->getMessage().'.',
                ]);
            }
            report($e);

            return redirect()->route('admin.events.create')->withInput()->withErrors(['event' => $e->userMessage()]);
        }

        return redirect()->route('admin.events.index')->with('status', "Evento «{$event->name}» creado.");
    }

    public function updateStatus(Request $request, string $eventId): RedirectResponse
    {
        $status = $request->validate(['status' => ['required', Rule::in(['enabled', 'cancelled'])]])['status'];

        try {
            $event = $this->api->changeEventStatus($eventId, $status);
        } catch (EventsApiException $e) {
            return redirect()->route('admin.events.index')->withErrors(['event' => $e->userMessage()]);
        }

        $label = $event->isCancelled() ? 'cancelado' : 'habilitado';

        return redirect()->route('admin.events.index')->with('status', "Evento «{$event->name}» {$label}.");
    }
}
