-- Drill down on a single service UUID. Each row = one broadcast burst
-- (one window the aggregator saw). RPAs rotate, so the same physical
-- device shows up with different MACs over time — but the RSSI band
-- clusters by room, which lets you attribute each burst to a specific
-- physical device.
--
-- Default UUID: 0x3e1d50cd... = Amazon Fire TV remote (Style B). To inspect
-- a different UUID, edit the value below to whichever you saw in
-- ./run.sh 35.

SET VARIABLE target_uuid = '3e1d50cd7e3e427d8e1cb78aa87fe624';

SELECT
    oj.first_seen                          AS ts,
    oj.address                             AS mac,
    oj.address_type                        AS addr_type,
    o.sample_count                         AS samples,
    o.rssi_max                             AS rssi_max,
    o.rssi_min                             AS rssi_min,
    round(o.rssi_mean, 0)::INT             AS rssi_avg,
    CASE
        WHEN o.rssi_mean > -65 THEN 'office (scanner room)'
        WHEN o.rssi_mean > -80 THEN 'directly downstairs'
        WHEN o.rssi_mean > -94 THEN 'downstairs far'
        ELSE                       'outside / next door'
    END                                    AS location,
    coalesce(o.local_name, '')             AS local_name,
    coalesce(oj.mfr_name, '(none)')        AS manufacturer
FROM obs o
JOIN obs_json oj
    ON oj.address = o.address
   AND oj.first_seen = o.first_seen
WHERE list_contains(oj.svc_uuids, getvariable('target_uuid'))
ORDER BY ts;
