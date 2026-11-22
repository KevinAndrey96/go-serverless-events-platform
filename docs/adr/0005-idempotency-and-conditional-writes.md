# ADR-0005: Idempotency & Concurrency Control with DynamoDB Conditional Writes

## Status
`Accepted`

## Context
In live event ticketing, two critical concurrency failure modes frequently arise:
1. **Double-Booking / Capacity Overselling:** Multiple simultaneous purchasers attempt to claim the final ticket for a limited-capacity event.
2. **Double Check-In:** Network retries from mobile ticket scanners at physical venue gates, or duplicated QR scans, could mark a ticket as entered multiple times, leading to security breaches or fraudulent pass sharing.

## Decision
We enforce **Optimistic Concurrency Control and Idempotency** via **Amazon DynamoDB Conditional Writes (`attribute_not_exists` and `attribute_exists`)**.

### Check-in Idempotency Pattern:
When scanning an attendee ticket at the venue gate, the `checkin` handler issues an atomic `UpdateItem` with a condition expression:
```
ConditionExpression: "attribute_not_exists(checkin_timestamp) AND payment_status = :paid"
```
If the ticket was already checked in (or has unpaid status), DynamoDB rejects the write atomically with a `ConditionalCheckFailedException`, returning HTTP `409 Conflict` to the scanner application without requiring distributed distributed locks (e.g. Redis Redlock).

### State Transition Idempotency:
Transitions from `PENDING` to `PAID` or `CANCELLED` utilize conditional evaluations guaranteeing that state mutations only succeed if the entity is currently in the expected preceding lifecycle state.

## Consequences
- **Positive:**
  - **Zero Over-Selling & Fraud Elimination:** Atomic hardware-level isolation at the DynamoDB partition level prevents race conditions.
  - **No Distributed Lock Overhead:** Avoids maintaining external lock state in Redis or Memcached, reducing architectural complexity and infrastructure costs.
- **Negative:**
  - Client applications must gracefully handle and classify `ConditionalCheckFailedException` responses.
