-- TPMS (tire pressure sensor) candidate detector.
--
-- TPMS broadcasts share a fingerprint that's pretty distinct:
--   * random MAC (often static random — rotates only on power cycle)
--   * NO manufacturer ID, NO service UUIDs, NO local name (cheap chipsets
--     just emit flags + raw payload bytes)
--   * broadcast at roughly 1 Hz while the vehicle is in motion or
--     recently powered
--   * appear in CLUSTERS — N sensors per vehicle (4 for cars/vans, 6 for
--     larger trucks) all transmit simultaneously when the vehicle is near
--
-- This query surfaces records matching the chipset fingerprint and groups
-- them by time-cluster so 4-wheel batches show up together. A burst of
-- 3-6 such MACs within a single 5-min window almost certainly = one
-- vehicle in/leaving range.

WITH candidates AS (
    SELECT
        oj.first_seen,
        oj.last_seen,
        oj.address,
        o.sample_count,
        round(o.window_ms / 1000.0, 0)  AS window_s,
        o.rssi_max,
        o.rssi_min,
        round(o.rssi_mean, 0)::INT      AS rssi_avg,
        -- Bucket sensors that arrived within ~5 minutes into the same
        -- "arrival cluster" — TPMS units come and go together.
        date_trunc('minute', oj.first_seen) -
            INTERVAL (extract('minute' FROM oj.first_seen)::INT % 5) MINUTE
                                        AS cluster_5min
    FROM obs o
    JOIN obs_json oj
        ON oj.address = o.address AND oj.first_seen = o.first_seen
    WHERE
        o.address_type = 'random'
        AND o.mfr_id IS NULL
        AND coalesce(o.local_name, '') = ''
        AND (oj.svc_uuids IS NULL OR len(oj.svc_uuids) = 0)
        AND o.sample_count >= 10        -- sustained, not a single blip
        AND o.window_ms BETWEEN 10000 AND 300000   -- transient (~10s-5min)
        AND o.rssi_max > -90            -- in / near the house
)
SELECT
    cluster_5min                        AS arrival_window,
    count(*)                            AS sensors_in_burst,
    list(address)                       AS macs,
    sum(sample_count)                   AS total_samples,
    min(rssi_min)                       AS rssi_min,
    max(rssi_max)                       AS rssi_max,
    round(avg(rssi_avg), 0)::INT        AS avg_rssi,
    round(avg(window_s), 0)             AS avg_dwell_s,
    min(first_seen)                     AS first_seen,
    max(last_seen)                      AS last_seen
FROM candidates
GROUP BY cluster_5min
HAVING count(*) >= 3                    -- 3+ matching MACs in the same window = vehicle
ORDER BY arrival_window DESC;
