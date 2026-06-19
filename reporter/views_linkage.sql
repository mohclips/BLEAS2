-- Linkage views.
--
-- Pulls stable / semi-stable identifiers out of the parsed payloads and
-- exposes them as a single `fingerprints` table:
--
--   (address, first_seen, fp_type, fp_value, lifetime)
--
-- One row per (observation, fingerprint). A single observation can produce
-- multiple rows if its payload carries several linkable fields.
--
-- Lifetime:
--   permanent  — fingerprint never changes (HomeKit accessory ID,
--                Apple ID hashes, beacon UUID).
--   daily      — fingerprint rotates with the calendar day (iCloud DSID hash,
--                Find My public key).
--
-- Loads on top of views.sql (so obs / obs_json are already available).

CREATE OR REPLACE VIEW fp_homekit AS
SELECT
    address,
    first_seen,
    'homekit'                                                                AS fp_type,
    json_extract_string(mfr_details_json, '$.apple.homekit.device_id')       AS fp_value,
    'permanent'                                                              AS lifetime
FROM obs_json
WHERE json_extract_string(mfr_details_json, '$.apple.homekit.device_id') IS NOT NULL;

CREATE OR REPLACE VIEW fp_ibeacon AS
SELECT
    address,
    first_seen,
    'ibeacon'                                                                AS fp_type,
    json_extract_string(mfr_details_json, '$.apple.ibeacon.uuid')            AS fp_value,
    'permanent'                                                              AS lifetime
FROM obs_json
WHERE json_extract_string(mfr_details_json, '$.apple.ibeacon.uuid') IS NOT NULL;

CREATE OR REPLACE VIEW fp_airdrop AS
SELECT
    address,
    first_seen,
    'airdrop'                                                                AS fp_type,
    concat_ws(':',
        json_extract_string(mfr_details_json, '$.apple.airdrop.apple_id_hash'),
        json_extract_string(mfr_details_json, '$.apple.airdrop.phone_hash'),
        json_extract_string(mfr_details_json, '$.apple.airdrop.email_hash'),
        json_extract_string(mfr_details_json, '$.apple.airdrop.email2_hash')
    )                                                                        AS fp_value,
    'permanent'                                                              AS lifetime
FROM obs_json
WHERE json_extract_string(mfr_details_json, '$.apple.airdrop.apple_id_hash') IS NOT NULL;

-- iCloud DSID hash from tethering_target rotates daily — bind the date into
-- the fingerprint so the same hash on different days doesn't get merged.
CREATE OR REPLACE VIEW fp_icloud AS
SELECT
    address,
    first_seen,
    'icloud'                                                                                AS fp_type,
    json_extract_string(mfr_details_json, '$.apple.tethering_target.icloud_id')
        || ':' || strftime(first_seen, '%Y-%m-%d')                                          AS fp_value,
    'daily'                                                                                 AS lifetime
FROM obs_json
WHERE json_extract_string(mfr_details_json, '$.apple.tethering_target.icloud_id') IS NOT NULL;

-- Find My separated-mode public key — rotates every 24h, identifies a
-- specific Find My accessory (AirTag, separated iPhone, etc.).
CREATE OR REPLACE VIEW fp_findmy AS
SELECT
    address,
    first_seen,
    'findmy'                                                                                AS fp_type,
    json_extract_string(mfr_details_json, '$.apple.findmy.public_key')
        || ':' || strftime(first_seen, '%Y-%m-%d')                                          AS fp_value,
    'daily'                                                                                 AS lifetime
FROM obs_json
WHERE json_extract_string(mfr_details_json, '$.apple.findmy.variant') = 'separated'
  AND json_extract_string(mfr_details_json, '$.apple.findmy.public_key') IS NOT NULL;

CREATE OR REPLACE VIEW fingerprints AS
SELECT * FROM fp_homekit
UNION ALL
SELECT * FROM fp_ibeacon
UNION ALL
SELECT * FROM fp_airdrop
UNION ALL
SELECT * FROM fp_icloud
UNION ALL
SELECT * FROM fp_findmy;
