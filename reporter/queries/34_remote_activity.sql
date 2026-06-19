-- Firestick remote (and other minimal-broadcast BLE peripheral) activity
-- log. Each row = one press / wake event = one moment someone interacted
-- with a paired BLE device.
--
-- Three fingerprints supported (Firestick generations + other minimal
-- peripherals):
--   * random MAC, just Flags AD field (older voice remotes)
--   * random MAC, 128-bit service UUID + short local name fragment
--      (newer voice remotes — Amazon's HID-over-GATT)
--   * public MAC, Amazon manufacturer + HID 16-bit service 0x1812
--      (some Fire TV remotes use static addresses)
--
-- RSSI tells you which room — calibrate by pressing each remote in turn
-- and noting the band.

SELECT
    oj.first_seen                                 AS ts,
    oj.address                                    AS mac,
    oj.address_type                               AS addr_type,
    coalesce(o.local_name, '')                    AS local_name,
    coalesce(oj.mfr_name, '(none)')               AS manufacturer,
    coalesce(array_to_string(oj.svc_uuids, ','), '') AS service_uuids,
    o.sample_count                                AS samples,
    o.rssi_max                                    AS rssi_max,
    o.rssi_min                                    AS rssi_min,
    round(o.rssi_mean, 0)::INT                    AS rssi_avg,
    CASE
        WHEN o.rssi_mean > -65 THEN 'office (scanner room)'
        WHEN o.rssi_mean > -80 THEN 'directly downstairs (lounge / dining?)'
        WHEN o.rssi_mean > -94 THEN 'downstairs far (dining far / downstairs office?)'
        ELSE                       'outside / next door'
    END                                           AS room_guess
FROM obs o
JOIN obs_json oj ON oj.address = o.address AND oj.first_seen = o.first_seen
WHERE
    -- minimal-broadcast random-MAC (most older Firestick remotes + privacy-minimal devices)
    (
        o.address_type = 'random'
        AND o.mfr_id IS NULL
        AND o.sample_count < 50
    )
    OR
    -- Amazon-manufactured devices (Firestick remotes that use public MACs)
    (
        oj.mfr_name LIKE 'Amazon%'
    )
    OR
    -- BLE HID devices (keyboards / remotes broadcasting service 0x1812)
    (
        oj.svc_uuids IS NOT NULL
        AND list_contains(oj.svc_uuids, '1812')
    )
ORDER BY ts DESC
LIMIT 100;
