# Journalctl History Scan And Timestamp Resume

## Goal

Journalctl input supports two modes:

1. Realtime follow only, which keeps the existing behavior.
2. Optional history scan, which can read historical journal entries before optionally following new entries.

When timestamp resume is enabled, EventGuard stores the latest processed journal timestamp in a checkpoint file. On the next start, it passes that timestamp back to `journalctl --since` so scanning resumes from the last known point.

## Configuration

History scanning is disabled by default.

```json
{
  "source": {
    "type": "journalctl",
    "unit": "sing-box",
    "history": {
      "enabled": true,
      "since": "",
      "follow": true,
      "resume": true,
      "checkpoint_file": ".eventguard/journalctl.checkpoint"
    }
  }
}
```

Fields:

- `enabled`: enables historical scan mode.
- `since`: optional starting timestamp for the first scan, such as `2026-07-06 00:00:00` or `1 day ago`.
- `follow`: when true, EventGuard continues following new entries after history catches up.
- `resume`: when true, EventGuard reads `checkpoint_file` at startup and uses it as the effective `since` value when present.
- `checkpoint_file`: file used to persist the latest processed journal timestamp.

If `resume` is true and `checkpoint_file` is empty, EventGuard uses `.eventguard/journalctl.checkpoint`.

## Journalctl Commands

Realtime-only mode:

```bash
journalctl -f -o short-iso -u sing-box
```

History scan without follow:

```bash
journalctl --no-pager -o short-iso -u sing-box
```

History scan with follow:

```bash
journalctl --no-pager -o short-iso -u sing-box -f --no-tail
```

Resume from checkpoint:

```bash
journalctl --no-pager -o short-iso -u sing-box --since "2026-07-06T21:47:27+08:00" -f --no-tail
```

## Resume Semantics

Timestamp resume is at-least-once:

- EventGuard writes the checkpoint after a line has gone through the engine.
- On restart, `journalctl --since` may include entries that share the same timestamp as the checkpoint.
- This avoids skipping events, but can re-read a small timestamp boundary.

The current rule storage is in memory, so duplicate suggestions across process restarts are still possible. Durable action state should be added before automatic firewall actions are enabled.

## Implementation Notes

- The reader owns journalctl process construction.
- The engine owns processing order and calls `CommitLine` after a line is handled.
- Checkpoint writing is atomic: write a temporary file, then rename it over the checkpoint file.
- Timestamp extraction expects journalctl `short-iso` output at the start of each line.
