# ADR-0002: Diseño de tabla única en DynamoDB

## Estado

Aceptada.

## Contexto

Los patrones de acceso son pocos y conocidos: listar eventos por fecha, leer un evento con sus asistentes, inscribir sin sobrecupo, marcar pagos y validar entradas en la puerta. Las inscripciones llegan en ráfagas cuando se publica un evento, y el ingreso concentra lecturas y escrituras en pocos minutos. Una base relacional detrás de Lambda obliga a administrar conexiones; DynamoDB escala por solicitud y cobra por uso.

## Decisión

Una sola tabla con llave compuesta `PK`/`SK` y un índice global `GSI1`:

| Entidad | PK | SK | GSI1PK | GSI1SK |
| :--- | :--- | :--- | :--- | :--- |
| Evento | `EVENT#<id>` | `EVENT` | `EVENTS` | `<inicio UTC>#<id>` |
| Asistente | `EVENT#<id>` | `ATTENDEE#<asistente>` | | |
| Guarda de correo | `EVENT#<id>` | `EMAIL#<correo>` | | |

- El evento y sus asistentes comparten partición: la lista de asistentes es un solo `Query`.
- `GSI1` ordena los eventos por fecha de inicio; el recordatorio consulta un rango de fechas sin `Scan`.
- La guarda de correo convierte "una inscripción por correo y evento" en una escritura condicional.
- `startsAtUnix` guarda el inicio como número para compararlo en condiciones sin depender de zonas horarias.

## Consecuencias

- Todas las operaciones son lecturas por llave o consultas acotadas, con latencia estable.
- `GSI1` concentra todos los eventos en una partición (`EVENTS`). Es adecuado para cientos de eventos activos; con volúmenes mayores habría que fragmentar la llave por mes.
- Los reportes ad hoc requieren exportar los datos (por ejemplo, exportación a S3 y Athena).
