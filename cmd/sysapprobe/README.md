# sysapprobe

Measures how a System Access Point treats websocket clients. Run it after a
firmware update to check whether the assumptions `fahapi` builds on still hold.

It speaks the websocket protocol directly rather than going through the library,
because the point is to test the layer the library sits on.

## Modes

| Mode | Question it answers |
| --- | --- |
| `none` | Does the SysAP keep an idle connection open at all? |
| `ping` | Does it answer ping frames? This is what the library does. |
| `text` | What happens with the keepalive the library used to send? |
| `text-once` | Is a single text frame enough, or is it the rate? |
| `ladder` | How many simultaneous connections does it serve? |
| `collateral` | Does one client's text frame disturb the others? |

`text`, `text-once` and `collateral` send a text frame, which disconnects
**every** websocket client of the SysAP — the free@home app included. They refuse
to run without `-disruptive`.

## Measurements against software 2.6

Taken with no other client connected, which matters: any measurement is
worthless while something else is talking to the SysAP.

| Question | Result |
| --- | --- |
| Keepalive needed to stay connected? | No — 300s idle, 99 messages, no disconnect |
| Are pings answered? | Yes, 59 of 59, even at one per second |
| Is a single text frame fatal? | Yes, connection closed after 2.2s, 3 of 3 |
| Is it the rate? | No — a single frame does it |
| Connection limit? | None at 8; all held for their full duration |
| Does a text frame hit other clients? | Yes, all of them, 3 of 3 |
| Handshake duration | 0.1–0.3s idle, up to 10.3s while recovering |

The last row is why `fahapi` allows 30s for a handshake: a reconnect happens
exactly when the SysAP is recovering, which is when handshakes are slowest.

## Usage

```sh
sysapprobe -mode ping -interval 5s -d 60s
sysapprobe -mode ladder -n 8 -stagger 15s -d 240s
sysapprobe -mode collateral -disruptive
```

Reads `~/.fahapi-config.json` by default; `-c` points elsewhere. Only `Host`,
`Username` and `Password` are used. Every mode is read-only — it never writes a
datapoint.
