-- The "nearby" advertisement also carries a status bitmask: AirPods
-- connected, WiFi on, Apple Watch locked, Auto-Unlock enabled, etc.
--
-- This query unnests the mask array (already decoded by the parser) and
-- shows how often each flag appears per address.

WITH masks AS (
    SELECT
        address,
        unnest(from_json(json_extract(mfr_details_json, '$.apple.nearby.masks'), '["VARCHAR"]')) AS flag
    FROM obs_json
    WHERE json_extract(mfr_details_json, '$.apple.nearby.masks') IS NOT NULL
      AND json_extract(mfr_details_json, '$.apple.nearby.masks') != 'null'
)
SELECT address, flag, count(*) AS observations
FROM masks
GROUP BY address, flag
ORDER BY observations DESC;
