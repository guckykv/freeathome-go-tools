# Changelog

## Fixed

* **fahvswitch forwarded nothing** unless `-n` was given — including for the invocation its
  own usage text suggests. The device list was only filled in the `-n` branch.
* **fahinflux** opened and closed an InfluxDB client for every batch, throwing away the
  batching it was configured for, and ended with `os.Exit`, which skipped the final flush.
* Nil dereferences on optional library fields in `fahvswitch` and the InfluxDB tags.
* The kong struct tags used a syntax `reflect.StructTag` cannot parse, which made `go vet`
  abort for the whole repository and hid real findings.
* `fahcli` ended the process on a missing config file instead of reporting it.
* Both daemons ignored the error the library reports when the SysAP cannot be reached.

## Added

* **`sysapprobe`** — measures how the SysAP treats websocket clients: which keepalive it
  tolerates, how many connections it serves, how long a handshake takes. Worth running
  after a firmware update. Modes that disturb other clients need `-disruptive`.
* **`fahinflux`: InfluxDB 2.x support.** `InfluxOrg` / `INFLUX_ORG` was missing entirely,
  so 2.x rejected every write with `bucket not found`. The target is logged at startup.
* Signal handling in the daemons: `SIGTERM`/`SIGINT` shut down cleanly, `SIGHUP` forces a
  full flush. See *Running as a service* in the README.
* `make check`, `make check-noworkspace`, `make smoke`, `make all-pi64`, and CI.

## Changed

* Moved to the library's `Client` API — one `fahapi.Client` per tool instead of package
  globals.
* `fahvswitch` sends its proxy PUTs from a worker, so a slow SysAP cannot stall the
  websocket loop.

# Upgrading

**Config** — for InfluxDB 2.x, `fahinflux` needs the organisation:

```json
"InfluxDB":  "your_bucket",
"InfluxOrg": "your_org",
"InfluxToken": "your_token"
```

`InfluxDB` is the bucket name on 2.x and the database name on 1.8.x. Leave `InfluxOrg`
empty for 1.8.x. Unknown keys are ignored, so old entries may stay. The startup line tells
you what was actually read:

```
influx: writing to http://host:8086, bucket "your_bucket", org "your_org"
```

**systemd** — a daemon reconnects instead of exiting, so `Restart=on-failure` rarely
fires; prefer `Restart=always` with `KillSignal=SIGTERM`.

**Stop the old version.** Its keepalive sends text frames, which disconnect every
websocket client of the SysAP — the new daemons and the free@home app included.

**Building** — the library is a workspace neighbour during development. `go.work` in the
parent directory joins both repositories; `make check-noworkspace` verifies the build
without it.
