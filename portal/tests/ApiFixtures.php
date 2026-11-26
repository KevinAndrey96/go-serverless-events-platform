<?php

namespace Tests;

trait ApiFixtures
{
    protected const EVENT_ID = 'abcdefghijklmnop';

    protected const ATTENDEE_ID = 'AAAAAAAAAAAAAAAAAAAAAA';

    /**
     * @param  array<string, mixed>  $overrides
     * @return array<string, mixed>
     */
    protected function eventPayload(array $overrides = []): array
    {
        return array_merge([
            'id' => self::EVENT_ID,
            'name' => 'Foro de Ingeniería',
            'startsAt' => now()->addDays(3)->toIso8601String(),
            'seatsLeft' => 8,
            'priceCop' => 30000,
            'bank' => 'Bancolombia',
            'account' => '123-456',
            'capacity' => 10,
            'registered' => 2,
            'status' => 'enabled',
            'createdAt' => now()->toIso8601String(),
        ], $overrides);
    }

    /**
     * @param  array<string, mixed>  $overrides
     * @return array<string, mixed>
     */
    protected function attendeePayload(array $overrides = []): array
    {
        return array_merge([
            'id' => self::ATTENDEE_ID,
            'eventId' => self::EVENT_ID,
            'ticket' => self::EVENT_ID.'.'.self::ATTENDEE_ID,
            'firstName' => 'Ana',
            'lastName' => 'Pérez',
            'email' => 'ana@example.com',
            'scholarship' => false,
            'paymentStatus' => 'pending',
            'attendance' => 'registered',
            'registeredAt' => now()->toIso8601String(),
        ], $overrides);
    }

    /**
     * @return array<string, mixed>
     */
    protected function apiError(string $code, ?string $field = null): array
    {
        return array_filter(['error' => $code, 'field' => $field, 'message' => $code]);
    }
}
