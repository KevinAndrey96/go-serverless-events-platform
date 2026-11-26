@extends('layouts.public', ['title' => 'Próximos eventos'])

@section('content')
    <h1 class="text-3xl font-bold text-slate-900">Próximos eventos</h1>
    <p class="mt-2 text-slate-600">Elige un evento e inscríbete. Recibirás tu entrada con código QR por correo.</p>

    @if ($error)
        <p class="mt-8 rounded border border-amber-200 bg-amber-50 px-4 py-3 text-amber-800" role="alert">{{ $error }}</p>
    @elseif ($events === [])
        <p class="mt-8 rounded border border-slate-200 bg-white px-4 py-6 text-center text-slate-500">No hay eventos abiertos por ahora.</p>
    @else
        <ul class="mt-8 grid gap-4 sm:grid-cols-2">
            @foreach ($events as $event)
                <li class="flex flex-col rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
                    <h2 class="text-lg font-semibold text-slate-900">{{ $event->name }}</h2>
                    <p class="mt-1 text-sm text-slate-600">{{ $event->startsAt->translatedFormat('l j \d\e F \d\e Y, g:i a') }}</p>
                    <dl class="mt-4 flex gap-6 text-sm">
                        <div><dt class="text-slate-500">Valor</dt><dd class="font-medium">{{ $event->formattedPrice() }}</dd></div>
                        <div><dt class="text-slate-500">Cupos</dt><dd class="font-medium">{{ $event->seatsLeft }}</dd></div>
                    </dl>
                    <div class="mt-5">
                        @if ($event->seatsLeft > 0)
                            <a href="{{ route('registrations.create', $event->id) }}" class="inline-block rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500">Inscribirme</a>
                        @else
                            <span class="inline-block rounded bg-slate-200 px-4 py-2 text-sm text-slate-600">Cupos agotados</span>
                        @endif
                    </div>
                </li>
            @endforeach
        </ul>
    @endif
@endsection
