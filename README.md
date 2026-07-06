# EventGuard

EventGuard 是一个基于 Go 的模块化日志事件分析与自动响应系统。

它的核心模型是：

```text
Log -> Event -> Rule -> Action
```

当前 MVP 聚焦 Sing-box 日志，支持识别 REALITY 无效握手事件，按 IP 计数，并在达到阈值时输出建议封禁结果。当前阶段不会修改防火墙。

## 当前功能

- 支持从 `stdin` 读取日志，便于测试和管道输入。
- 支持从 `journalctl` 实时读取日志。
- 支持可选的 `journalctl` 历史扫描。
- 支持基于时间戳 checkpoint 的断点续传。
- 支持 JSON 配置文件。
- 支持命令行参数覆盖配置文件。
- 支持内存 Storage 和 SQLite Storage。
- 支持持久化 Event Database。
- 支持持久化 IP State Database。
- 支持 Sing-box REALITY 无效连接/无效握手日志解析。
- 支持按 IP 统计事件次数。
- 支持达到阈值后输出一次建议封禁，避免同一 IP 重复刷屏。

## 项目结构

```text
cmd/eventguard/              CLI 入口
internal/action/             Action 实现
internal/config/             配置加载
internal/core/               Event 和 Trigger 核心模型
internal/engine/             事件处理引擎和核心接口
internal/parser/singbox/     Sing-box parser
internal/reader/             stdin / journalctl reader
internal/rule/               规则实现
internal/storage/            内存状态存储
internal/storage/sqlite.go   SQLite 持久化存储
docs/                        开发文档
```

核心接口包括：

- `Reader`
- `Parser`
- `Rule`
- `Action`
- `Storage`

## 快速开始

运行全部测试：

```powershell
go test ./...
```

从 stdin 测试：

```powershell
1..5 | ForEach-Object { '2026-07-07T00:00:00+08:00 host sing-box[123]: WARN reality: invalid handshake from 45.227.254.152:443' } | go run ./cmd/eventguard -source stdin -threshold 5
```

预期输出：

```text
Need Block:

45.227.254.152

Reason:

REALITY Invalid Handshake

Count:

5
```

测试防重复输出：

```powershell
1..7 | ForEach-Object { '2026-07-07T00:00:00+08:00 host sing-box[123]: WARN reality: invalid handshake from 45.227.254.152:443' } | go run ./cmd/eventguard -source stdin -threshold 5
```

即使输入 7 条，仍然只会输出一次 `Need Block`。

## 命令行参数

```text
-config string
    optional JSON config file

-source string
    log source: stdin or journalctl

-unit string
    systemd unit used when source=journalctl

-threshold int
    event count threshold per IP
```

命令行参数会覆盖配置文件中的对应值。

示例：

```powershell
go run ./cmd/eventguard -config config.example.json -threshold 10
```

## 配置文件

示例配置见 [config.example.json](config.example.json)。

```json
{
  "source": {
    "type": "stdin",
    "unit": "sing-box",
    "history": {
      "enabled": false,
      "since": "",
      "follow": true,
      "resume": false,
      "checkpoint_file": ".eventguard/journalctl.checkpoint"
    }
  },
  "storage": {
    "type": "memory",
    "path": ".eventguard/eventguard.db"
  },
  "rules": {
    "reality_invalid_handshake": {
      "enabled": true,
      "threshold": 5
    }
  }
}
```

配置说明：

- `source.type`: 日志来源，支持 `stdin` 或 `journalctl`。
- `source.unit`: systemd unit 名称，例如 `sing-box`。
- `source.history.enabled`: 是否启用 journalctl 历史扫描。
- `source.history.since`: 历史扫描起点，例如 `2026-07-06 00:00:00` 或 `1 day ago`。
- `source.history.follow`: 历史扫描结束后是否继续实时跟随新日志。
- `source.history.resume`: 是否启用基于 checkpoint 的断点续传。
- `source.history.checkpoint_file`: checkpoint 文件路径。
- `storage.type`: 存储类型，支持 `memory` 或 `sqlite`。
- `storage.path`: SQLite 数据库文件路径。`memory` 模式下该字段不会使用。
- `rules.reality_invalid_handshake.enabled`: 是否启用 Sing-box REALITY 无效握手规则。
- `rules.reality_invalid_handshake.threshold`: 同一 IP 触发建议封禁所需的事件次数。

## 数据库

EventGuard 的 Storage 分为两类数据：

- Event Database: 保存不可变历史事件，用于审计、统计、趋势分析和规则调试。
- IP State Database: 保存每个 IP 的当前状态，例如计数、最后出现时间和状态。

当前提供两种 Storage：

- `memory`: 默认模式。计数和状态只存在于进程内，重启后丢失。
- `sqlite`: 持久化模式。事件、IP 状态、计数器和已触发标记会写入 SQLite 文件。

启用 SQLite：

```json
{
  "storage": {
    "type": "sqlite",
    "path": ".eventguard/eventguard.db"
  }
}
```

运行：

```powershell
go run ./cmd/eventguard -config config.example.json
```

SQLite 会自动创建数据库文件和表结构。

当前 SQLite 表：

- `events`: 不可变事件历史。
- `ip_states`: 每个 IP 的当前状态。
- `counters`: 规则计数器。
- `marks`: 幂等标记，例如某个 IP 是否已经触发过建议。

查询示例：

```bash
sqlite3 .eventguard/eventguard.db "select timestamp, ip, event_type from events order by id desc limit 10;"
```

```bash
sqlite3 .eventguard/eventguard.db "select ip, count, last_seen, status from ip_states order by count desc limit 10;"
```

## stdin 模式

stdin 模式适合本地测试、管道输入和离线日志验证。

```powershell
go run ./cmd/eventguard -source stdin -threshold 5
```

也可以读取本地日志文件：

```powershell
Get-Content log.txt | go run ./cmd/eventguard -source stdin -threshold 5
```

## journalctl 实时模式

在 Linux 服务器上实时跟随 Sing-box 日志：

```bash
go run ./cmd/eventguard -source journalctl -unit sing-box -threshold 5
```

当前实时模式等价于：

```bash
journalctl -o short-iso -u sing-box -f
```

## journalctl 历史扫描

历史扫描是选择性启用的。配置示例：

```json
{
  "source": {
    "type": "journalctl",
    "unit": "sing-box",
    "history": {
      "enabled": true,
      "since": "2026-07-06 00:00:00",
      "follow": false,
      "resume": false,
      "checkpoint_file": ".eventguard/journalctl.checkpoint"
    }
  },
  "rules": {
    "reality_invalid_handshake": {
      "enabled": true,
      "threshold": 5
    }
  }
}
```

当 `history.enabled=true` 且 `history.follow=false` 时，EventGuard 会扫描历史日志，扫完后退出。

等价 journalctl 命令类似：

```bash
journalctl -o short-iso --no-pager -u sing-box --since "2026-07-06 00:00:00"
```

## 历史扫描后继续实时跟随

配置：

```json
{
  "source": {
    "type": "journalctl",
    "unit": "sing-box",
    "history": {
      "enabled": true,
      "since": "2026-07-06 00:00:00",
      "follow": true,
      "resume": false,
      "checkpoint_file": ".eventguard/journalctl.checkpoint"
    }
  }
}
```

等价 journalctl 命令类似：

```bash
journalctl -o short-iso --no-pager -u sing-box --since "2026-07-06 00:00:00" -f --no-tail
```

## 断点续传

断点续传通过时间戳 checkpoint 实现。

配置：

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

行为：

- EventGuard 每处理一行 journalctl 日志，会尝试提取该行开头的 `short-iso` 时间戳。
- 提取成功后写入 `checkpoint_file`。
- 下次启动时，如果 `resume=true` 且 checkpoint 文件存在，会优先使用 checkpoint 中的时间戳作为 `journalctl --since`。

断点续传是 at-least-once 语义：

- 不会故意跳过 checkpoint 时间点的日志。
- 重启后可能重新读取同一秒内的少量日志。
- 当前规则状态仍是内存态，所以进程重启后同一 IP 可能再次输出建议。

更详细的开发说明见 [docs/journalctl-history-resume.md](docs/journalctl-history-resume.md)。

## Sing-box 事件识别

当前 parser 会识别包含 `sing-box` 或 `singbox` 的日志行，并将以下模式归类为：

```text
singbox.reality_invalid_handshake
```

支持的日志特征包括：

- 同时包含 `reality`、`invalid`、`handshake`
- 或包含 `TLS handshake: REALITY: processed invalid connection`

示例：

```text
2026-07-06T21:47:27+08:00 tw sing-box[2631710]: ERROR inbound/vless[vless-in]: process connection from 45.194.67.28:51078: TLS handshake: REALITY: processed invalid connection
```

## 输出格式

同一 IP 达到阈值后输出：

```text
Need Block:

45.194.67.28

Reason:

REALITY Invalid Handshake

Count:

5
```

当前 Action 只是打印建议，不会执行封禁。

## 测试

运行所有测试：

```powershell
go test ./...
```

测试内容覆盖：

- 配置默认值
- 配置文件加载
- SQLite schema 初始化
- Event 持久化
- IP State 持久化
- journalctl 参数构造
- checkpoint 写入
- Sing-box parser
- IP 阈值规则
- 防重复触发

## 当前限制

- 只支持 IPv4 提取。
- 只实现 Sing-box REALITY 无效握手/无效连接事件。
- `memory` Storage 下，进程重启后计数和已建议状态会丢失。
- checkpoint 只记录 journalctl 时间戳，不保存完整处理状态。
- 当前不会自动修改 nftables、ipset 或防火墙。

## 后续方向

- 增加更多 Storage 查询接口。
- 增加白名单和黑名单。
- 增加 nftables Action。
- 增加自动解封。
- 增加 SSH、Nginx、Docker 等更多 Parser。
- 增加统计和 JSON 导出。
