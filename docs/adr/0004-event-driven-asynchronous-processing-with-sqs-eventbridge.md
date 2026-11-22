# ADR-0004: Asynchronous Reminder & Notification Processing via SQS/EventBridge

## Status
`Accepted`

## Context
When an event status changes or bulk reminder emails must be dispatched to thousands of attendees (e.g. 24 hours prior to event start), executing synchronous loops in the API handler causes HTTP request timeouts in AWS API Gateway (maximum 29-second ceiling) and creates cascaded failures if downstream email providers experience transient throttling.

## Decision
We decouple notification triggers from execution using an **Asynchronous Event-Driven Architecture**.

1. **Trigger Phase:** The primary event service publishes state change events to Amazon EventBridge or pushes work batches to an Amazon SQS FIFO queue.
2. **Worker Processing:** A dedicated background worker (`backend/reminder/v1`) consumes batches with controlled concurrency, rate-limits outbound SMTP/SES transactions, and leverages Dead Letter Queues (DLQ) for failed message isolation.
3. **Resilience & Exponential Backoff:** Transient network failures trigger automatic retries with exponential jitter without impacting client-facing API response latency.

## Consequences
- **Positive:**
  - **Instantaneous Client Response:** User requests return immediately with `202 Accepted` or `200 OK` regardless of attendee list size.
  - **Fault Isolation:** Email provider degradation or SMTP rate limits do not block core ticket booking or QR check-in flows.
  - **Automatic Dead Letter Queueing (DLQ):** Poison pills or malformed email payloads are quarantined for inspection without stalling the processing pipeline.
- **Negative:**
  - Requires eventual consistency management on the frontend status display.
