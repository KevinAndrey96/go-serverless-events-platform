<?php

return [
    'accepted' => 'Debes aceptar :attribute.',
    'boolean' => 'El campo :attribute debe ser verdadero o falso.',
    'confirmed' => 'La confirmación de :attribute no coincide.',
    'date_format' => 'El campo :attribute no tiene el formato :format.',
    'email' => 'El campo :attribute debe ser un correo válido.',
    'enum' => 'El valor de :attribute no es válido.',
    'file' => 'El campo :attribute debe ser un archivo.',
    'in' => 'El valor de :attribute no es válido.',
    'integer' => 'El campo :attribute debe ser un número entero.',
    'max' => [
        'file' => 'El campo :attribute no debe pesar más de :max kilobytes.',
        'numeric' => 'El campo :attribute no debe ser mayor que :max.',
        'string' => 'El campo :attribute no debe tener más de :max caracteres.',
    ],
    'mimes' => 'El campo :attribute debe ser un archivo de tipo: :values.',
    'min' => [
        'numeric' => 'El campo :attribute debe ser al menos :min.',
        'string' => 'El campo :attribute debe tener al menos :min caracteres.',
    ],
    'password' => [
        'letters' => 'El campo :attribute debe contener al menos una letra.',
        'numbers' => 'El campo :attribute debe contener al menos un número.',
    ],
    'regex' => 'El formato de :attribute no es válido.',
    'required' => 'El campo :attribute es obligatorio.',
    'required_unless' => 'El campo :attribute es obligatorio cuando el evento tiene costo.',
    'string' => 'El campo :attribute debe ser texto.',
    'unique' => 'Ya existe una cuenta con ese :attribute.',

    'attributes' => [
        'email' => 'correo',
        'password' => 'contraseña',
        'name' => 'nombre',
        'role' => 'rol',
        'status' => 'estado',
        'payment_status' => 'estado del pago',
        'ticket' => 'código',
    ],
];
