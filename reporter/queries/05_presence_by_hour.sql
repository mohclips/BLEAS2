-- Activity by hour of day — observation counts and distinct devices.
-- A busy thoroughfare will show clear commute peaks; a quiet area shows
-- a flat baseline of resident devices.

SELECT
    date_trunc('hour', first_seen) AS hour,
    count(*)                       AS observations,
    count(DISTINCT address)        AS unique_addrs,
    sum(sample_count)              AS samples
FROM obs
GROUP BY hour
ORDER BY hour;
