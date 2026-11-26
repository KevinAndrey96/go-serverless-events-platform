@extends('layouts.admin', ['title' => 'Nueva cuenta'])

@section('content')
    <h1 class="text-2xl font-bold text-slate-900">Nueva cuenta del equipo</h1>

    <form method="POST" action="{{ route('admin.staff.store') }}" class="mt-6 grid max-w-xl gap-5 rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
        @csrf
        <label class="block text-sm">
            <span class="font-medium">Nombre</span>
            <input name="name" value="{{ old('name') }}" required maxlength="100" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Correo</span>
            <input type="email" name="email" value="{{ old('email') }}" required maxlength="254" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Rol</span>
            <select name="role" class="mt-1 w-full rounded border-slate-300">
                @foreach ($roles as $role)
                    <option value="{{ $role->value }}" @selected(old('role', 'staff') === $role->value)>{{ $role->label() }}</option>
                @endforeach
            </select>
        </label>
        <label class="block text-sm">
            <span class="font-medium">Contraseña <span class="text-slate-400">(mínimo 12 caracteres, letras y números)</span></span>
            <input type="password" name="password" required minlength="12" autocomplete="new-password" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Confirmar contraseña</span>
            <input type="password" name="password_confirmation" required minlength="12" autocomplete="new-password" class="mt-1 w-full rounded border-slate-300">
        </label>
        <div class="flex gap-3">
            <button type="submit" class="rounded bg-indigo-600 px-5 py-2 font-medium text-white hover:bg-indigo-500">Crear cuenta</button>
            <a href="{{ route('admin.staff.index') }}" class="rounded px-5 py-2 text-slate-600 hover:bg-slate-100">Cancelar</a>
        </div>
    </form>
@endsection
