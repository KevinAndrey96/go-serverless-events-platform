@extends('layouts.public', ['title' => 'Ingreso de organizadores'])

@section('content')
    <form method="POST" action="{{ route('login.store') }}" class="mx-auto max-w-sm space-y-5 rounded-lg border border-slate-200 bg-white p-8 shadow-sm">
        @csrf
        <h1 class="text-xl font-bold text-slate-900">Ingreso de organizadores</h1>

        @error('email')
            <p class="rounded border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-800" role="alert">{{ $message }}</p>
        @enderror

        <label class="block text-sm">
            <span class="font-medium">Correo</span>
            <input type="email" name="email" value="{{ old('email') }}" required autofocus autocomplete="username" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Contraseña</span>
            <input type="password" name="password" required autocomplete="current-password" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="flex items-center gap-2 text-sm">
            <input type="checkbox" name="remember" value="1" class="rounded border-slate-300"> Mantener la sesión
        </label>
        <button type="submit" class="w-full rounded bg-indigo-600 px-4 py-2 font-medium text-white hover:bg-indigo-500">Ingresar</button>
    </form>
@endsection
