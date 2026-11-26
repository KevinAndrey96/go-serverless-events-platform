<?php

namespace App\Support;

use chillerlan\QRCode\Output\QRGdImagePNG;
use chillerlan\QRCode\Output\QRMarkupSVG;
use chillerlan\QRCode\QRCode;
use chillerlan\QRCode\QROptions;

/**
 * Renders the ticket value as a QR code: SVG for pages, PNG for email attachments.
 */
final class TicketQr
{
    public function svg(string $ticket): string
    {
        return (new QRCode(new QROptions([
            'outputInterface' => QRMarkupSVG::class,
            'outputBase64' => false,
            'addQuietzone' => true,
            'svgAddXmlHeader' => false,
        ])))->render($ticket);
    }

    public function png(string $ticket): string
    {
        return (new QRCode(new QROptions([
            'outputInterface' => QRGdImagePNG::class,
            'outputBase64' => false,
            'scale' => 10,
            'addQuietzone' => true,
        ])))->render($ticket);
    }
}
