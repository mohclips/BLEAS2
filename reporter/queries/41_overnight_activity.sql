-- Overnight BLE activity (default: 23:00-06:00 local, treating BST=UTC+1
-- so the captured timestamps are 22:00-05:00 UTC). Adjust the variables
-- below to scan a different night or a different window.
--
-- Useful for finding:
--   * What's *always* broadcasting (TVs in standby, smart speakers,
--     smart bulbs, smart-home hubs, the AC if it stays on)
--   * What suddenly *starts* overnight (alarm-driven devices, scheduled
--     vacuum / appliance, smart bulbs turning on for security)
--   * Passing late-night traffic — much quieter than daytime, so the
--     few cars that do appear stand out
--   * Sleeping family-member phones — Android Find My continues at night,
--     iPhones reduce broadcast frequency but still emit

SET VARIABLE start_utc = '2026-06-17T22:00:00Z';   -- = 23:00 BST yesterday
SET VARIABLE end_utc   = '2026-06-18T05:00:00Z';   -- = 06:00 BST today

WITH overnight AS (
    SELECT *
    FROM obs
    WHERE first_seen >= getvariable('start_utc')::TIMESTAMP
      AND first_seen <  getvariable('end_utc')::TIMESTAMP
),
-- Pull Google Find My Device traffic aside — 3 family phones each rotate
-- RPAs every ~15 min, so 7 hours of overnight produces ~84 noise rows
-- that drown out actual stationary devices. Aggregate them into one row
-- so the rest of the result is readable.
findmy_macs AS (
    SELECT DISTINCT oj.address
    FROM obs_json oj
    -- Google Find My Device data is published under service-data UUID 0xfcf1,
    -- which lands in our details dict under the key `service_fcf1` (no
    -- dedicated parser). Check there rather than the svc_uuids list.
    WHERE json_extract(oj.svc_details_json, '$.service_fcf1') IS NOT NULL
      AND oj.first_seen >= getvariable('start_utc')::TIMESTAMP
      AND oj.first_seen <  getvariable('end_utc')::TIMESTAMP
),
findmy_summary AS (
    SELECT
        'FIND-MY-COMBINED'                AS address,
        'Google Find My Device (3 family phones)' AS manufacturer,
        'Google LLC'                      AS all_manufacturers,
        ''                                AS local_name,
        'random'                          AS address_type,
        count(*)                          AS windows,
        sum(sample_count)                 AS samples,
        round(avg(rssi_mean), 0)::INT     AS avg_rssi,
        min(rssi_min)                     AS rssi_min,
        max(rssi_max)                     AS rssi_max,
        0.0                               AS rssi_stddev,
        min(first_seen)                   AS first_seen,
        max(last_seen)                    AS last_seen,
        count(DISTINCT date_trunc('hour', first_seen)) AS hours_active,
        count(DISTINCT o.address)         AS unique_rpas
    FROM overnight o
    JOIN findmy_macs fm ON fm.address = o.address
),
per_addr AS (
    SELECT
        o.address,
        any_value(o.mfr_name)              AS manufacturer,
        -- All distinct mfr names this MAC has broadcast in the window.
        -- A device that shows up as e.g. "Apple, Inc., Sony Group Corp."
        -- is dual-broadcasting (Sony Bravia TVs do this — they speak
        -- iBeacon + Apple Continuity for AirPlay 2 / HomeKit *and*
        -- Sony's own mfr 0x012d. Without this column the "any_value"
        -- in the manufacturer field hides the dual-stack identity.)
        string_agg(DISTINCT o.mfr_name, ', ' ORDER BY o.mfr_name)
            FILTER (WHERE o.mfr_name IS NOT NULL)  AS all_manufacturers,
        any_value(o.local_name)            AS local_name,
        o.address_type,
        count(*)                           AS windows,
        sum(o.sample_count)                AS samples,
        round(avg(o.rssi_mean), 0)::INT    AS avg_rssi,
        min(o.rssi_min)                    AS rssi_min,
        max(o.rssi_max)                    AS rssi_max,
        round(coalesce(stddev_pop(o.rssi_mean), 0), 1) AS rssi_stddev,
        min(o.first_seen)                  AS first_seen,
        max(o.last_seen)                   AS last_seen,
        count(DISTINCT date_trunc('hour', o.first_seen)) AS hours_active,
        NULL::BIGINT                       AS unique_rpas
    FROM overnight o
    LEFT JOIN findmy_macs fm ON fm.address = o.address
    WHERE fm.address IS NULL                -- exclude individual Find My RPAs
    GROUP BY o.address, o.address_type
)
SELECT
    address,
    coalesce(manufacturer, '(none)') AS manufacturer,
    coalesce(all_manufacturers, '')  AS all_mfrs,
    coalesce(local_name, '')         AS local_name,
    address_type                     AS addr_type,
    windows,
    samples,
    coalesce(unique_rpas::VARCHAR, '') AS unique_rpas,
    avg_rssi,
    rssi_min,
    rssi_max,
    rssi_stddev,
    hours_active,
    CASE
        WHEN avg_rssi > -65 THEN 'office'
        WHEN avg_rssi > -80 THEN 'directly downstairs'
        WHEN avg_rssi > -94 THEN 'downstairs far'
        ELSE                     'outside / next door'
    END                              AS location,
    CASE
        WHEN address = 'FIND-MY-COMBINED'          THEN 'family phones (Find My)'
        WHEN hours_active >= 6 AND rssi_stddev < 5 THEN 'always-on (stationary)'
        WHEN hours_active >= 3                     THEN 'recurring overnight'
        WHEN windows = 1                           THEN 'one-off (passer-by?)'
        ELSE                                            'intermittent'
    END                              AS pattern,
    first_seen,
    last_seen
FROM (
    SELECT * FROM per_addr
    UNION ALL
    SELECT * FROM findmy_summary
)
ORDER BY samples DESC;
