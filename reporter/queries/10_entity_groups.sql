-- Fingerprints that link multiple BLE addresses. This is "MAC rotation
-- defeated" — for each persistent identifier, list every address that
-- broadcast it. If mac_count > 1, the device is rotating its BLE MAC but
-- leaking a stable payload-level identifier.

SELECT
    fp_type,
    fp_value,
    count(DISTINCT address)             AS mac_count,
    array_agg(DISTINCT address)         AS addresses,
    min(first_seen)                     AS first_seen,
    max(first_seen)                     AS last_seen
FROM fingerprints
GROUP BY fp_type, fp_value
HAVING count(DISTINCT address) > 1
ORDER BY mac_count DESC, fp_type
LIMIT 50;
