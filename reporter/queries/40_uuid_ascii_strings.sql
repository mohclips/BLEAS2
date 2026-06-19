-- Hidden ASCII in 128-bit service UUIDs. Some vendors encode the product
-- or company name as ASCII in the first bytes of their custom UUID — the
-- "BYD AUTO" sighting earlier was found that way (`425944204155544f...`
-- decodes to `BYD AUTO` then random suffix).
--
-- This query scans every 128-bit UUID seen in the captures, extracts
-- substrings of 3+ consecutive bytes that fall in the printable ASCII
-- letter/digit/space range (0x20, 0x30-0x39, 0x41-0x5A, 0x61-0x7A), and
-- surfaces the decoded string alongside the addresses that broadcast it.
--
-- Common findings: vendor names ("BYD", "GOOGLE"), product codes,
-- developer initials. Random UUIDv4s rarely match (the regex requires
-- 3+ consecutive printable bytes which is improbable in pure random).

WITH all_uuids AS (
    SELECT DISTINCT u.uuid AS svc_uuid
    FROM obs_json oj, UNNEST(oj.svc_uuids) AS u(uuid)
    WHERE length(u.uuid) > 4
),
matches AS (
    SELECT
        svc_uuid,
        -- Pairs of hex chars matching the printable letter/digit/space range:
        --   20      = ' '
        --   3[0-9]  = '0'-'9'
        --   4[1-9a-f] = 'A'-'O'
        --   5[0-9a] = 'P'-'Z'
        --   6[1-9a-f] = 'a'-'o'
        --   7[0-9a] = 'p'-'z'
        regexp_extract(
            svc_uuid,
            '((?:20|3[0-9]|4[1-9a-f]|5[0-9a]|6[1-9a-f]|7[0-9a]){3,})',
            1
        ) AS letter_hex
    FROM all_uuids
),
decoded AS (
    SELECT
        svc_uuid,
        letter_hex,
        TRY_CAST(from_hex(letter_hex) AS VARCHAR) AS decoded_ascii
    FROM matches
    WHERE letter_hex IS NOT NULL AND letter_hex != ''
),
with_addrs AS (
    SELECT
        d.svc_uuid,
        d.decoded_ascii,
        count(*)                  AS records,
        count(DISTINCT oj.address) AS unique_macs,
        any_value(oj.address)     AS sample_mac,
        any_value(oj.mfr_name)    AS sample_mfr
    FROM decoded d
    JOIN obs_json oj
        ON list_contains(oj.svc_uuids, d.svc_uuid)
    GROUP BY d.svc_uuid, d.decoded_ascii
)
SELECT *
FROM with_addrs
ORDER BY length(decoded_ascii) DESC, records DESC;
