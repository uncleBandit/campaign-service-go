-- Insert 3 sample campaigns
INSERT INTO campaigns (name, channel, status, base_template, scheduled_at, created_at)
VALUES
('Summer Sale 2025', 'sms', 'draft', 'Hi {first_name}, check out {preferred_product} in {location}!', NOW() + INTERVAL '1 day', NOW()),
('Winter Promotion', 'whatsapp', 'draft', 'Hello {first_name}, enjoy our {preferred_product} this winter!', NOW() + INTERVAL '2 days', NOW()),
('Flash Deals', 'sms', 'draft', 'Hi {first_name}, don''t miss {preferred_product} in {location}!', NOW(), NOW())
