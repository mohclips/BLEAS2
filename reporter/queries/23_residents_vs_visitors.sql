-- Heuristic classifier:
--   resident  — same MAC observed across many distinct hour-buckets and
--               with consistent RSSI (stationary nearby device)
--   regular   — appears in several hour-buckets but with varying RSSI
--               (a person who comes and goes)
--   transient — observed in only 1-2 buckets (passers-by, freshly rotated
--               Apple MACs)
--
-- The thresholds are intentionally simple; tune them once you have a feel
-- for the dataset.

WITH per_addr AS (
    SELECT
        address,
        coalesce(any_value(mfr_name), '(none)')                            AS manufacturer,
        any_value(address_type)                                            AS address_type,
        count(DISTINCT date_trunc('hour', first_seen))                     AS hours_seen,
        count(*)                                                           AS windows_seen,
        sum(sample_count)                                                  AS total_samples,
        round(avg(rssi_mean)::DOUBLE, 1)                                   AS avg_rssi,
        round(stddev_pop(rssi_mean)::DOUBLE, 1)                            AS rssi_stddev,
        min(first_seen)                                                    AS first_seen,
        max(last_seen)                                                     AS last_seen
    FROM obs
    GROUP BY address
)
SELECT
    CASE
        WHEN hours_seen >= 4 AND coalesce(rssi_stddev, 0) < 3 THEN 'resident'
        WHEN hours_seen >= 4                                  THEN 'regular'
        WHEN hours_seen >= 2                                  THEN 'recurring'
        ELSE 'transient'
    END                                                                    AS classification,
    count(*)                                                               AS addrs,
    sum(windows_seen)                                                      AS total_windows,
    round(avg(hours_seen)::DOUBLE, 1)                                      AS avg_hours_seen
FROM per_addr
GROUP BY classification
ORDER BY classification;
