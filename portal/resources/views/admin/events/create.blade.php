@extends('layouts.admin', ['title' => 'Nuevo evento'])

@section('content')
    <h1 class="text-2xl font-bold text-slate-900">Nuevo evento</h1>

    <form method="POST" action="{{ route('admin.events.store') }}" class="mt-6 grid max-w-2xl gap-5 rounded-lg border border-slate-200 bg-white p-6 shadow-sm sm:grid-cols-2">
        @csrf
        <label class="block text-sm sm:col-span-2">
            <span class="font-medium">Nombre</span>
            <input name="name" value="{{ old('name') }}" required maxlength="200" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Fecha</span>
            <input type="date" name="date" value="{{ old('date') }}" required class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Hora <span class="text-slate-400">({{ config('events.timezone') }})</span></span>
            <input type="time" name="time" value="{{ old('time') }}" required class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Cupos</span>
            <input type="number" name="capacity" value="{{ old('capacity') }}" required min="1" max="100000" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Valor en COP <span class="text-slate-400">(0 si es gratis)</span></span>
            <input type="number" name="price" value="{{ old('price', 0) }}" required min="0" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Banco</span>
            <input name="bank" value="{{ old('bank') }}" maxlength="100" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Cuenta para consignar</span>
            <input name="account" value="{{ old('account') }}" maxlength="100" class="mt-1 w-full rounded border-slate-300">
        </label>
        <div class="flex gap-3 sm:col-span-2">
            <button type="submit" class="rounded bg-indigo-600 px-5 py-2 font-medium text-white hover:bg-indigo-500">Crear evento</button>
            <a href="{{ route('admin.events.index') }}" class="rounded px-5 py-2 text-slate-600 hover:bg-slate-100">Cancelar</a>
        </div>
    </form>
@endsection
