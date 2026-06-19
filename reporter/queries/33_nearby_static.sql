-- Nearby static devices: persistent broadcasters that don't move and have
-- strong-enough signal to be inside the house (or right outside the wall).
--
-- Three filters AND'd together:
--   * proximity: avg RSSI > -94 dBm (inside the house — calibrated against
--                                    a remote press from the downstairs
--                                    office at far end of bungalow which
--                                    lands at -86 to -92)
--   * static  : stddev of per-window mean RSSI < 5 dBm
--                (RSSI within a single packet is noisy by nature — measuring
--                 variance across window means is what tracks actual motion)
--   * present : seen in 2 or more distinct dedup windows
--
-- That gives you "stuff that lives at this address" — your AC, smart
-- bulbs, beacons, fixed neighbour devices through a wall. Movers, walk-bys
-- and one-shot passers fall out.

SELECT
    address,
    coalesce(mfr_name, '(none)')                 AS manufacturer,
    address_type,
    coalesce(local_name, '')                     AS local_name,
    count(*)                                     AS windows,
    sum(sample_count)                            AS samples,
    round(avg(rssi_mean), 1)                     AS avg_rssi,
    round(coalesce(stddev_pop(rssi_mean), 0), 1) AS rssi_stddev,
    min(rssi_min)                                AS rssi_min,
    max(rssi_max)                                AS rssi_max,
    (max(rssi_max) - min(rssi_min))              AS rssi_span,
    min(first_seen)                              AS first_seen,
    max(last_seen)                               AS last_seen,
    -- Bands tuned for this UK dormer bungalow with scanner in upstairs
    -- office. Reference: far-downstairs office press = -86 to -92 dBm.
    CASE
        WHEN avg(rssi_mean) > -65 THEN 'office (scanner room)'
        WHEN avg(rssi_mean) > -80 THEN 'upstairs / directly downstairs'
        ELSE                            'downstairs (mid/far)'
    END                                          AS proximity
FROM obs
GROUP BY address, mfr_name, address_type, local_name
HAVING
        count(*) >= 2
    AND avg(rssi_mean) > -94
    AND coalesce(stddev_pop(rssi_mean), 0) < 5
ORDER BY avg_rssi DESC;
