-- Chattiest broadcasters in this capture: highest sample_count per window.
-- A device with thousands of samples in a 5-minute window is broadcasting
-- ~once per second — likely a beacon (iBeacon, Eddystone) or a tethering
-- accessory keeping itself discoverable.

SELECT
    address,
    address_type,
    coalesce(mfr_name, '(none)') AS manufacturer,
    coalesce(local_name, '')     AS local_name,
    sample_count,
    rssi_min, rssi_max, rssi_range,
    first_seen, last_seen
FROM obs
ORDER BY sample_count DESC
LIMIT 30;
