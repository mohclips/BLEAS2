-- Apple Continuity subtype usage. Counts how many observations carried each
-- subtype across the capture. Useful for:
--   - confirming parsers are firing (no subtype should be a surprise)
--   - inferring device behaviour (a flurry of hey_siri = wake-word fires;
--     handoff bursts = clipboard activity)
--
-- The DuckDB struct schema is the union of every key ever seen, so each
-- record carries NULL placeholders for subtypes it doesn't actually have.
-- We filter those out so the counts reflect only real emissions.

SELECT
    subtype,
    count(*)                AS observations,
    count(DISTINCT address) AS unique_addrs
FROM (
    SELECT
        address,
        je.key   AS subtype,
        je.value AS payload
    FROM obs_json, json_each(mfr_details_json, '$.apple') je
    WHERE je.value IS NOT NULL AND je.value != 'null'
)
GROUP BY subtype
ORDER BY observations DESC;
