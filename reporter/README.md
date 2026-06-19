# Reporter

Read-only DuckDB queries over the scanner's JSONL captures. The scanner is the
source of truth — the reporter does not store anything of its own; rerun any
query as the capture file grows.

## Setup

Install DuckDB (single binary, no service):

```sh
curl -L https://github.com/duckdb/duckdb/releases/latest/download/duckdb_cli-linux-amd64.zip -o /tmp/duckdb.zip
cd /tmp && unzip duckdb.zip && sudo mv duckdb /usr/local/bin/
```

(Or `pip install duckdb` for the Python API, or `brew install duckdb` on macOS.)

## Running queries

```sh
# everything
./reporter/run.sh

# just the summary
./reporter/run.sh 01

# specific queries
./reporter/run.sh 04 06 07
```

Filters are substring matches on the file name in `queries/`.

## What each query tells you

| Query | Tells you                                                                    |
| ----- | ---------------------------------------------------------------------------- |
| 01    | Capture span, total observations, unique addresses, public/random mix        |
| 02    | Top manufacturers by observation and unique-MAC count                        |
| 03    | Chattiest broadcasters (highest sample_count per window) — typically beacons |
| 04    | Devices whose RSSI varied by ≥5 dB (motion candidates, walk-bys, drive-bys)  |
| 05    | Activity histogram by hour of day                                            |
| 06    | Apple Continuity subtype breakdown — confirms what's being parsed            |
| 07    | Undocumented Apple subtypes (`unknown_0xNN`) for further research            |
| 08    | Address-type usage per manufacturer — who still uses public MACs             |
| 09    | Persistence histogram: 1-window MACs (transients) vs many-window (residents) |
| 10    | **Entity groups** — fingerprints linking multiple MACs (rotation defeated)     |
| 11    | Per-address fingerprint inventory                                            |
| 12    | Fingerprint type summary                                                     |
| 20    | Apple activity-state leak (audio playing / on call / driving / etc.)         |
| 21    | Apple nearby status-flag leak (AirPods connected, WiFi on, Watch locked)     |
| 22    | 5-minute presence grid per address                                           |
| 23    | Resident / regular / recurring / transient classifier                        |
| 24    | Event bursts: hey_siri (wake word), handoff, airdrop sharing                 |
| 25    | First/last seen per address — arrival/departure log                          |
| 26    | AirPods battery + lid state (model, charging, in-case vs in-ear)             |
| 27    | HomeKit accessory inventory + global state counter (usage frequency)         |
| 28    | "Who is here right now" — last 10 minutes                                    |

## Editing queries

Each `.sql` file is independent. Edit one, rerun. The first commented line is
shown as the heading by `run.sh`, so keep that line as a one-sentence summary.

Two view files are always loaded:

- **`views.sql`** defines `obs` (flat table, one row per observation window
  with typed columns) and `obs_json` (same rows but with details kept as
  `JSON` for dynamic key inspection via `json_each` / `json_keys`).
- **`views_linkage.sql`** defines `fingerprints` — every stable identifier
  found in the parsed payloads, plus per-type views `fp_homekit`,
  `fp_ibeacon`, `fp_airdrop`, `fp_icloud`, `fp_findmy`. Schema:
  `(address, first_seen, fp_type, fp_value, lifetime)` where lifetime is
  `permanent` or `daily`.

## Linkage notes

The linkage views surface fingerprints from these payload fields:

| Source                              | Lifetime  | What it identifies                |
| ----------------------------------- | --------- | --------------------------------- |
| `homekit.device_id`                 | permanent | A specific HomeKit accessory      |
| `ibeacon.uuid`                      | permanent | A beacon product/brand            |
| `airdrop.{appleid,phone,email}_hash`| permanent | An Apple ID (8 bytes of SHA prefix)|
| `tethering_target.icloud_id`        | daily     | An iCloud account (rotates daily) |
| `findmy.public_key` (separated)     | daily     | A lost Find My device             |

Query 10 finds fingerprints broadcast by multiple BLE addresses — the
"MAC rotation defeated" signal. Empty result = no observed devices leak
enough cross-MAC identifier to link yet.

## Data shape reference

Each JSONL line:

```json
{
  "@timestamp": "2026-06-17T11:29:25Z",
  "Common": {
    "address": "79:51:ac:99:de:65",
    "address_type": "random",
    "first_seen": "2026-06-17T11:29:25Z",
    "last_seen": "2026-06-17T11:29:28Z",
    "advertisement": "..."
  },
  "observation": {
    "count": 3,
    "rssi_min": -84,
    "rssi_max": -78,
    "rssi_mean": -81.0,
    "rssi_samples": [-78, -82, -84]
  },
  "manufacturerdata": {
    "id": 76,
    "name": "Apple, Inc.",
    "details": { "apple": { "nearby": { ... } } }
  },
  "servicedata": { "details": { ... } }
}
```
