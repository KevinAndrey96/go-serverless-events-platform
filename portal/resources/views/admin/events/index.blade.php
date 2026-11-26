@extends('layouts.admin', ['title' => 'Eventos'])

@section('content')
    <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-slate-900">Eventos</h1>
        @can('admin')
            <a href="{{ route('admin.events.create') }}" class="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500">Nuevo evento</a>
        @endcan
    </div>

    @if ($error)
        <p class="mt-6 rounded border border-amber-200 bg-amber-50 px-4 py-3 text-amber-800" role="alert">{{ $error }}</p>
    @endif

    <div class="mt-6 overflow-x-auto rounded-lg border border-slate-200 bg-white shadow-sm">
        <table class="min-w-full divide-y divide-slate-200 text-sm">
            <thead class="bg-slate-50 text-left text-slate-600">
                <tr>
                    <th class="px-4 py-3 font-medium">Evento</th>
                    <th class="px-4 py-3 font-medium">Fecha</th>
                    <th class="px-4 py-3 font-medium">Inscritos</th>
                    <th class="px-4 py-3 font-medium">Valor</th>
                    <th class="px-4 py-3 font-medium">Estado</th>
                    <th class="px-4 py-3"></th>
                </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
                @forelse ($events as $event)
                    <tr>
                        <td class="px-4 py-3 font-medium text-slate-900">{{ $event->name }}</td>
                        <td class="px-4 py-3">{{ $event->startsAt->format('d/m/Y g:i a') }}</td>
                        <td class="px-4 py-3">{{ $event->registered }} / {{ $event->capacity }}</td>
                        <td class="px-4 py-3">{{ $event->formattedPrice() }}</td>
                        <td class="px-4 py-3">
                            @if ($event->isCancelled())
                                <span class="rounded bg-rose-100 px-2 py-0.5 text-rose-800">Cancelado</span>
                            @elseif ($event->startsAt->isPast())
                                <span class="rounded bg-slate-200 px-2 py-0.5 text-slate-700">Realizado</span>
                            @else
                                <span class="rounded bg-emerald-100 px-2 py-0.5 text-emerald-800">Abierto</span>
                            @endif
                        </td>
                        <td class="flex justify-end gap-3 px-4 py-3">
                            <a href="{{ route('admin.attendees.index', $event->id) }}" class="text-indigo-700 hover:underline">Asistentes</a>
                            @can('admin')
                                <form method="POST" action="{{ route('admin.events.status', $event->id) }}">
                                    @csrf
                                    @method('PATCH')
                                    <input type="hidden" name="status" value="{{ $event->isCancelled() ? 'enabled' : 'cancelled' }}">
                                    <button type="submit" class="{{ $event->isCancelled() ? 'text-emerald-700' : 'text-rose-700' }} hover:underline">
                                        {{ $event->isCancelled() ? 'Habilitar' : 'Cancelar' }}
                                    </button>
                                </form>
                            @endcan
                        </td>
                    </tr>
                @empty
                    <tr><td colspan="6" class="px-4 py-8 text-center text-slate-500">Aún no hay eventos.</td></tr>
                @endforelse
            </tbody>
        </table>
    </div>
@endsection
