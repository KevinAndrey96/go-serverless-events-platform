<?php

namespace App\Mail;

use App\EventsApi\Attendee;
use App\EventsApi\Event;
use App\Support\TicketQr;
use Illuminate\Mail\Attachment;
use Illuminate\Mail\Mailable;
use Illuminate\Mail\Mailables\Content;
use Illuminate\Mail\Mailables\Envelope;

final class TicketMail extends Mailable
{
    public function __construct(
        public readonly Event $event,
        public readonly Attendee $attendee,
    ) {}

    public function envelope(): Envelope
    {
        return new Envelope(subject: 'Tu entrada para '.$this->event->name);
    }

    public function content(): Content
    {
        return new Content(view: 'mail.ticket');
    }

    /**
     * @return array<int, Attachment>
     */
    public function attachments(): array
    {
        $png = app(TicketQr::class)->png($this->attendee->ticket);

        return [
            Attachment::fromData(fn () => $png, 'entrada-qr.png')->withMime('image/png'),
        ];
    }
}
