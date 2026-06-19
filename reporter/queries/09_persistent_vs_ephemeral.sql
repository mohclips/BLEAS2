-- For each BLE address, how many distinct windows did we observe it across?
--   - "ephemeral" (1 window): a passer-by, or a freshly-rotated MAC of a
--     resident device. Common with phones in the wild.
--   - "persistent" (many windows): a stationary device that doesn't rotate
--     its address — beacons, HomeKit accessories, smart speakers,
--     unrandomized peripherals.
--
-- The persistence histogram is a quick visual: tall left bar (lots of
-- one-window MACs) vs the long-tail of always-on devices.

SELECT
    windows_seen,
    count(*) AS addresses
FROM (
    SELECT
        address,
        count(DISTINCT first_seen) AS windows_seen
    FROM obs
    GROUP BY address
)
GROUP BY windows_seen
ORDER BY windows_seen;
