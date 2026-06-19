-- Base views over the scanner's JSONL captures. Sourced by every query in
-- queries/ via run.sh, so changes here propagate everywhere.
--
-- DuckDB infers the schema by sampling rows, so the manufacturerdata.details
-- struct only contains keys actually seen. That's fine for the typed
-- per-vendor queries; for "what subtype keys exist" exploration we re-read
-- the same files as raw JSON and inspect dynamically.

INSTALL json;
LOAD json;

CREATE OR REPLACE VIEW obs AS
SELECT
    "@timestamp"::TIMESTAMP                            AS ts,
    Common.address                                     AS address,
    Common.address_type                                AS address_type,
    Common.first_seen::TIMESTAMP                       AS first_seen,
    Common.last_seen::TIMESTAMP                        AS last_seen,
    Common.name                                        AS local_name,
    observation.count                                  AS sample_count,
    observation.rssi_min                               AS rssi_min,
    observation.rssi_max                               AS rssi_max,
    observation.rssi_mean                              AS rssi_mean,
    observation.rssi_max - observation.rssi_min        AS rssi_range,
    observation.rssi_samples                           AS rssi_samples,
    epoch_ms(Common.last_seen::TIMESTAMP) - epoch_ms(Common.first_seen::TIMESTAMP) AS window_ms,
    manufacturerdata.id                                AS mfr_id,
    manufacturerdata.name                              AS mfr_name,
    manufacturerdata.details                           AS mfr_details,
    servicedata.name                                   AS svc_name,
    servicedata.id                                     AS svc_id
FROM read_json_auto('captures/*.jsonl', union_by_name=true);

-- Raw JSON view for dynamic key inspection (e.g. discovering new Apple
-- subtypes appearing as unknown_0xNN). One column per top-level group, kept
-- as JSON so we can json_keys() against it without schema lock-in.
CREATE OR REPLACE VIEW obs_json AS
SELECT
    Common.address                AS address,
    Common.address_type           AS address_type,
    Common.first_seen::TIMESTAMP  AS first_seen,
    Common.last_seen::TIMESTAMP   AS last_seen,
    observation.count             AS sample_count,
    manufacturerdata.id           AS mfr_id,
    manufacturerdata.name         AS mfr_name,
    manufacturerdata.details::JSON AS mfr_details_json,
    servicedata.uuids             AS svc_uuids,
    servicedata.details::JSON     AS svc_details_json
FROM read_json_auto('captures/*.jsonl', union_by_name=true);
