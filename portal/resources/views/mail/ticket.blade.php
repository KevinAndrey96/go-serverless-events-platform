<!DOCTYPE html>
<html lang="es">
<body style="font-family: Arial, sans-serif; color: #1e293b; line-height: 1.5;">
    <p>Hola {{ $attendee->firstName }},</p>
    <p>Tu inscripción a <strong>{{ $event->name }}</strong> quedó registrada. El evento empieza el
        {{ $event->startsAt->translatedFormat('l j \d\e F \d\e Y \a \l\a\s g:i a') }}.</p>
    <p>Adjuntamos tu entrada con código QR. Preséntala en la puerta; cada código sirve para un solo ingreso.
        Si no puedes abrir la imagen, muestra este código: <code>{{ $attendee->ticket }}</code></p>
    @unless ($event->isFree())
        <p>El equipo revisará tu comprobante de pago antes del evento. Sin pago confirmado no podrás ingresar.</p>
    @endunless
    <p>Nos vemos pronto.</p>
</body>
</html>
