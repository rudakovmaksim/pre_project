CREATE TABLE IF NOT EXISTS rates
(
    id SERIAL PRIMARY KEY,
    title_id INT NOT NULL,
    day_id INT NOT NULL,
    cost DOUBLE PRECISION NOT NULL,
    event_time TIME NOT NULL,
    FOREIGN KEY (title_id) REFERENCES titles(id),
    FOREIGN KEY (day_id) REFERENCES dates(id)
);