# EventGuard

EventGuard is a modular log event processing engine built around:

```text
Log -> Event -> Rule -> Action
```

The MVP supports Sing-box REALITY invalid-handshake detection and prints block suggestions without changing firewall state.

## Run

From stdin:

```powershell
go run ./cmd/eventguard -source stdin -threshold 5
```

From journalctl on Linux:

```bash
go run ./cmd/eventguard -source journalctl -unit sing-box -threshold 5
```

Example input:

```text
2026-07-07T00:00:00+08:00 host sing-box[123]: WARN reality: invalid handshake from 45.227.254.152:443
```

When the same IP reaches the threshold, EventGuard prints:

```text
Need Block:

45.227.254.152

Reason:

REALITY Invalid Handshake

Count:

5
```
