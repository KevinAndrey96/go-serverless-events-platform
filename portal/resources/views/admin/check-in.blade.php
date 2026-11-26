@extends('layouts.admin', ['title' => 'Control de ingreso'])

@push('scripts')
    @vite('resources/js/check-in.js')
@endpush

@section('content')
    <h1 class="text-2xl font-bold text-slate-900">Control de ingreso</h1>
    <p class="mt-1 text-slate-600">Escanea la entrada con la cámara o escribe el código impreso debajo del QR.</p>

    <div class="mt-6 grid gap-6 lg:grid-cols-2">
        <div id="scanner" class="overflow-hidden rounded-lg border border-slate-200 bg-white"></div>

        <div class="space-y-4">
            <form id="ticket-form" data-url="{{ route('admin.check-in.store') }}" class="flex gap-2">
                <input id="ticket" name="ticket" required maxlength="100" placeholder="Código de la entrada" class="w-full rounded border-slate-300 font-mono text-sm">
                <button type="submit" class="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500">Validar</button>
            </form>
            <div id="result" class="hidden rounded-lg px-5 py-6 text-center text-lg font-semibold" role="status" aria-live="assertive"></div>
        </div>
    </div>
@endsection
