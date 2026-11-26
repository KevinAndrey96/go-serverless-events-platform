<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="csrf-token" content="{{ csrf_token() }}">
    <title>{{ $title ?? 'Panel' }} · {{ config('app.name') }}</title>
    @vite(['resources/css/app.css', 'resources/js/app.js'])
    @stack('scripts')
</head>
<body class="min-h-screen bg-slate-100 text-slate-800 antialiased">
    <header class="bg-indigo-800 text-white">
        <nav class="mx-auto flex max-w-6xl flex-wrap items-center gap-x-6 gap-y-2 px-4 py-3 text-sm">
            <span class="font-semibold">{{ config('app.name') }}</span>
            <a href="{{ route('admin.events.index') }}" class="hover:underline">Eventos</a>
            <a href="{{ route('admin.check-in.show') }}" class="hover:underline">Control de ingreso</a>
            @can('admin')
                <a href="{{ route('admin.staff.index') }}" class="hover:underline">Equipo</a>
            @endcan
            <form method="POST" action="{{ route('admin.logout') }}" class="ml-auto flex items-center gap-3">
                @csrf
                <span class="text-indigo-200">{{ auth()->user()->name }} · {{ auth()->user()->role->label() }}</span>
                <button type="submit" class="rounded bg-indigo-700 px-3 py-1 hover:bg-indigo-600">Salir</button>
            </form>
        </nav>
    </header>

    <main class="mx-auto max-w-6xl px-4 py-8">
        @if (session('status'))
            <p class="mb-6 rounded border border-emerald-200 bg-emerald-50 px-4 py-3 text-emerald-800" role="status">{{ session('status') }}</p>
        @endif
        @if ($errors->any())
            <div class="mb-6 rounded border border-rose-200 bg-rose-50 px-4 py-3 text-rose-800" role="alert">
                @foreach ($errors->all() as $message)
                    <p>{{ $message }}</p>
                @endforeach
            </div>
        @endif

        @yield('content')
    </main>
</body>
</html>
