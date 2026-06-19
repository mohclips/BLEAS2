-- Quick overall view of what linkable identifiers are present in the
-- capture and how widely they spread across addresses.

SELECT
    fp_type,
    lifetime,
    count(*)                             AS observations,
    count(DISTINCT address)              AS distinct_addrs,
    count(DISTINCT fp_value)             AS distinct_fingerprints,
    -- ratio < 1 = at least one fingerprint observed from >1 MAC (linkage win)
    round(count(DISTINCT fp_value)::DOUBLE / count(DISTINCT address), 2) AS fp_per_mac
FROM fingerprints
GROUP BY fp_type, lifetime
ORDER BY observations DESC;
