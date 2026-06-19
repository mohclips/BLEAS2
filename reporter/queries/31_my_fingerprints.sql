-- Identity fingerprints ranked by persistence. Your own devices live in
-- your environment all day, so their stable identifiers appear in many
-- distinct time windows. Strangers passing by produce one or two windows.
--
-- Copy the high-persistence rows into exclude.airdrop_hashes etc. in
-- cmd/scanner/config.yml to filter your own devices out of future captures
-- — even when they rotate their BLE MAC.

WITH airdrop AS (
    SELECT
        'airdrop_tuple' AS kind,
        concat_ws(':',
            json_extract_string(mfr_details_json, '$.apple.airdrop.apple_id_hash'),
            json_extract_string(mfr_details_json, '$.apple.airdrop.phone_hash'),
            json_extract_string(mfr_details_json, '$.apple.airdrop.email_hash'),
            json_extract_string(mfr_details_json, '$.apple.airdrop.email2_hash')
        ) AS fingerprint,
        address, first_seen
    FROM obs_json
    WHERE json_extract_string(mfr_details_json, '$.apple.airdrop.apple_id_hash') IS NOT NULL
),
icloud AS (
    SELECT
        'icloud_id' AS kind,
        json_extract_string(mfr_details_json, '$.apple.tethering_target.icloud_id') AS fingerprint,
        address, first_seen
    FROM obs_json
    WHERE json_extract_string(mfr_details_json, '$.apple.tethering_target.icloud_id') IS NOT NULL
),
homekit AS (
    SELECT
        'homekit_device_id' AS kind,
        json_extract_string(mfr_details_json, '$.apple.homekit.device_id') AS fingerprint,
        address, first_seen
    FROM obs_json
    WHERE json_extract_string(mfr_details_json, '$.apple.homekit.device_id') IS NOT NULL
),
findmy AS (
    SELECT
        'findmy_public_key' AS kind,
        json_extract_string(mfr_details_json, '$.apple.findmy.public_key') AS fingerprint,
        address, first_seen
    FROM obs_json
    WHERE json_extract_string(mfr_details_json, '$.apple.findmy.variant') = 'separated'
)
SELECT
    kind,
    fingerprint,
    count(*)                                          AS observations,
    count(DISTINCT address)                           AS distinct_macs,
    count(DISTINCT date_trunc('hour', first_seen))    AS hours_seen,
    min(first_seen)                                   AS first_seen,
    max(first_seen)                                   AS last_seen
FROM (
    SELECT * FROM airdrop
    UNION ALL SELECT * FROM icloud
    UNION ALL SELECT * FROM homekit
    UNION ALL SELECT * FROM findmy
)
GROUP BY kind, fingerprint
ORDER BY observations DESC, hours_seen DESC;
