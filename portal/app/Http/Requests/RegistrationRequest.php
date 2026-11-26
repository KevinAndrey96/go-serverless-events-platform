<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;

final class RegistrationRequest extends FormRequest
{
    /**
     * API field names mapped to the form fields they come from.
     */
    public const API_FIELDS = [
        'firstName' => 'first_name',
        'lastName' => 'last_name',
        'email' => 'email',
        'phone' => 'phone',
        'gender' => 'gender',
    ];

    /**
     * @return array<string, mixed>
     */
    public function rules(): array
    {
        return [
            'first_name' => ['required', 'string', 'max:100'],
            'last_name' => ['required', 'string', 'max:100'],
            'email' => ['required', 'email:rfc', 'max:254'],
            'phone' => ['nullable', 'string', 'max:30', 'regex:/^[0-9 +()-]+$/'],
            'gender' => ['nullable', Rule::in(['femenino', 'masculino', 'otro', 'prefiero-no-decir'])],
            'scholarship' => ['sometimes', 'boolean'],
            'receipt' => ['nullable', 'file', 'mimes:jpg,jpeg,png,pdf', 'max:'.config('events.receipt_max_kb')],
            'privacy' => ['accepted'],
        ];
    }

    /**
     * @return array{firstName: string, lastName: string, email: string, phone: ?string, gender: ?string, scholarship: bool}
     */
    public function registration(): array
    {
        return [
            'firstName' => $this->string('first_name')->trim()->value(),
            'lastName' => $this->string('last_name')->trim()->value(),
            'email' => $this->string('email')->trim()->lower()->value(),
            'phone' => $this->filled('phone') ? $this->string('phone')->trim()->value() : null,
            'gender' => $this->filled('gender') ? $this->string('gender')->value() : null,
            'scholarship' => $this->boolean('scholarship'),
        ];
    }

    /**
     * @return array<string, string>
     */
    public function attributes(): array
    {
        return [
            'first_name' => 'nombres',
            'last_name' => 'apellidos',
            'email' => 'correo',
            'phone' => 'teléfono',
            'gender' => 'género',
            'receipt' => 'comprobante de pago',
            'privacy' => 'tratamiento de datos',
        ];
    }
}
