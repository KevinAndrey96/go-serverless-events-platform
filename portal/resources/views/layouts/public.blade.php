<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{ $title ?? 'Eventos' }} · {{ config('app.name') }}</title>
    @vite(['resources/css/app.css', 'resources/js/app.js'])
</head>
<body class="min-h-screen bg-slate-50 text-slate-800 antialiased">
    <header class="border-b border-slate-200 bg-white">
        <div class="mx-auto flex max-w-5xl items-center justify-between px-4 py-4">
            <a href="{{ route('events.index') }}" class="text-lg font-semibold text-indigo-700">{{ config('app.name') }}</a>
            <a href="{{ route('login') }}" class="text-sm text-slate-500 hover:text-slate-800">Organizadores</a>
        </div>
    </header>

    <main class="mx-auto max-w-5xl px-4 py-10">
        @yield('content')
    </main>
</body>
</html>
