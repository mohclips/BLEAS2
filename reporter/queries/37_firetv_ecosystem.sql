-- Map the Fire TV ecosystem in the house. The same service UUID
-- (0x3e1d50cd...) is broadcast by BOTH the TVs (idle while powered) and
-- the remotes (brief presses).
--
-- Discriminator: how spread out the samples are within their 5-min
-- aggregation window.
--   window_ms >  120000 (>2 min) → sustained → TV-idle broadcaster
--   window_ms <=  60000 (<1 min) → time-clustered → remote press burst
--   in between                   → ambiguous
--
-- (Sample-count alone is fooled — 10 quick remote presses can produce as
-- many samples as a TV pinging every 20s for 5 min.)
--
-- Location uses MEDIAN of per-window RSSI samples — robust against the
-- wide RSSI spread caused by transient body-blocking or interference.

SET VARIABLE firetv_uuid = '3e1d50cd7e3e427d8e1cb78aa87fe624';

WITH firetv AS (
    SELECT
        oj.first_seen,
        oj.address,
        o.sample_count,
        o.rssi_mean,
        o.rssi_min,
        o.rssi_max,
        o.window_ms,
        list_aggregate(o.rssi_samples, 'median')::INT AS rssi_median,
        CASE
            WHEN o.window_ms > 120000 THEN 'TV idle (sustained)'
            WHEN o.window_ms <  60000 THEN 'remote press (burst)'
            ELSE                            'ambiguous'
        END                       AS role,
        CASE
            WHEN list_aggregate(o.rssi_samples, 'median')::INT > -65 THEN 'office (scanner room)'
            WHEN list_aggregate(o.rssi_samples, 'median')::INT > -80 THEN 'directly downstairs'
            WHEN list_aggregate(o.rssi_samples, 'median')::INT > -95 THEN 'downstairs far'
            ELSE                                                         'outside / next door'
        END                       AS location
    FROM obs o
    JOIN obs_json oj
        ON oj.address = o.address AND oj.first_seen = o.first_seen
    WHERE list_contains(oj.svc_uuids, getvariable('firetv_uuid'))
)
SELECT
    location,
    role,
    count(*)                          AS rows,
    count(DISTINCT address)           AS unique_macs,
    sum(sample_count)                 AS samples,
    round(avg(rssi_median), 0)::INT   AS avg_median_rssi,
    min(rssi_min)                     AS rssi_min,
    max(rssi_max)                     AS rssi_max,
    round(avg(window_ms / 1000.0), 0) AS avg_span_s,
    min(first_seen)                   AS first_seen,
    max(first_seen)                   AS last_seen
FROM firetv
GROUP BY location, role
ORDER BY location, role;
