CREATE TABLE services (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL UNIQUE,
  url VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE health_checks (
  id SERIAL PRIMARY KEY,
  service_id INTEGER REFERENCES services(id),
  status VARCHAR(20) NOT NULL,
  response_time INTEGER,
  status_code INTEGER,
  error_message TEXT,
  checked_at TIMESTAMP DEFAULT NOW()
);

