CREATE SCHEMA IF NOT EXISTS virgin_media_outage_reporter;

SET search_path TO virgin_media_outage_reporter, public;

CREATE TABLE IF NOT EXISTS outages
(
    id           SERIAL       NOT NULL PRIMARY KEY,
    vm_outage_id VARCHAR(255) NOT NULL,
    data         TEXT         NOT NULL,

    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);