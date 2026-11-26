# ADR-0005: Autorización IAM para las rutas de organizadores

## Estado

Aceptada.

## Contexto

La API original no tenía autenticación y respondía con CORS `*`: cualquiera podía crear eventos, cambiar pagos o validar entradas. Los organizadores usan el portal Laravel, que ya autentica a su equipo; el navegador nunca llama la API directamente.

## Decisión

- Las rutas públicas (`GET /events` y `POST /events/{eventId}/attendees`) no requieren autorización y tienen límites de tasa propios en API Gateway.
- Las rutas `/admin/*` usan el autorizador `AWS_IAM` del HTTP API. El portal firma esas solicitudes con SigV4 usando la cadena de credenciales por defecto (rol de la tarea, perfil o variables de entorno).
- La plantilla publica la política administrada `PortalInvokePolicy`, que solo permite `execute-api:Invoke` sobre `/admin/*`, para asignarla al rol del portal.
- Cada función tiene su propia política con las acciones mínimas sobre la tabla (por ejemplo, la validación de entradas solo puede `UpdateItem` y `GetItem`).

## Consecuencias

- No hay secretos compartidos que rotar entre el portal y la API; las credenciales son temporales.
- Otro cliente de organizadores (por ejemplo, una app de escáner) necesita credenciales IAM o un autorizador adicional.
- Localmente el servidor de desarrollo no firma ni valida; solo escucha en la red de Docker Compose.
