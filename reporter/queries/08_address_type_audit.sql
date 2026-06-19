-- Address-type breakdown per manufacturer. Privacy-relevant: devices that
-- still use the "public" BLE address haven't adopted MAC randomization and
-- can be tracked across long time spans. RPA (Resolvable Private Address)
-- is the modern privacy default; NRPA rotates aggressively.

SELECT
    coalesce(mfr_name, '(none)') AS manufacturer,
    address_type,
    count(*)                     AS observations,
    count(DISTINCT address)      AS unique_addrs
FROM obs
GROUP BY manufacturer, address_type
ORDER BY manufacturer, observations DESC;
