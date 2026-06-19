-- Surfaces every undocumented Apple subtype seen in the capture, with a
-- representative payload. Use this to guide further research: which subtypes
-- deserve a real parser, which look transient and ignorable.

SELECT
    subtype,
    count(*)                        AS observations,
    count(DISTINCT address)         AS unique_addrs,
    min(first_seen)                 AS first_seen,
    max(first_seen)                 AS last_seen,
    any_value(payload)::VARCHAR     AS sample_payload
FROM (
    SELECT
        address,
        first_seen,
        je.key   AS subtype,
        je.value AS payload
    FROM obs_json, json_each(mfr_details_json, '$.apple') je
    WHERE je.value IS NOT NULL AND je.value != 'null'
)
WHERE subtype LIKE 'unknown_%'
GROUP BY subtype
ORDER BY observations DESC;
