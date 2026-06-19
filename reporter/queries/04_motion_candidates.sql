-- Devices whose RSSI varied by >= 5 dB during a window. Stationary devices
-- show near-constant RSSI; a moving device (someone walking past, a car
-- driving by) will sweep through several dB of signal strength. Larger
-- range = more likely a transient passer-by.

SELECT
    address,
    coalesce(mfr_name, '(none)') AS manufacturer,
    sample_count,
    rssi_min,
    rssi_max,
    rssi_range,
    rssi_mean,
    first_seen,
    last_seen
FROM obs
WHERE rssi_range >= 5
ORDER BY rssi_range DESC, sample_count DESC
LIMIT 50;
