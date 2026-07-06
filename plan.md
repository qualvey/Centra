# EventGuard 项目目标与开发计划

## 项目简介

EventGuard 是一个基于 Go 开发的模块化日志事件分析与自动响应系统。

项目的目标不是解析某一种日志，而是构建一个通用的 **事件处理引擎（Event Processing Engine）**。

任何日志（如 Sing-box、SSH、Nginx、Docker 等）都可以经过对应的 Parser 转换为统一的 Event，由规则引擎进行分析，并执行相应的 Action。

整个项目遵循：

> **Log → Event → Rule → Action**

而不是：

> **Log → if/else → Firewall**

项目强调模块化、低耦合、高扩展性，避免随着功能增加导致代码复杂度不断上升。

---

# 设计目标

## 1. 模块职责单一（Single Responsibility）

每个模块只负责一件事情。

例如：

* Reader 负责读取日志
* Parser 负责解析日志
* Rule Engine 负责判断事件
* Action 负责执行动作
* Storage 负责状态持久化

任何模块不得承担多个职责。

---

## 2. 面向接口开发

所有核心组件必须定义接口。

例如：

```text
Reader
Parser
Rule
Action
Storage
```

任何实现都应通过接口注册，而不是硬编码。

例如以后新增：

* Sing-box Parser
* SSH Parser
* Nginx Parser

不应修改已有 Parser。

---

## 3. Event 作为统一数据模型

整个系统内部禁止直接处理日志字符串。

所有日志必须先转换为统一 Event。

例如：

```
Log
 ↓
Parser
 ↓
Event
 ↓
Rule
 ↓
Action
```

Event 应包含统一字段，例如：

* Timestamp
* Source
* Service
* EventType
* Level
* IP
* Message
* Metadata

以后新增任何 Parser，都应输出同一种 Event。

---

## 4. Rule Engine 与 Parser 解耦

Parser 不允许：

* 统计次数
* 判断攻击
* 封禁 IP
* 写数据库

Parser 只负责：

```
Log -> Event
```

Rule Engine 负责：

```
Event -> Decision
```

---

## 5. Action 与 Rule 解耦

Rule 不负责执行具体动作。

Rule 只返回：

```
Trigger
```

Action 决定：

* Block IP
* Send Notification
* Write Report
* Ignore
* Webhook

以后新增 Action 时，不需要修改 Rule。

---

## 6. Storage 独立

所有统计信息必须通过 Storage 管理。

例如：

* 最近出现时间
* 命中次数
* Reputation Score
* Block 状态

Rule 不应维护自己的状态。

---

# 第一阶段（MVP）

仅支持 Sing-box。

实现功能：

* 实时读取 journalctl
* 解析 Sing-box 日志
* 提取 Event
* 统计同一 IP 的事件数量
* 根据阈值输出建议封禁

例如：

```
Need Block:

45.227.254.152

Reason:

REALITY Invalid Handshake

Count:

5
```

第一阶段不直接修改防火墙。

仅输出建议。

---

# 第二阶段

增加自动响应。

支持：

* nftables
* ipset（可选）
* 白名单
* 黑名单
* 自动封禁
* 自动解封

Rule Engine 根据配置触发对应 Action。

---

# 第三阶段

支持更多日志来源。

例如：

* SSH
* Nginx
* Docker
* Systemd
* Fail2Ban 日志（兼容）

每种日志来源应实现独立 Parser。

不得修改已有 Parser。

---

# 第四阶段

增加 Reputation Engine。

每个 IP 建立信誉模型。

例如：

```
REALITY Invalid +20

SSH Password Failed +10

HTTP Scan +15

Port Scan +30
```

达到配置阈值后：

```
Score >= 100

↓

Block
```

信誉分应支持衰减。

例如：

24 小时后自动降低。

---

# 第五阶段

增加统计模块。

支持：

* 每日统计
* Top 攻击 IP
* Top Event
* Top 国家（可扩展）
* 时间趋势
* 导出 JSON

统计模块只读取 Event，不影响 Rule。

---

# 架构

```
              +----------------+
              |  Log Reader    |
              +----------------+
                      |
                      v
              +----------------+
              |    Parser      |
              +----------------+
                      |
                      v
              +----------------+
              |     Event      |
              +----------------+
                      |
                      v
              +----------------+
              |  Rule Engine   |
              +----------------+
               /      |       \
              /       |        \
             v        v         v
      +---------+ +---------+ +---------+
      | Action  | | Storage | | Metrics |
      +---------+ +---------+ +---------+
```

---

# 开发原则

1. 优先保证架构清晰，而不是快速实现功能。
2. 优先新增模块，而不是修改已有模块。
3. 所有组件尽量通过接口组合。
4. 避免大型 God Object。
5. 避免业务逻辑散落在多个模块。
6. 保持低耦合、高内聚。
7. 每个阶段都应保持可运行状态。
8. 先实现 MVP，再逐步扩展，不提前设计未使用的复杂功能。

---

# 最终目标

EventGuard 最终定位不是一个 Sing-box 工具，而是一个通用的日志事件分析与自动响应框架。

它能够持续接收不同来源的日志，将日志转换为统一事件，使用规则引擎分析行为，并根据策略自动执行响应动作。

通过统一的 Event 模型和模块化架构，实现良好的扩展性，使未来新增日志来源、检测规则或响应动作时，只需增加对应模块，而无需修改系统核心逻辑。
