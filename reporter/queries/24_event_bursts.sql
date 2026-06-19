-- Apple subtype bursts that correspond to user actions:
--   hey_siri burst   → wake word triggered
--   handoff burst    → clipboard or handoff event between devices
--   airdrop burst    → user opened the share sheet looking for receivers
--
-- A "burst" here is any observation containing one of these subtypes.
-- Per-event timing isn't precise (the dedup window collapses runs), but
-- the first_seen on each row tells you when the event began.

WITH events AS (
    SELECT
        address, first_seen,
        CASE
            WHEN json_extract(mfr_details_json, '$.apple.hey_siri') IS NOT NULL
                AND json_extract(mfr_details_json, '$.apple.hey_siri') != 'null' THEN 'hey_siri'
            WHEN json_extract(mfr_details_json, '$.apple.handoff') IS NOT NULL
                AND json_extract(mfr_details_json, '$.apple.handoff') != 'null' THEN 'handoff'
            WHEN json_extract(mfr_details_json, '$.apple.airdrop') IS NOT NULL
                AND json_extract(mfr_details_json, '$.apple.airdrop') != 'null' THEN 'airdrop'
        END                                                                 AS event_type
    FROM obs_json
)
SELECT
    address,
    event_type,
    count(*)              AS events,
    min(first_seen)       AS first_seen,
    max(first_seen)       AS last_seen
FROM events
WHERE event_type IS NOT NULL
GROUP BY address, event_type
ORDER BY events DESC;
