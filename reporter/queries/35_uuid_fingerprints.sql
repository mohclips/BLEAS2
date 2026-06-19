-- Group BLE observations by service UUID, not by MAC. This defeats MAC
-- randomisation — devices with rotating RPAs that broadcast a stable
-- service UUID get collapsed into a single row per UUID, with the count
-- of distinct MACs showing how aggressive the rotation is.
--
-- Standard pattern to spot:
--   * 1 UUID broadcast by 1 MAC, lots of samples → stable beacon, no
--     rotation (Chipolo, public-MAC Firestick remote, smart lock, etc.)
--   * 1 UUID broadcast by N>5 MACs → rotation in flight; the UUID is the
--     real identity (newer Fire TV remotes, AirTag-style Find-My-network
--     devices, some Samsung trackers)
--   * 1 UUID broadcast by N=2..5 MACs → probably the same device caught
--     across a couple of rotation cycles in your capture window

WITH uuid_obs AS (
    SELECT
        oj.first_seen,
        oj.address,
        oj.address_type,
        u.uuid                                   AS svc_uuid,
        o.rssi_mean,
        o.rssi_min,
        o.rssi_max,
        o.sample_count,
        o.local_name,
        oj.mfr_name
    FROM obs o
    JOIN obs_json oj
        ON oj.address = o.address AND oj.first_seen = o.first_seen,
    UNNEST(oj.svc_uuids) AS u(uuid)
    WHERE oj.svc_uuids IS NOT NULL
),
agg AS (
    SELECT
        svc_uuid,
        count(*)                                 AS observations,
        count(DISTINCT address)                  AS unique_macs,
        sum(sample_count)                        AS samples,
        round(avg(rssi_mean), 0)::INT            AS avg_rssi,
        min(rssi_min)                            AS rssi_min,
        max(rssi_max)                            AS rssi_max,
        any_value(local_name)                    AS sample_local_name,
        any_value(mfr_name)                      AS sample_mfr,
        any_value(address_type)                  AS sample_addr_type,
        min(first_seen)                          AS first_seen,
        max(first_seen)                          AS last_seen
    FROM uuid_obs
    GROUP BY svc_uuid
)
SELECT
    svc_uuid,
    coalesce(n.vendor_name, '(no SIG name)')     AS vendor,
    CASE
        WHEN length(svc_uuid) > 4 THEN '128-bit'
        ELSE '16-bit'
    END                                          AS uuid_bits,
    CASE
        WHEN unique_macs >= 5 THEN 'rotating (RPA)'
        WHEN unique_macs >  1 THEN 'multi-MAC'
        ELSE                       'single MAC'
    END                                          AS mac_pattern,
    unique_macs                                  AS macs_seen,
    observations,
    samples,
    avg_rssi,
    rssi_min,
    rssi_max,
    CASE
        WHEN avg_rssi > -65 THEN 'office (scanner room)'
        WHEN avg_rssi > -80 THEN 'directly downstairs'
        WHEN avg_rssi > -94 THEN 'downstairs far'
        ELSE                     'outside / next door'
    END                                          AS location,
    coalesce(sample_local_name, '')              AS local_name,
    coalesce(sample_mfr, '(none)')               AS manufacturer,
    sample_addr_type                             AS addr_type,
    first_seen,
    last_seen
FROM agg
LEFT JOIN service_uuid_names n ON n.uuid = agg.svc_uuid
ORDER BY observations DESC;
