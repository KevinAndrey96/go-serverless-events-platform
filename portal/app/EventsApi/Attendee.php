<?php

namespace App\EventsApi;

use Carbon\CarbonImmutable;

final readonly class Attendee
{
    public function __construct(
        public string $id,
        public string $eventId,
        public string $ticket,
        public string $firstName,
        public string $lastName,
        public string $email,
        public ?string $phone,
        public bool $scholarship,
        public string $paymentStatus,
        public string $attendance,
        public CarbonImmutable $registeredAt,
        public ?CarbonImmutable $checkedInAt,
    ) {}

    /**
     * @param  array<string, mixed>  $data
     */
    public static function fromArray(array $data): self
    {
        $zone = config('events.timezone');

        return new self(
            id: (string) $data['id'],
            eventId: (string) $data['eventId'],
            ticket: (string) $data['ticket'],
            firstName: (string) $data['firstName'],
            lastName: (string) $data['lastName'],
            email: (string) $data['email'],
            phone: isset($data['phone']) ? (string) $data['phone'] : null,
            scholarship: (bool) ($data['scholarship'] ?? false),
            paymentStatus: (string) $data['paymentStatus'],
            attendance: (string) $data['attendance'],
            registeredAt: CarbonImmutable::parse((string) $data['registeredAt'])->setTimezone($zone),
            checkedInAt: isset($data['checkedInAt'])
                ? CarbonImmutable::parse((string) $data['checkedInAt'])->setTimezone($zone)
                : null,
        );
    }

    public function fullName(): string
    {
        return trim($this->firstName.' '.$this->lastName);
    }

    public function isPaid(): bool
    {
        return $this->paymentStatus === 'paid';
    }

    public function hasAttended(): bool
    {
        return $this->attendance === 'attended';
    }
}
