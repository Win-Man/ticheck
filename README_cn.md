# ticheck

**ticheck** 是一个用 Go 编写的全面的 TiDB 集群健康检查工具。它可以对 TiDB 集群的 CDC、数据库、TiDB 自适应同步模式（dr-auto-sync）等组件进行检查，以确保它们正常运行。

## 功能

- **数据库检查**: 验证数据库连接性、性能和配置
- **CDC 检查**: 监控变更数据捕获功能
- **dr-auto-sync 自适应同步模式查看**: 配合 [官方"同城两中心自适应同步模式部署"](https://docs.pingcap.com/zh/tidb/stable/two-data-centers-in-one-city-deployment/) 场景，一键可视化主中心 / DR 中心拓扑 —— 通过 PD HTTP API 拉取 replication mode、label-key 与 primary 主中心、各 TiKV store 按 location labels 派生的 Role、dr-auto-sync 运行时状态、PD 成员、`wait-store-timeout` 与 `wait-sync-timeout`
- **参数验证**: 全面的标志位和参数验证

## 先决条件

- Go 1.19+
- 访问 TiDB 集群组件的权限

## 安装

### 从源码安装

```bash
git clone https://github.com/YOUR_USERNAME/ticheck.git
cd ticheck
make build
```

### 使用 Go Install

```bash
go install github.com/YOUR_USERNAME/ticheck@latest
```

## 快速开始

所有子命令共享同一组全局 flag，每个子命令的实际配置通过匹配的 TOML 文件传入。
下方示例均指向仓库自带的 `.toml.example`，生产环境请先复制并按需修改（如
`cp config/dbcheck.toml.example dbcheck.toml`）。

```bash
# 连 TiDB，按 db-check-items 跑用户自定义 SQL
./ticheck db-check --config=./config/dbcheck.toml.example

# 通过 CDC HTTP API 查询 captures / changefeeds 状态
./ticheck cdc-check --config=./config/cdccheck.toml.example

# 一键查看 dr-auto-sync 自适应同步模式下的主中心 / DR 中心拓扑与状态
./ticheck dr-check --config=./config/drcheck.toml.example

# 将 TiDB / PD / TiKV 配置和 TiDB 变量与期望值做 PASS / NOPASS 比对
./ticheck args-check --config=./config/argscheck.toml.example
```

## 命令

### 常用标志位

所有子命令通用（`args-check` 额外提供 `--output`）：

- `--config`, `-C`：TOML 配置文件路径（执行任何子命令都需要，可参考 `config/*toml.example`）
- `--log-level`, `-L`：日志级别（`info` / `debug` / `warn` / `error` / `fatal`）
- `--log-path`：日志文件输出路径（会覆盖 TOML 中 `[log] log-path`）

### 可用命令

#### `db-check`
检查数据库连接、性能和配置。

#### `cdc-check`
检查变更数据捕获进程和复制健康状态。

#### `dr-check`
一键查看 **同城两中心自适应同步模式（dr-auto-sync）** 的拓扑与当前状态，用于快速核对主中心 / DR 中心的副本布局是否符合 `replication-mode.dr-auto-sync` 的设计意图。

它调用 PD 的 HTTP API（`/pd/api/v1/config`、`/pd/api/v1/stores`、
`/pd/api/v1/replication_mode/status`、`/pd/api/v1/members`、
`/pd/api/v1/config/rules`），输出：

- `replication-mode`：官方取值 `majority`（多数派、异步）或 `dr-auto-sync`（自适应同步模式）
- `label-key` 与 `primary`：用于标识**主中心（primary data center）** 的标签键与取值；与之相对的另一个取值即 DR 中心
- 两个超时参数（来自 `replication-mode.dr-auto-sync`）：
  - `wait-store-timeout`：在主 / DR 中心 store 下线超过该时间后，将副本模式降级为 `async`
  - `wait-sync-timeout`：从 `async` 恢复时，等待各 Region 重新变 `synchronous` 的超时
- 每个 TiKV store 的信息：location labels、根据 [placement rules](https://docs.pingcap.com/zh/tidb/stable/configure-placement-rules/) 推导出的 Role（`voter` / `learner` / `voter-follower` 等）、Region 数与 Leader 数
- dr-auto-sync **运行时状态**：官方仅有 `sync`（双中心同步）、`sync-recover`（恢复中）、`async`（已降级异步）三种
- PD 成员列表（含 priority 并标记 `Leader`）

> 详细架构说明请参阅官方文档 [同城两中心自适应同步模式部署](https://docs.pingcap.com/zh/tidb/stable/two-data-centers-in-one-city-deployment/)。

#### `args-check`
验证提供给工具的参数。

## 配置

可以从 TOML 文件加载配置。可用示例配置文件包括：

- `config/argscheck_config.example.toml`
- `config/dbcheck_config.example.toml`
- `config/drcheck_config.example.toml`
- `config/cdccheck_config.example.toml`

## 架构

- **命令结构**: 使用 Cobra CLI 框架构建
- **数据库层**: 使用 GORM ORM 和 MySQL 适配器
- **进程管理**: 自定义进度指示器
- **Linux 操作**: 分布式系统的远程执行实用程序

## 许可证

本项目基于 Apache License 2.0 许可 - 详情请参阅 [LICENSE](./LICENSE) 文件。

## 贡献

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/awesome-feature`)
3. 进行更改
4. 提交更改 (`git commit -m 'feat: 添加 awesome 功能'`)
5. 推送到分支 (`git push origin feature/awesome-feature`)
6. 开始一个 Pull Request

---
Copyright © 2023-present TiDB 社区