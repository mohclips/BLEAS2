-- Capture summary: span, observation count, unique addresses, address-type
-- breakdown. Gives the at-a-glance shape of a capture run.

SELECT 'observations'    AS metric, count(*)::VARCHAR AS value FROM obs
UNION ALL
SELECT 'unique_addrs',     count(DISTINCT address)::VARCHAR FROM obs
UNION ALL
SELECT 'earliest_seen',    min(first_seen)::VARCHAR FROM obs
UNION ALL
SELECT 'latest_seen',      max(last_seen)::VARCHAR FROM obs
UNION ALL
SELECT 'span',             age(max(last_seen), min(first_seen))::VARCHAR FROM obs
UNION ALL
SELECT 'total_samples',    sum(sample_count)::VARCHAR FROM obs;

SELECT address_type,
       count(*) AS observations,
       count(DISTINCT address) AS unique_addrs
FROM obs
GROUP BY address_type
ORDER BY observations DESC;
