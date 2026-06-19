-- Apple "nearby" advertisements carry an activity-state byte that leaks
-- what the user is doing right now: locked phone, audio playing, phone
-- call, driving (CarPlay), etc. This query counts how many observations
-- carried each state, broken out per address.
--
-- One row per (address, state). Use ORDER BY observations DESC to see the
-- most common states first.

SELECT
    address,
    json_extract_string(mfr_details_json, '$.apple.nearby.state') AS state,
    count(*)         AS observations,
    sum(sample_count) AS samples
FROM obs_json
WHERE json_extract_string(mfr_details_json, '$.apple.nearby.state') IS NOT NULL
GROUP BY address, state
ORDER BY observations DESC;
