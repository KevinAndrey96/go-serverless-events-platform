@extends('layouts.public', ['title' => 'Tu entrada'])

@section('content')
    <div class="mx-auto max-w-md rounded-lg border border-slate-200 bg-white p-8 text-center shadow-sm">
        <p class="text-sm font-medium uppercase tracking-wide text-emerald-700">Inscripción confirmada</p>
        <h1 class="mt-2 text-2xl font-bold text-slate-900">{{ $ticket['event'] }}</h1>
        <p class="mt-1 text-slate-600">{{ \Carbon\CarbonImmutable::parse($ticket['startsAt'])->translatedFormat('l j \d\e F \d\e Y, g:i a') }}</p>

        <div class="mx-auto mt-6 w-64" aria-label="Código QR de la entrada">{!! $qr !!}</div>
        <p class="mt-2 break-all font-mono text-xs text-slate-500">{{ $ticket['ticket'] }}</p>

        <p class="mt-6 text-sm text-slate-600">
            {{ $ticket['name'] }}, enviamos esta entrada a <strong>{{ $ticket['email'] }}</strong>.
            Preséntala en la puerta; cada código sirve para un solo ingreso.
        </p>
    </div>
@endsection
