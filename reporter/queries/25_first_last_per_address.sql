-- Arrival/departure timeline. For each unique BLE address, when did we
-- first see it and when last? Sort by first_seen to read it as an arrivals
-- log; sort by last_seen DESC for "who is here right now".

SELECT
    address,
    coalesce(mfr_name, '(none)') AS manufacturer,
    address_type,
    coalesce(local_name, '')     AS local_name,
    min(first_seen)              AS first_seen,
    max(last_seen)               AS last_seen,
    age(max(last_seen), min(first_seen)) AS observed_span,
    count(*)                     AS windows,
    sum(sample_count)            AS samples,
    round(avg(rssi_mean), 1)     AS avg_rssi
FROM obs
GROUP BY address, mfr_name, address_type, local_name
ORDER BY first_seen;
