CREATE TABLE IF NOT EXISTS dates
(
    id SERIAL PRIMARY KEY,
    event_date  DATE NOT NULL UNIQUE
);