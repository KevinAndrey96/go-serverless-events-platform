<?php

namespace App\Http\Controllers;

use App\EventsApi\EventsApi;
use App\EventsApi\EventsApiException;
use Illuminate\Contracts\View\View;

final class PublicEventController extends Controller
{
    public function __construct(private readonly EventsApi $api) {}

    public function index(): View
    {
        try {
            $events = $this->api->upcomingEvents();
        } catch (EventsApiException $e) {
            report($e);

            return view('public.events', ['events' => [], 'error' => $e->userMessage()]);
        }

        return view('public.events', ['events' => $events, 'error' => null]);
    }
}
