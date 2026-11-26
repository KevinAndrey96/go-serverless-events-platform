# ADR-0001: Registrar las decisiones de arquitectura

## Estado

Aceptada.

## Contexto

La plataforma combina funciones Lambda en Go, una tabla DynamoDB, una API HTTP con autorización IAM y un portal Laravel. Varias decisiones (modelo de datos, control de concurrencia, autorización) no se deducen leyendo el código y conviene dejar escrito por qué se tomaron.

## Decisión

Cada decisión con impacto en la arquitectura se documenta en `docs/adr/` con el formato de Michael Nygard: contexto, decisión y consecuencias. Una decisión reemplazada no se borra; se marca como sustituida y se enlaza la nueva.

## Consecuencias

- Quien llegue al proyecto entiende los compromisos sin depender de la memoria del equipo.
- Cambiar una decisión exige escribir el ADR correspondiente.
