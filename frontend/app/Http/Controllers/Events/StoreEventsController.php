<?php

namespace App\Http\Controllers\Events;

use App\Http\Controllers\Controller;
use App\Repositories\Contracts\Events\EventRepositoryInterface;
use Illuminate\Http\Request;
use Webpatser\Uuid\Uuid;

class StoreEventsController extends Controller
{
    private EventRepositoryInterface $eventRepository;

    public function __construct(EventRepositoryInterface $eventRepository)
    {
        $this->eventRepository = $eventRepository;
    }

    public function __invoke(Request $request)
    {
        $request->validate([
            'name' => 'required|string|max:255',
            'capacity' => 'required',
            'date' => 'required|date',
            'hour' => 'required',
            'price' => 'required',
        ]);

        $data = [
            "pk" => "Event#".substr(''.Uuid::generate(), 0, 6),
            "sk" => "METADATA#EVENT",
            "capacity" => strval($request->input('capacity')),
            "date" => strval($request->input('date')),
            "hour" => strval($request->input('hour')),
            "name" => strval($request->input('name')),
            "price" => strval($request->input('price')),
            "status" => 'enabled',
            "bank" => strval($request->input('bank')),
            "account" => strval($request->input('account')),
        ];

        $success = $this->eventRepository->createItem($data);

        if (!$success) {
            return redirect()->back()
                ->withInput()
                ->with('error', 'No se pudo registrar el evento. Por favor intente nuevamente.');
        }

        return redirect()->route('events.index')->with('eventRegistered', 'Evento registrado con éxito');
    }
}
