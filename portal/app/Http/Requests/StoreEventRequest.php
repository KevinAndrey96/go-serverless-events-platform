<?php

namespace App\Http\Requests;

use Carbon\CarbonImmutable;
use Illuminate\Foundation\Http\FormRequest;

final class StoreEventRequest extends FormRequest
{
    public const API_FIELDS = [
        'name' => 'name',
        'startsAt' => 'date',
        'capacity' => 'capacity',
        'priceCop' => 'price',
        'account' => 'account',
    ];

    /**
     * @return array<string, mixed>
     */
    public function rules(): array
    {
        return [
            'name' => ['required', 'string', 'max:200'],
            'date' => ['required', 'date_format:Y-m-d'],
            'time' => ['required', 'date_format:H:i'],
            'capacity' => ['required', 'integer', 'min:1', 'max:100000'],
            'price' => ['required', 'integer', 'min:0'],
            'bank' => ['nullable', 'required_unless:price,0', 'string', 'max:100'],
            'account' => ['nullable', 'required_unless:price,0', 'string', 'max:100'],
        ];
    }

    public function startsAt(): CarbonImmutable
    {
        return CarbonImmutable::createFromFormat(
            'Y-m-d H:i',
            $this->string('date').' '.$this->string('time'),
            (string) config('events.timezone'),
        ) ?: CarbonImmutable::now();
    }

    /**
     * @return array<string, string>
     */
    public function attributes(): array
    {
        return [
            'name' => 'nombre',
            'date' => 'fecha',
            'time' => 'hora',
            'capacity' => 'cupos',
            'price' => 'precio',
            'bank' => 'banco',
            'account' => 'cuenta',
        ];
    }
}
