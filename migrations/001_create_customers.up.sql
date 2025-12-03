
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    location VARCHAR(100),
    preferred_product VARCHAR(100)
);
