-- HomeKit accessory inventory: every device broadcasting a HomeKit
-- advertisement, with its category (light bulb / door lock / camera / etc.).
-- HomeKit accessories don't rotate MAC, so each row is a stable device.
--
-- The "global_state_number" increments on every state change — leaking how
-- often the accessory is used. min/max within the capture is a usage
-- counter.

SELECT
    address,
    json_extract_string(mfr_details_json, '$.apple.homekit.device_id')     AS device_id,
    json_extract_string(mfr_details_json, '$.apple.homekit.category_name') AS category,
    min(first_seen)                                                        AS first_seen,
    max(first_seen)                                                        AS last_seen,
    count(*)                                                               AS windows,
    min(cast(json_extract(mfr_details_json, '$.apple.homekit.global_state_number') AS BIGINT)) AS state_min,
    max(cast(json_extract(mfr_details_json, '$.apple.homekit.global_state_number') AS BIGINT)) AS state_max
FROM obs_json
WHERE json_extract_string(mfr_details_json, '$.apple.homekit.device_id') IS NOT NULL
GROUP BY address, device_id, category
ORDER BY last_seen DESC;
