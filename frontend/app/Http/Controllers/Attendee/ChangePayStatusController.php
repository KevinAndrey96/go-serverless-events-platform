<?php

namespace App\Http\Controllers\Attendee;

use App\Http\Controllers\Controller;
use App\Repositories\Contracts\Events\EventRepositoryInterface;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Log;

class ChangePayStatusController extends Controller
{
    private EventRepositoryInterface $eventRepository;

    public function __construct(EventRepositoryInterface $eventRepository)
    {
        $this->eventRepository = $eventRepository;
    }

    public function __invoke(Request $request)
    {
        try {
            $this->eventRepository->changePayStatus($request->input('pk'), $request->input('sk'), $request->input('isPayed'));
            return back()->with('success', 'Estado de pago actualizado');
        } catch (\Throwable $e) {
            Log::error('Error changing payment status', ['error' => $e->getMessage()]);
            return back()->with('error', 'No fue posible actualizar el estado de pago');
        }
    }
}
