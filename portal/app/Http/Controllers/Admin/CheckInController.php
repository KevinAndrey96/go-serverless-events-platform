<?php

namespace App\Http\Controllers\Admin;

use App\EventsApi\EventsApi;
use App\EventsApi\EventsApiException;
use App\Http\Controllers\Controller;
use Illuminate\Contracts\View\View;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

final class CheckInController extends Controller
{
    public function __construct(private readonly EventsApi $api) {}

    public function show(): View
    {
        return view('admin.check-in');
    }

    public function store(Request $request): JsonResponse
    {
        $ticket = $request->validate(['ticket' => ['required', 'string', 'max:100']])['ticket'];

        try {
            $attendee = $this->api->checkIn($ticket);
        } catch (EventsApiException $e) {
            if ($e->status >= 500) {
                report($e);
            }

            return response()->json(['admitted' => false, 'message' => $e->userMessage()], $e->status >= 500 ? 503 : 422);
        }

        return response()->json([
            'admitted' => true,
            'message' => "Bienvenido, {$attendee->fullName()}.",
            'email' => $attendee->email,
        ]);
    }
}
