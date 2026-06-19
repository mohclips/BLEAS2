-- Five-minute presence grid: for each address, which 5-minute slots was it
-- observed in? The output is "long" form (address, bucket, observations)
-- so you can pivot to a wide grid externally if desired.
--
-- Useful for spotting:
--   - residents: large block of contiguous buckets
--   - commute patterns: gaps + bursts at consistent times
--   - visitor MACs: a single bucket (often paired with rotation)

SELECT
    address,
    date_trunc('minute', first_seen) - INTERVAL (extract('minute' FROM first_seen)::INT % 5) MINUTE
        AS bucket_start,
    count(*)            AS observations,
    sum(sample_count)   AS samples
FROM obs
GROUP BY address, bucket_start
ORDER BY address, bucket_start;
