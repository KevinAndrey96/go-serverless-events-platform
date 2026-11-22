# 🏛️ Architecture Decision Records (ADRs)

This directory documents the core architectural and design decisions for the **Serverless Events Platform** monorepo using Michael Nygard's Architecture Decision Record (ADR) format.

## Architectural Governance & Principles

1. **Decoupled Asynchrony:** Non-blocking operations (notifications, reminder dispatches, status fanouts) must execute asynchronously via queues and events rather than blocking synchronous HTTP request/response loops.
2. **Deterministic Cold Starts:** Dependencies in AWS Lambda execution environments must be resolved at compile time to eliminate runtime reflection overhead and minimize cold start latency (<50ms).
3. **Optimistic Concurrency & Idempotency:** State transitions (ticket check-ins, payment confirmations) must guarantee atomic execution via conditional expressions to eliminate double-spending and race conditions.
4. **Predictable Query Latency:** High-volume operational reads must achieve sub-10ms latency regardless of database scale using pre-computed DynamoDB Single-Table access patterns and sparse indexes.

---

## 📑 Index of Architectural Decisions

| ADR | Title | Status | Date | Decision Area |
| :---: | :--- | :---: | :---: | :--- |
| **[ADR-0001](0001-record-architecture-decisions.md)** | Record Architecture Decisions using ADR Format | `Accepted` | 2026-11-20 | Architecture Governance |
| **[ADR-0002](0002-dynamodb-single-table-design.md)** | Amazon DynamoDB Single-Table Design Architecture | `Accepted` | 2026-11-20 | Data Modeling & Storage |
| **[ADR-0003](0003-compile-time-dependency-injection-with-wire.md)** | Compile-Time Dependency Injection using Google Wire | `Accepted` | 2026-11-21 | Performance & Cold Starts |
| **[ADR-0004](0004-event-driven-asynchronous-processing-with-sqs-eventbridge.md)** | Asynchronous Reminder & Notification Processing via SQS/EventBridge | `Accepted` | 2026-11-21 | Messaging & Resilience |
| **[ADR-0005](0005-idempotency-and-conditional-writes.md)** | Idempotency & Concurrency Control with DynamoDB Conditional Writes | `Accepted` | 2026-11-21 | Distributed Consistency |
