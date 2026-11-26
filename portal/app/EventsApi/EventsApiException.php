<?php

namespace App\EventsApi;

use Illuminate\Http\Client\Response;
use RuntimeException;
use Throwable;

final class EventsApiException extends RuntimeException
{
    private const MESSAGES = [
        'event_not_found' => 'El evento no existe.',
        'attendee_not_found' => 'El asistente no existe.',
        'invalid_ticket' => 'El código QR no corresponde a ninguna inscripción.',
        'event_full' => 'El evento ya no tiene cupos disponibles.',
        'event_closed' => 'El evento no está recibiendo inscripciones.',
        'already_registered' => 'Este correo ya está inscrito en el evento.',
        'already_checked_in' => 'Este código ya fue usado para ingresar.',
        'payment_pending' => 'El pago de este asistente aún no está confirmado.',
        'validation' => 'Revisa los datos enviados.',
        'unavailable' => 'El servicio de eventos no está disponible. Intenta de nuevo en unos minutos.',
    ];

    public function __construct(
        public readonly string $errorCode,
        public readonly int $status,
        public readonly ?string $field = null,
        string $message = '',
        ?Throwable $previous = null,
    ) {
        parent::__construct($message !== '' ? $message : $errorCode, 0, $previous);
    }

    public static function fromResponse(Response $response): self
    {
        $body = $response->json();
        $body = is_array($body) ? $body : [];

        return new self(
            errorCode: is_string($body['error'] ?? null) ? $body['error'] : 'unavailable',
            status: $response->status(),
            field: is_string($body['field'] ?? null) ? $body['field'] : null,
            message: is_string($body['message'] ?? null) ? $body['message'] : '',
        );
    }

    public static function unavailable(Throwable $previous): self
    {
        return new self('unavailable', 503, previous: $previous);
    }

    /**
     * Spanish text that can be shown to the person using the portal.
     */
    public function userMessage(): string
    {
        return self::MESSAGES[$this->errorCode] ?? self::MESSAGES['unavailable'];
    }

    public function isValidation(): bool
    {
        return $this->errorCode === 'validation';
    }
}
