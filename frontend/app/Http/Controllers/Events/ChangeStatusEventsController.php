<?php

namespace App\Http\Controllers\Events;

use App\Http\Controllers\Controller;
use App\Repositories\Contracts\Events\EventRepositoryInterface;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Log;

class ChangeStatusEventsController extends Controller
{
    private EventRepositoryInterface $eventRepository;

    public function __construct(EventRepositoryInterface $eventRepository)
    {
        $this->eventRepository = $eventRepository;
    }

    public function __invoke(Request $request)
    {
        try {
            $this->eventRepository->changeStatus(
                $request->input('pk'),
                $request->input('sk'),
                $request->input('status')
            );
            return back()->with('success', 'Estado del evento actualizado');
        } catch (\Throwable $e) {
            Log::error('Error changing event status', ['error' => $e->getMessage()]);
            return back()->with('error', 'No fue posible actualizar el estado del evento');
        }
    }
}
