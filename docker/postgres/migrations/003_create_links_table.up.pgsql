CREATE TABLE IF NOT EXISTS
    wikipediascraper.links (
        id UUID PRIMARY KEY NOT NULL,
        job_id UUID NOT NULL REFERENCES wikipediascraper.jobs (job_id) ON DELETE CASCADE,
        link VARCHAR(255) NOT NULL UNIQUE,
        title VARCHAR(255) NOT NULL
    );