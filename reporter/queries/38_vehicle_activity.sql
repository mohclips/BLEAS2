-- Vehicle / dashcam activity timeline. Picks up:
--   * Garmin/Kenwood DRV-series dashcams ("DRV-AXXXW" local name)
--   * Anything else with a local name containing "CAR-BT" or "ŠKODA"
--   * Bluetooth car kits with HFP / A2DP service UUIDs (rare on BLE side)
--
-- Two interesting situations to distinguish at a busy-crossroads house:
--
--   A) Your own vehicle parked at / near the house — sustained presence,
--      stronger RSSI when truly in the driveway, repeated appearances
--      across the day. The dashcam is the proxy for "engine on".
--   B) A stranger's car stopped at the traffic lights — one short window,
--      weak RSSI, distinct local-name suffix, never seen again.
--
-- The verdict column applies a simple heuristic; tune as you collect more.

WITH vehicles AS (
    SELECT
        address,
        local_name,
        first_seen,
        last_seen,
        sample_count,
        rssi_min,
        rssi_max,
        rssi_mean,
        window_ms
    FROM obs
    WHERE local_name LIKE 'DRV-%' OR local_name LIKE 'CAR-BT-%' OR local_name LIKE '%ŠKODA%'
),
per_addr AS (
    SELECT
        address,
        any_value(local_name)             AS local_name,
        count(*)                          AS visits,
        sum(sample_count)                 AS total_samples,
        round(avg(window_ms / 1000.0), 0) AS avg_span_s,
        max(window_ms / 1000)             AS max_span_s,
        min(rssi_min)                     AS rssi_min,
        max(rssi_max)                     AS rssi_max,
        round(avg(rssi_mean), 0)::INT     AS avg_rssi,
        min(first_seen)                   AS first_seen,
        max(last_seen)                    AS last_seen
    FROM vehicles
    GROUP BY address
)
SELECT
    local_name,
    address,
    visits,
    total_samples,
    avg_span_s,
    max_span_s,
    rssi_min,
    rssi_max,
    avg_rssi,
    CASE
        WHEN visits >= 2 AND max_span_s >= 30   THEN 'YOURS — repeat visits + sustained presence'
        WHEN visits = 1 AND max_span_s >= 60    THEN 'YOURS — single but long dwell (parked nearby)'
        WHEN visits = 1 AND max_span_s <  30    THEN 'PASSING — brief one-off (traffic lights?)'
        ELSE                                         'AMBIGUOUS — needs more data'
    END                                          AS verdict,
    first_seen,
    last_seen
FROM per_addr
ORDER BY total_samples DESC;
