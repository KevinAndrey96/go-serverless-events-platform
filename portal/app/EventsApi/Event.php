<?php

namespace App\EventsApi;

use Carbon\CarbonImmutable;

final readonly class Event
{
    public function __construct(
        public string $id,
        public string $name,
        public CarbonImmutable $startsAt,
        public int $seatsLeft,
        public int $priceCop,
        public ?string $bank,
        public ?string $account,
        public ?int $capacity = null,
        public ?int $registered = null,
        public ?string $status = null,
    ) {}

    /**
     * @param  array<string, mixed>  $data
     */
    public static function fromArray(array $data): self
    {
        return new self(
            id: (string) $data['id'],
            name: (string) $data['name'],
            startsAt: CarbonImmutable::parse((string) $data['startsAt'])->setTimezone(config('events.timezone')),
            seatsLeft: (int) $data['seatsLeft'],
            priceCop: (int) ($data['priceCop'] ?? 0),
            bank: isset($data['bank']) ? (string) $data['bank'] : null,
            account: isset($data['account']) ? (string) $data['account'] : null,
            capacity: isset($data['capacity']) ? (int) $data['capacity'] : null,
            registered: isset($data['registered']) ? (int) $data['registered'] : null,
            status: isset($data['status']) ? (string) $data['status'] : null,
        );
    }

    public function isFree(): bool
    {
        return $this->priceCop === 0;
    }

    public function isCancelled(): bool
    {
        return $this->status === 'cancelled';
    }

    public function formattedPrice(): string
    {
        return $this->isFree() ? 'Gratis' : '$ '.number_format($this->priceCop, 0, ',', '.').' COP';
    }
}
