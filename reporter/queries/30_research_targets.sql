-- Research-prioritization view: every manufacturer ID and service UUID
-- that the scanner doesn't deep-parse, ranked by frequency, with sample
-- bytes for protocol research and RSSI-based proximity scoring so you
-- can tell "inside my house" devices from "passing on the street" ones.
--
-- Proximity guide (your thresholds will vary by scanner placement):
--   strong (>-65)  — same room or directly adjacent
--   indoor (>-80)  — likely somewhere in the house
--   nearby (>-90)  — could be inside, could be just outside walls
--   far    (else)  — almost certainly passing outside / next door

WITH mfr_targets AS (
    SELECT
        'manufacturer'                              AS kind,
        printf('0x%04x', oj.mfr_id)                 AS id,
        oj.mfr_name                                 AS vendor,
        o.rssi_mean,
        o.rssi_min,
        o.rssi_max,
        oj.address,
        json_extract_string(oj.mfr_details_json, '$.unparsed') AS sample_payload_b64
    FROM obs_json oj
    JOIN obs o ON o.address = oj.address AND o.first_seen = oj.first_seen
    WHERE json_extract_string(oj.mfr_details_json, '$.unparsed') IS NOT NULL
),
svc_unknown AS (
    SELECT
        oj.address,
        oj.first_seen,
        je.key                                                 AS svc_key,
        je.value                                               AS svc_value,
        coalesce(
            json_extract_string(je.value, '$.name'),
            n.vendor_name,
            '(unknown)'
        )                                                      AS vendor
    FROM obs_json oj, json_each(oj.svc_details_json) je
    LEFT JOIN service_uuid_names n ON n.uuid = substring(je.key, 9)
    WHERE je.key LIKE 'service_%'
),
svc_targets AS (
    SELECT
        'service'   AS kind,
        u.svc_key   AS id,
        u.vendor,
        o.rssi_mean,
        o.rssi_min,
        o.rssi_max,
        u.address,
        u.svc_value::VARCHAR AS sample_payload_b64
    FROM svc_unknown u
    JOIN obs o ON o.address = u.address AND o.first_seen = u.first_seen
),
all_targets AS (
    SELECT * FROM mfr_targets
    UNION ALL
    SELECT * FROM svc_targets
)
SELECT
    kind,
    id,
    any_value(vendor)               AS vendor,
    count(*)                        AS observations,
    count(DISTINCT address)         AS unique_addrs,
    round(avg(rssi_mean), 0)::INT   AS avg_rssi,
    min(rssi_min)                   AS rssi_min,
    max(rssi_max)                   AS rssi_max,
    -- Bands tuned for this UK dormer bungalow with scanner in upstairs
    -- office. Reference: far-downstairs office press = -86 to -92 dBm.
    CASE
        WHEN avg(rssi_mean) > -65 THEN 'office (scanner room)'
        WHEN avg(rssi_mean) > -80 THEN 'upstairs / directly downstairs'
        WHEN avg(rssi_mean) > -94 THEN 'downstairs (mid/far)'
        ELSE                            'outside / next door'
    END                             AS proximity,
    any_value(sample_payload_b64)   AS sample_payload_b64
FROM all_targets
GROUP BY kind, id
ORDER BY observations DESC;
