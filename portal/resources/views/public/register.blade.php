@extends('layouts.public', ['title' => 'Inscripción'])

@section('content')
    <a href="{{ route('events.index') }}" class="text-sm text-indigo-700 hover:underline">← Volver a los eventos</a>
    <h1 class="mt-4 text-3xl font-bold text-slate-900">{{ $event->name }}</h1>
    <p class="mt-1 text-slate-600">{{ $event->startsAt->translatedFormat('l j \d\e F \d\e Y, g:i a') }} · {{ $event->formattedPrice() }} · {{ $event->seatsLeft }} cupos</p>

    @unless ($event->isFree())
        <div class="mt-6 rounded border border-indigo-200 bg-indigo-50 px-4 py-3 text-sm text-indigo-900">
            Consigna {{ $event->formattedPrice() }} en <strong>{{ $event->bank }}</strong>, cuenta <strong>{{ $event->account }}</strong>,
            y adjunta el comprobante. El equipo lo revisará antes del evento; sin pago confirmado no podrás ingresar.
        </div>
    @endunless

    @if ($errors->any())
        <div class="mt-6 rounded border border-rose-200 bg-rose-50 px-4 py-3 text-rose-800" role="alert">
            @foreach ($errors->all() as $message)
                <p>{{ $message }}</p>
            @endforeach
        </div>
    @endif

    <form method="POST" action="{{ route('registrations.store', $event->id) }}" enctype="multipart/form-data" class="mt-8 grid max-w-2xl gap-5 rounded-lg border border-slate-200 bg-white p-6 shadow-sm sm:grid-cols-2">
        @csrf
        <label class="block text-sm">
            <span class="font-medium">Nombres</span>
            <input name="first_name" value="{{ old('first_name') }}" required maxlength="100" autocomplete="given-name" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Apellidos</span>
            <input name="last_name" value="{{ old('last_name') }}" required maxlength="100" autocomplete="family-name" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Correo</span>
            <input type="email" name="email" value="{{ old('email') }}" required maxlength="254" autocomplete="email" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Teléfono <span class="text-slate-400">(opcional)</span></span>
            <input type="tel" name="phone" value="{{ old('phone') }}" maxlength="30" autocomplete="tel" class="mt-1 w-full rounded border-slate-300">
        </label>
        <label class="block text-sm">
            <span class="font-medium">Género <span class="text-slate-400">(opcional)</span></span>
            <select name="gender" class="mt-1 w-full rounded border-slate-300">
                <option value="">Sin especificar</option>
                @foreach (['femenino' => 'Femenino', 'masculino' => 'Masculino', 'otro' => 'Otro', 'prefiero-no-decir' => 'Prefiero no decir'] as $value => $label)
                    <option value="{{ $value }}" @selected(old('gender') === $value)>{{ $label }}</option>
                @endforeach
            </select>
        </label>
        <label class="flex items-center gap-2 self-end text-sm">
            <input type="checkbox" name="scholarship" value="1" @checked(old('scholarship')) class="rounded border-slate-300">
            Tengo beca para este evento
        </label>
        @unless ($event->isFree())
            <label class="block text-sm sm:col-span-2">
                <span class="font-medium">Comprobante de pago <span class="text-slate-400">(JPG, PNG o PDF, máximo 5 MB)</span></span>
                <input type="file" name="receipt" accept=".jpg,.jpeg,.png,.pdf" class="mt-1 block w-full text-sm">
            </label>
        @endunless
        <label class="flex items-start gap-2 text-sm sm:col-span-2">
            <input type="checkbox" name="privacy" value="1" required class="mt-1 rounded border-slate-300">
            Autorizo el tratamiento de mis datos personales para gestionar mi inscripción (Ley 1581 de 2012).
        </label>
        <div class="sm:col-span-2">
            <button type="submit" class="rounded bg-indigo-600 px-5 py-2 font-medium text-white hover:bg-indigo-500">Confirmar inscripción</button>
        </div>
    </form>
@endsection
