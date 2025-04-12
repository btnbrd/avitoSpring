CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email VARCHAR(255) NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       role VARCHAR(50) NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pvz (
                                   id SERIAL PRIMARY KEY,
                                   city VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS goods_receipt (
                                             id SERIAL PRIMARY KEY,
                                             pvz_id INTEGER REFERENCES pvz(id),
    receipt_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) CHECK (status IN ('in_progress', 'close')) NOT NULL
    );

CREATE TABLE IF NOT EXISTS goods (
                                     id SERIAL PRIMARY KEY,
                                     receipt_id INTEGER REFERENCES goods_receipt(id),
    type VARCHAR(50) CHECK (type IN ('electronics', 'clothes', 'shoes')) NOT NULL,
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
