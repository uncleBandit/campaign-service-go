# SMSLeopard Backend – System Overview

This document provides an overview of the SMSLeopard backend design, focusing on the campaign dispatch system, queue worker, data model, and personalization logic.

---

## 1. Data Model

The system uses a relational database (PostgreSQL preferred) with three main entities:

### `customers`
| Column            | Type    | Notes       |
| ----------------- | ------- | ----------- |
| id                | integer | primary key |
| phone             | string  |             |
| first_name        | string  |             |
| last_name         | string  |             |
| location          | string  |             |
| preferred_product | string  |             |

Indexes:
- `phone` (unique)
- Optional: `location` for targeted campaigns

---

### `campaigns`
| Column        | Type      | Notes                                                  |
| ------------- | --------- | ------------------------------------------------------ |
| id            | integer   | primary key                                            |
| name          | string    |                                                        |
| channel       | string    | `sms` or `whatsapp`                                   |
| status        | string    | `draft`, `scheduled`, `sending`, `sent`, `failed`    |
| base_template | text      | e.g., `"Hi {first_name}, check out {preferred_product}"` |
| scheduled_at  | timestamp | nullable                                               |
| created_at    | timestamp |                                                        |

Indexes:
- `status` and `created_at` for efficient filtering and ordering

---

### `outbound_messages`
| Column           | Type      | Notes                           |
| ---------------- | --------- | ------------------------------- |
| id               | integer   | primary key                     |
| campaign_id      | integer   | foreign key → campaigns         |
| customer_id      | integer   | foreign key → customers         |
| status           | string    | `pending`, `sent`, `failed`    |
| rendered_content | text      | final personalized message      |
| last_error       | text      | nullable                        |
| retry_count      | integer   | defaults to 0                   |
| created_at       | timestamp |                                 |
| updated_at       | timestamp |                                 |

Indexes:
- `campaign_id` + `status` for querying messages by campaign and status

---

## 2. Request Flow: `POST /campaigns/{id}/send`

1. **Input:** `customer_ids` array
2. **Validation:** Confirm campaign exists and status is `draft` or `scheduled`
3. **Outbound Messages:** Create `outbound_messages` rows in the database with `status = pending`
4. **Queue Publish:** Push each `outbound_message_id` to the queue (`campaign_sends`)
5. **Update Campaign Status:** Set to `sending`
6. **Response:** Return `campaign_id`, number of messages queued, and status (`sending`)

---

## 3. Queue Worker & Retry Logic

- **Listening:** Worker subscribes to `campaign_sends` queue
- **Processing:**
  1. Fetch `outbound_message` with related `campaign` and `customer`
  2. Render message using `base_template` + customer data
  3. Call **mock sender**
     - Succeeds by default (failure can be simulated in tests)
  4. Update `outbound_messages.status`:
     - `sent` on success
     - `failed` on failure; increment `retry_count`
  5. Retry Logic:
     - Max retries = 3
     - Failed messages re-queued for retry
     - After max retries, message marked as `failed`
- **Acknowledgements:** Only ack messages after successful DB update

---

## 4. Pagination Strategy

- `GET /campaigns` supports `page` and `page_size` parameters
- **Ordering:** `id DESC` ensures newest campaigns first
- **Avoiding duplicates/missing records:**
  - Stable sorting by `id` prevents duplicate/missing records when new campaigns are added during pagination
- **Filtering:** Optional by `status` or `channel`
- **Default:** `page = 1`, `page_size = 20`, max `page_size = 100`

---

## 5. Personalization Approach

- **Template System:**
  - Supports placeholders: `{first_name}`, `{last_name}`, `{preferred_product}`, `{location}`
  - Replaces missing/null customer fields with `[unknown]`
- **Rendering:** Done in the worker before sending messages
- **Preview Endpoint:** `POST /campaigns/{id}/personalized-preview` allows rendering a message for a single customer without queueing
- **Extension Points:**
  - AI-driven personalization can replace template substitution in future
  - Custom dynamic variables could be added per campaign
  - Integration with external CRM or analytics systems

---

## Notes

- Database relationships: `campaigns` → `outbound_messages` → `customers`
- Queue system: RabbitMQ for production; in-memory queue for testing
- Scheduled dispatch worker implemented as an extra feature (polling for campaigns with `scheduled_at <= now()`)
- The system prioritizes reliability, simplicity, and testability

