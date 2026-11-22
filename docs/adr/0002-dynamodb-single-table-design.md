# ADR-0002: Amazon DynamoDB Single-Table Design Architecture

## Status
`Accepted`

## Context
Event management systems experience bursty, high-concurrency traffic during ticket sales and live venue check-ins (1,000+ requests per second per venue). Traditional relational databases face connection pool exhaustion (RDBMS connection starvation in serverless environments) and degraded query performance due to multi-table `JOIN` operations across `Events`, `Users`, `Tickets`, and `Attendees`.

Additionally, hosting individual DynamoDB tables per entity leads to fragmented read/write capacity allocation, higher costs, and requires multiple round-trip HTTP requests to assemble aggregate domain models.

## Decision
We adopt **Amazon DynamoDB Single-Table Design** across the platform (`EventsPlatformTable`). All core entities (`Event`, `User`, `EventUser`, `CheckIn`) are co-located within a single unified table schema utilizing composite Partition Keys (`pk`) and Sort Keys (`sk`), supplemented by two Global Secondary Indexes (`sk-index` and `eventid-index`).

### Schema Design & Partition Key Patterns:
| Entity | Partition Key (`pk`) | Sort Key (`sk`) | Description |
| :--- | :--- | :--- | :--- |
| **Event Metadata** | `EVENT#<event_id>` | `METADATA#<event_id>` | Core event details (name, capacity, date, status). |
| **Event Registration** | `EVENT#<event_id>` | `USER#<user_id>` | Attendee registration & payment state. |
| **User Profile** | `USER#<user_id>` | `PROFILE#<user_id>` | User contact & authentication info. |
| **Live QR Check-in** | `EVENT#<event_id>` | `CHECKIN#<ticket_id>` | Real-time entry validation and scanner timestamp. |

### Global Secondary Indexes (GSIs):
1. **`sk-index` (GSI 1):** Inverted partition index querying all events associated with a specific `USER#<user_id>` without triggering a full table scan.
2. **`eventid-index` (GSI 2):** Sparse projection index optimized for batch attendee retrieval and scheduled reminder notifications.

## Consequences
- **Positive:**
  - **Predictable Single-Digit Latency:** Operational queries return in <10ms regardless of database volume by pre-computing access patterns.
  - **Single Network Round-Trip:** An event and all its attendees can be retrieved in a single `Query` call (`pk = EVENT#123 AND begins_with(sk, "USER#")`).
  - **Serverless Connection Resilience:** DynamoDB communicates over stateless HTTP/2 (HTTPS), eliminating connection pooling exhaustion common in relational databases under Lambda scaling.
- **Negative:**
  - **Rigid Query Patterns:** Ad-hoc analytics or multi-dimensional filters require ETL streaming (DynamoDB Streams to Athena / OpenSearch).
  - **Learning Curve:** Developers must model around access patterns rather than normalized relational schemas.
