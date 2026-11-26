# ADR-0003: Escrituras condicionales para cupos e ingreso

## Estado

Aceptada.

## Contexto

Dos carreras críticas aparecen en cualquier evento con cupos:

1. Varias personas toman el último cupo al mismo tiempo y el evento queda sobrevendido.
2. Dos personas en la puerta escanean la misma entrada (o un escáner reintenta) y entran dos personas con un solo código.

La versión anterior consultaba todos los asistentes y comparaba en memoria, sin ninguna garantía.

## Decisión

**Inscripción.** Una transacción `TransactWriteItems` con tres operaciones:

- `Update` del evento: `registered = registered + 1` con la condición `status = enabled AND registered < capacity AND startsAtUnix > :now`.
- `Put` del asistente con `attribute_not_exists(PK)`.
- `Put` de la guarda de correo con `attribute_not_exists(PK)`.

Si alguna condición falla, nada se escribe. El motivo (cupo agotado, evento cerrado o correo repetido) se lee en `CancellationReasons`; los conflictos entre transacciones se reintentan con espera exponencial con jitter.

**Ingreso.** Un `UpdateItem` con la condición `paymentStatus = paid AND attendance = registered`. La entrada es `<evento>.<asistente>`, con 128 bits aleatorios en el identificador del asistente, así que no se puede adivinar.

**Pagos.** Los eventos gratuitos inscriben con el pago confirmado; los pagos se revisan desde el portal.

## Consecuencias

- Las pruebas de integración contra DynamoDB Local lanzan 30 inscripciones simultáneas para 10 cupos (entran exactamente 10) y 5 validaciones simultáneas de la misma entrada (se admite exactamente 1).
- No hace falta un candado distribuido ni otra infraestructura.
- El cliente distingue cada rechazo por un código estable (`event_full`, `already_checked_in`, etc.).
