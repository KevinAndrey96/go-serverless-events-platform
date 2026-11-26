@extends('layouts.admin', ['title' => 'Equipo'])

@section('content')
    <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-slate-900">Equipo</h1>
        <a href="{{ route('admin.staff.create') }}" class="rounded bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-500">Nueva cuenta</a>
    </div>

    <ul class="mt-6 divide-y divide-slate-100 rounded-lg border border-slate-200 bg-white shadow-sm">
        @foreach ($users as $user)
            <li class="flex items-center justify-between px-4 py-3 text-sm">
                <div>
                    <p class="font-medium text-slate-900">{{ $user->name }}</p>
                    <p class="text-slate-500">{{ $user->email }}</p>
                </div>
                <span class="rounded bg-slate-100 px-2 py-0.5 text-slate-700">{{ $user->role->label() }}</span>
            </li>
        @endforeach
    </ul>
@endsection
