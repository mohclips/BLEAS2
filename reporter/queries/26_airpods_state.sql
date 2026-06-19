-- AirPods state telemetry: battery levels, charging status, lid open/closed.
-- When someone walks past wearing AirPods you can see:
--   - what model (AirPods Pro, AirPods 2, etc.)
--   - whether they're in the case or in the ear
--   - battery percentage of each earbud
--   - whether the case is charging
--
-- Lid status changes are particularly telling: open lid = user just took
-- earbuds out / putting them on.

SELECT
    address,
    json_extract_string(mfr_details_json, '$.apple.airpods.device_model') AS model,
    json_extract(mfr_details_json, '$.apple.airpods._lid')                AS lid_byte,
    json_extract(mfr_details_json, '$.apple.airpods.battery.batteryR')    AS battery_right,
    json_extract(mfr_details_json, '$.apple.airpods.battery.batteryL')    AS battery_left,
    json_extract(mfr_details_json, '$.apple.airpods.charging.C')          AS case_charging,
    json_extract(mfr_details_json, '$.apple.airpods.case_power')          AS case_power,
    first_seen, sample_count
FROM obs_json
WHERE json_extract_string(mfr_details_json, '$.apple.airpods.device_model') IS NOT NULL
ORDER BY first_seen DESC;
