CREATE TABLE outbound_messages (
    id SERIAL PRIMARY KEY,
    campaign_id INT NOT NULL REFERENCES campaigns(id),
    customer_id INT NOT NULL REFERENCES customers(id),
    status TEXT NOT NULL DEFAULT 'pending',
    rendered_content TEXT,
    last_error TEXT,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (campaign_id, customer_id)  -- ensures idempotency
);

