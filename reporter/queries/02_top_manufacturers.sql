-- Manufacturer breakdown. "(none)" rows are advertisers that didn't include
-- a Bluetooth SIG manufacturer ID (often beacons, BLE peripherals using only
-- service data).

SELECT
    coalesce(mfr_name, '(none)') AS manufacturer,
    count(*)                     AS observations,
    count(DISTINCT address)      AS unique_addrs,
    sum(sample_count)            AS total_samples
FROM obs
GROUP BY manufacturer
ORDER BY observations DESC
LIMIT 25;
