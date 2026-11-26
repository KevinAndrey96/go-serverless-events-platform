<?php

use App\Http\Controllers\Admin\AttendeeController;
use App\Http\Controllers\Admin\CheckInController;
use App\Http\Controllers\Admin\EventController;
use App\Http\Controllers\Admin\StaffController;
use App\Http\Controllers\Auth\LoginController;
use App\Http\Controllers\PublicEventController;
use App\Http\Controllers\RegistrationController;
use Illuminate\Support\Facades\Route;

Route::pattern('eventId', '[a-z2-7]{16}');
Route::pattern('attendeeId', '[A-Za-z0-9_-]{22}');

Route::get('/', [PublicEventController::class, 'index'])->name('events.index');
Route::get('/eventos/{eventId}/inscripcion', [RegistrationController::class, 'create'])->name('registrations.create');
Route::post('/eventos/{eventId}/inscripcion', [RegistrationController::class, 'store'])
    ->middleware('throttle:registrations')
    ->name('registrations.store');
Route::get('/entrada', [RegistrationController::class, 'ticket'])->name('tickets.show');

Route::middleware('guest')->group(function () {
    Route::get('/ingresar', [LoginController::class, 'create'])->name('login');
    Route::post('/ingresar', [LoginController::class, 'store'])->name('login.store');
});

Route::middleware('auth')->prefix('admin')->name('admin.')->group(function () {
    Route::post('/salir', [LoginController::class, 'destroy'])->name('logout');

    Route::get('/eventos', [EventController::class, 'index'])->name('events.index');
    Route::get('/eventos/{eventId}/asistentes', [AttendeeController::class, 'index'])->name('attendees.index');
    Route::get('/eventos/{eventId}/asistentes/{attendeeId}/comprobante', [AttendeeController::class, 'receipt'])
        ->name('attendees.receipt');

    Route::get('/ingreso', [CheckInController::class, 'show'])->name('check-in.show');
    Route::post('/ingreso', [CheckInController::class, 'store'])->middleware('throttle:check-ins')->name('check-in.store');

    Route::middleware('can:admin')->group(function () {
        Route::get('/eventos/crear', [EventController::class, 'create'])->name('events.create');
        Route::post('/eventos', [EventController::class, 'store'])->name('events.store');
        Route::patch('/eventos/{eventId}/estado', [EventController::class, 'updateStatus'])->name('events.status');
        Route::patch('/eventos/{eventId}/asistentes/{attendeeId}/pago', [AttendeeController::class, 'updatePayment'])
            ->name('attendees.payment');

        Route::get('/equipo', [StaffController::class, 'index'])->name('staff.index');
        Route::get('/equipo/crear', [StaffController::class, 'create'])->name('staff.create');
        Route::post('/equipo', [StaffController::class, 'store'])->name('staff.store');
    });
});
