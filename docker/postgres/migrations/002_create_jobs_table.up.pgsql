CREATE TYPE job_status AS ENUM(
    'started',
    'running',
    'completed',
    'failed',
    'cancelled'
);

CREATE EXTENSION pgcrypto;

CREATE TABLE IF NOT EXISTS
    wikipediascraper.jobs (
        id serial primary key,
        job_id UUID NOT NULL UNIQUE,
        status job_status,
        created_at timestamp default current_timestamp
    );