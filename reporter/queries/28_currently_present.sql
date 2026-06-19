-- Devices observed in the last 10 minutes of the capture: "who is here
-- right now". Run this against a live capture to get a snapshot of the
-- current BLE environment.

WITH cutoff AS (
    SELECT max(last_seen) - INTERVAL 10 MINUTE AS since FROM obs
)
SELECT
    address,
    coalesce(mfr_name, '(none)')             AS manufacturer,
    address_type,
    coalesce(local_name, '')                 AS local_name,
    max(last_seen)                           AS last_seen,
    count(*)                                 AS windows,
    sum(sample_count)                        AS samples,
    round(avg(rssi_mean), 1)                 AS avg_rssi,
    -- Bands tuned for this UK dormer bungalow with scanner in upstairs office:
    -- lounge = -76..-80, downstairs office (far end) / dining = -86..-92.
    CASE
        WHEN avg(rssi_mean) > -65 THEN 'office (scanner room)'
        WHEN avg(rssi_mean) > -80 THEN 'directly downstairs (lounge / near)'
        WHEN avg(rssi_mean) > -94 THEN 'downstairs far (dining / office)'
        ELSE                            'outside / next door'
    END                                      AS location
FROM obs, cutoff
WHERE last_seen >= cutoff.since
GROUP BY address, mfr_name, address_type, local_name
ORDER BY last_seen DESC;
