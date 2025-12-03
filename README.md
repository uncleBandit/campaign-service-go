# SMSLeopard Backend

This repository contains a Go backend service for managing and sending SMS/WhatsApp campaigns as part of the SMSLeopard Engineering Challenge.

---

## How to Run the Service

### Prerequisites
- Go 1.21+
- PostgreSQL (or SQLite for testing)
- RabbitMQ (optional; in-memory queue is supported for testing)

### Setup

1. Clone the repository:

```bash
git clone <your-repo-url>
cd smsleopard-backend
Configure environment variables in a .env file:

.env
Copy code
DB_HOST=localhost
DB_PORT=5433
DB_NAME=smsleopard
DB_USER=smsleopard
DB_PASSWORD=secret
QUEUE_URL=amqp://guest:guest@localhost:5672/
Run database migrations / seed sample data (5 customers, 2-3 campaigns).

Start the service:

bash
Copy code
go run ./cmd/server
The API will be available at http://localhost:8080.

Running Tests
bash
Copy code
go test ./...
This runs all unit and integration tests, including template rendering, campaign sending, and in-memory queue processing.

Assumptions Made
POST /campaigns/{id}/send allows manually specifying customer IDs for sending.

Campaigns with a future scheduled_at are not automatically dispatched (automatic scheduling is optional/extras).

For simplicity, messages are always sent successfully in the mock sender unless testing retry logic.

Only basic template placeholders are supported ({first_name}, {last_name}, {preferred_product}, {location}).

Template Handling
Placeholders in templates are replaced with the corresponding customer fields.

If a customer field is missing or null, the placeholder is replaced with [unknown].

Example:

Template: "Hi {first_name}, check out {preferred_product}!"
Customer: {FirstName: "", PreferredProduct: "Shoes"}
Rendered: "Hi [unknown], check out Shoes!"

Mock Sender Behavior
Implemented a mock sender in the queue worker.

All messages succeed by default.

In tests, failed messages can be simulated by manually setting a failure flag.

Upon sending, the message status in the database is updated to:

sent on success

failed on failure (with retry_count incremented if applicable)

Queue Choice
For production-like setup, RabbitMQ is used as the message queue.

For tests and demonstration, an in-memory channel queue is used.

Reasoning:

RabbitMQ allows realistic async processing.

In-memory queue simplifies testing without external dependencies.

Can switch between the two easily by toggling the controller method.

EXTRA_FEATURE
Feature: Scheduled dispatch worker

Reasoning: Campaigns with scheduled_at can be dispatched automatically without manual intervention, which mimics production behavior.

Implementation Details:

A background worker checks campaigns with scheduled_at <= now() every minute.

Adds outbound messages to the queue for processing.

Updates campaign status to sending once queued.

Limitations:

Scheduling interval is fixed (1 minute) and not highly precise.

No rate-limiting or distributed scheduler implemented.

Only works while the service is running (no persistent job scheduler).

Summary
This backend implements the following:

CRUD for campaigns

Personalized message templates

Queue-based asynchronous sending

In-memory and RabbitMQ queue support

Retry logic for failed messages

Tests covering template rendering, campaign sending, and worker processing

Contact
For any questions or feedback, please contact pitmwania@gmail.com.



### Idempotency

Idempotency

Our system guarantees that each customer receives at most one message per campaign, ensuring accurate reporting, efficient resource usage, and system stability.

Key Features:

Database-level uniqueness:
The (campaign_id, customer_id) unique constraint prevents duplicate outbound messages at the database level.

Graceful repository handling:
Attempts to create duplicates are safely ignored using ON CONFLICT DO NOTHING, ensuring that repeated send requests do not cause errors.

Safe queue processing:
The queue worker processes each outbound message exactly once, so duplicates are never created during asynchronous processing.

Idempotent API:
Calling POST /campaigns/{id}/send multiple times with the same customer IDs is safe; no duplicate messages will be generated.

Reasoning:

Accurate analytics:
Prevents skewed reporting by ensuring each customer is counted only once per campaign.

Resource efficiency:
Avoids unnecessary processing, network usage, and database writes caused by duplicate messages.

System stability:
Protects against accidental or repeated requests that could overload the worker or database.

Implementation Summary:

Unique constraint in the database enforces one message per customer per campaign.

Repository uses idempotent inserts to handle duplicates gracefully.

Queue worker and API logic ensure messages are never duplicated in practice.

Outcome:
The system is reliable, predictable, and safe to use even under retries or repeated requests, reflecting production-grade behavior.


##Time & Tools Note

Approximate Time Spent: ~9 hours

Tools & Environment: Developed on Ubuntu using VS Code. PostgreSQL was used as the database, and Go 1.21+ as the programming language.

Use of AI Tools: ChatGPT was leveraged for guidance on code structure, syntax clarification, and debugging assistance. All core logic and implementation decisions were authored manually.

Summary: The majority of time was spent designing the data model, implementing the queue worker, and ensuring idempotency, while AI helped accelerate development and reduce minor errors.
