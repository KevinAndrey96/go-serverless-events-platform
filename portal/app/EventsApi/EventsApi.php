<?php

namespace App\EventsApi;

use Carbon\CarbonInterface;

/**
 * Client of the Go events API. Every method throws EventsApiException on failure.
 */
interface EventsApi
{
    /** @return list<Event> */
    public function upcomingEvents(): array;

    /**
     * @param  array{firstName: string, lastName: string, email: string, phone: ?string, gender: ?string, scholarship: bool}  $registration
     */
    public function register(string $eventId, array $registration): Attendee;

    /** @return list<Event> */
    public function allEvents(): array;

    public function createEvent(string $name, CarbonInterface $startsAt, int $capacity, int $priceCop, ?string $bank, ?string $account): Event;

    public function changeEventStatus(string $eventId, string $status): Event;

    /** @return array{event: Event, attendees: list<Attendee>} */
    public function attendees(string $eventId): array;

    public function changePaymentStatus(string $eventId, string $attendeeId, string $paymentStatus): Attendee;

    public function checkIn(string $ticket): Attendee;
}
