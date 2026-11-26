# ADR-0004: Recordatorios programados con reclamo condicional

## Estado

Aceptada. Sustituye la propuesta de usar colas SQS para los recordatorios, que nunca se implementó.

## Contexto

Cada asistente debe recibir un recordatorio en las 24 horas previas al evento, una sola vez. La versión anterior enviaba correos a todos los inscritos cada seis horas.

## Decisión

Una función programada con EventBridge Scheduler corre cada hora:

1. Consulta en `GSI1` los eventos habilitados que empiezan en las próximas 24 horas.
2. Por cada asistente, reclama el recordatorio con `SET reminderSentAt` condicionado a `attribute_not_exists(reminderSentAt)`.
3. Solo quien gana el reclamo envía el correo por SES v2. Si el envío falla, libera el reclamo y la siguiente ejecución lo reintenta.

La función tiene concurrencia reservada de 1 y el programador no reintenta, así que no hay ejecuciones superpuestas; aun si las hubiera, el reclamo condicional evita duplicados.

## Consecuencias

- Cada persona recibe un recordatorio, aunque la función se ejecute muchas veces.
- Sin colas ni colas de mensajes fallidos que operar. Para decenas de miles de asistentes por hora convendría repartir el trabajo con SQS.
- SES exige una identidad verificada; el permiso `ses:SendEmail` se limita a esa identidad.
