@extends('layouts.admin', ['title' => 'Asistentes'])

@section('content')
    <a href="{{ route('admin.events.index') }}" class="text-sm text-indigo-700 hover:underline">← Eventos</a>
    <h1 class="mt-2 text-2xl font-bold text-slate-900">{{ $event->name }}</h1>
    <p class="text-slate-600">{{ $event->startsAt->format('d/m/Y g:i a') }} · {{ $event->registered }} de {{ $event->capacity }} cupos</p>

    <div class="mt-6 overflow-x-auto rounded-lg border border-slate-200 bg-white shadow-sm">
        <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50 text-left text-slate-600">
                <tr>
                    <th class="px-4 py-3 font-medium">Asistente</th>
                    <th class="px-4 py-3 font-medium">Contacto</th>
                    <th class="px-4 py-3 font-medium">Inscripción</th>
                    <th class="px-4 py-3 font-medium">Pago</th>
                    <th class="px-4 py-3 font-medium">Ingreso</th>
                </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
                @forelse ($attendees as $attendee)
                    <tr>
                        <td class="px-4 py-3">
                            <p class="font-medium text-slate-900">{{ $attendee->fullName() }}</p>
                            @if ($attendee->scholarship)<p class="text-xs text-indigo-700">Becado</p>@endif
                        </td>
                        <td class="px-4 py-3">
                            <p>{{ $attendee->email }}</p>
                            <p class="text-slate-500">{{ $attendee->phone }}</p>
                        </td>
                        <td class="px-4 py-3">{{ $attendee->registeredAt->format('d/m/Y g:i a') }}</td>
                        <td class="px-4 py-3">
                            <div class="flex flex-wrap items-center gap-3">
                                <span class="rounded px-2 py-0.5 {{ $attendee->isPaid() ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800' }}">
                                    {{ $attendee->isPaid() ? 'Confirmado' : 'Pendiente' }}
                                </span>
                                @if ($receipts->has($attendee->id))
                                    <a href="{{ route('admin.attendees.receipt', [$event->id, $attendee->id]) }}" target="_blank" rel="noopener" class="text-indigo-700 hover:underline">Comprobante</a>
                                @endif
                                @can('admin')
                                    <form method="POST" action="{{ route('admin.attendees.payment', [$event->id, $attendee->id]) }}">
                                        @csrf
                                        @method('PATCH')
                                        <input type="hidden" name="payment_status" value="{{ $attendee->isPaid() ? 'pending' : 'paid' }}">
                                        <button type="submit" class="text-slate-600 hover:underline">{{ $attendee->isPaid() ? 'Revertir' : 'Confirmar pago' }}</button>
                                    </form>
                                @endcan
                            </div>
                        </td>
                        <td class="px-4 py-3">
                            @if ($attendee->hasAttended())
                                <span class="text-emerald-700">{{ $attendee->checkedInAt?->format('d/m g:i a') }}</span>
                            @else
                                <span class="text-slate-400">No ha ingresado</span>
                            @endif
                        </td>
                    </tr>
                @empty
                    <tr><td colspan="5" class="px-4 py-8 text-center text-slate-500">Aún no hay inscritos.</td></tr>
                @endforelse
            </tbody>
        </table>
    </div>
@endsection
