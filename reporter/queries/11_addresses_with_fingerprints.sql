-- Inverse view: for each BLE address, which fingerprints did it broadcast?
-- An address that emits multiple fingerprint types is highly identifiable
-- (e.g. it broadcasts both HomeKit and AirDrop, so you know vendor +
-- accessory ID + Apple ID).

SELECT
    address,
    count(DISTINCT fp_type)                       AS fp_types,
    count(DISTINCT (fp_type, fp_value))           AS distinct_fps,
    list(DISTINCT fp_type ORDER BY fp_type)       AS types_seen,
    min(first_seen)                               AS first_seen,
    max(first_seen)                               AS last_seen
FROM fingerprints
GROUP BY address
ORDER BY distinct_fps DESC, address;
