-- Find Chipolo Bluetooth trackers in the capture.
--
-- Chipolo advertises under service UUIDs 0xfe65 (primary) and 0xfe33
-- (secondary). The 0xfe33 service data carries the device's own BD_ADDR
-- echoed in the payload — that's why the tracker can be identified even
-- if MAC randomization were in play.
--
-- This query covers BOTH the new parsed shape ($.chipolo.*) and the older
-- raw shape ($.service_fe33.*) so captures from before the Chipolo parser
-- was added still surface.

WITH new_shape AS (
    SELECT
        address,
        json_extract_string(svc_details_json, '$.chipolo.echo_mac')  AS echo_mac,
        json_extract_string(svc_details_json, '$.chipolo.uuid')      AS service_uuid,
        json_extract_string(svc_details_json, '$.chipolo.header')    AS header_hex,
        first_seen, last_seen, sample_count
    FROM obs_json
    WHERE json_extract(svc_details_json, '$.chipolo') IS NOT NULL
),
old_shape AS (
    SELECT
        address,
        NULL                                                                                 AS echo_mac,
        json_extract_string(svc_details_json, '$.service_fe33.uuid')                         AS service_uuid,
        substring(json_extract_string(svc_details_json, '$.service_fe33.data'), 1, 8)        AS header_hex,
        first_seen, last_seen, sample_count
    FROM obs_json
    WHERE json_extract(svc_details_json, '$.service_fe33') IS NOT NULL
)
SELECT
    address                        AS broadcast_mac,
    coalesce(echo_mac, '(raw)')    AS echo_mac,
    coalesce(service_uuid, 'fe33') AS service_uuid,
    header_hex,
    count(*)                       AS windows,
    sum(sample_count)              AS samples,
    min(first_seen)                AS first_seen,
    max(last_seen)                 AS last_seen
FROM (SELECT * FROM new_shape UNION ALL SELECT * FROM old_shape)
GROUP BY broadcast_mac, echo_mac, service_uuid, header_hex
ORDER BY windows DESC;
