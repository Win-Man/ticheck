# ticheck 使用说明文档

## 概述

ticheck 是一个用于检查 TiDB 集群健康状况的命令行工具，支持对 CDC、数据库、TiDB 自适应同步模式（dr-auto-sync，同城两中心场景）等进行快捷查看与参数校验。

## 功能特性

- **数据库检查**: 验证数据库连接、性能和配置
- **CDC 检查**: 验证 TiDB Change Data Capture 组件的状态  
- **dr-auto-sync 自适应同步模式查看**: 一键可视化主中心 / DR 中心的拓扑与 dr-auto-sync 运行时状态，对应 [官方"同城两中心自适应同步模式部署"](https://docs.pingcap.com/zh/tidb/stable/two-data-centers-in-one-city-deployment/) 场景
- **参数检查**: 验证所有命令行参数的有效性

## 环境需求

- Go 1.19 或更高版本
- 可访问 TiDB 集群中的相关组件
- 集群相关的网络权限

## 下载与编译

### 从源码编译

```bash
# 克隆代码库
git clone https://github.com/Win-Man/ticheck.git
cd ticheck

# 编译（包含竞态检测）
make build

# 或者针对特定平台编译
make arm64  # ARM64 架构
make amd64  # AMD64 架构
```

### 快速获取

```bash
# 使用 Go 直接获取
go install github.com/Win-Man/ticheck@latest
```

## 命令详解

### 参数检查 (args-check)

将当前集群中 TiDB / PD / TiKV 三类组件的配置项与 TiDB variables，与 TOML
`[check-template]` 中给出的期望值做 PASS / NOPASS 比对。是上线前做合规基线的
推荐入口。

#### 用法
```bash
./ticheck args-check [全局 flag]
```

#### 全局 flag（本子命令唯一支持）
- `--config`, `-C`：TOML 配置文件路径（必填）
- `--log-level`, `-L`：日志级别（`info` / `debug` / `warn` / `error` / `fatal`）
- `--log-path`：日志文件输出路径（覆盖 TOML 中 `[log] log-path`）
- `--output`：取值 `print`（默认，输出到标准输出）或 `file`（按 `result-file-path` 追加到结果文件）

业务参数（数据库连接 + 期望值清单）全部在 TOML 中：

| TOML 键 | 说明 |
| --- | --- |
| `[db-config].host / port / user / password / database` | TiDB 连接信息，命令会通过它连接集群以拉 `INFORMATION_SCHEMA.CLUSTER_INFO` / `SHOW VARIABLES` 等 |
| `[check-template].tidb-config = [{name, value}]` | TiDB 期望配置项列表，通过 `SHOW CONFIG WHERE TYPE='tidb'` 比对 |
| `[check-template].pd-config = [{name, value}]` | PD 期望配置项列表 |
| `[check-template].tikv-config = [{name, value}]` | TiKV 期望配置项列表 |
| `[check-template].tidb-variables = [{name, value}]` | TiDB variables 期望值，会逐一连接集群中每个 TiDB 实例执行 `SHOW VARIABLES LIKE` |
| `result-file-path` | `--output=file` 时的输出文件路径 |

#### 示例
```bash
# 使用自带示例：连接 test 库，比对若干 tidb/pd/tikv 配置
./ticheck args-check --config=./config/argscheck.toml.example

# 同时把 PASS / NOPASS 表追加写到结果文件
./ticheck args-check --config=./config/argscheck.toml.example --output=file
```

### 数据库检查 (db-check)

连接 TiDB（依赖 `[db-config]`），按 `[db-check-items].user-defined-sqls`
中的每条 SQL 执行并把结果以表格渲染到标准输出。常用于回答"集群当前是什么样"
这类临时调查问题。

#### 用法
```bash
./ticheck db-check [全局 flag]
```

#### 全局 flag（本子命令唯一支持）
- `--config`, `-C`：TOML 配置文件路径（必填）
- `--log-level`, `-L`：日志级别
- `--log-path`：日志文件输出路径

业务参数全部在 TOML 中：

| TOML 键 | 说明 |
| --- | --- |
| `[db-config].host / port / user / password / database` | TiDB 连接信息 |
| `[db-check-items].user-defined-sqls = [{name, sql}]` | 要执行的用户自定义 SQL 列表，name 会作为表格标题 |

#### 示例
```bash
./ticheck db-check --config=./config/dbcheck.toml.example
```

### CDC 检查 (cdc-check)

调用 TiCDC Open API（`GET /api/v2/captures`、`GET /api/v2/changefeeds`）
查询 capture 列表与 changefeed 运行状态（含 checkpoint 时间、运行错误等）。

#### 用法
```bash
./ticheck cdc-check [全局 flag]
```

#### 全局 flag（本子命令唯一支持）
- `--config`, `-C`：TOML 配置文件路径（必填）
- `--log-level`, `-L`：日志级别
- `--log-path`：日志文件输出路径

业务参数全部在 TOML 中：

| TOML 键 | 说明 |
| --- | --- |
| `[cdc-config].cdc-address` | CDC OpenAPI 地址列表（**字符串数组**，如 `["127.0.0.1:8300", "10.0.0.2:8300"]`，可轮询多个 capture 节点） |

#### 示例
```bash
./ticheck cdc-check --config=./config/cdccheck.toml.example
```

### dr-auto-sync 自适应同步模式查看 (dr-check)

一键可视化 **同城两中心自适应同步模式（dr-auto-sync）** 下的主中心 / DR 中心拓扑与当前状态，用于快速核对副本布局是否符合 `replication-mode.dr-auto-sync` 的设计意图。
该子命令**仅通过调用 PD HTTP API** 拉取信息（不连 TiDB、不读 BR 备份目录），对应官方文档 [同城两中心自适应同步模式部署](https://docs.pingcap.com/zh/tidb/stable/two-data-centers-in-one-city-deployment/)。

输出包含：

- `replication-mode`：官方取值 `majority`（多数派、异步）或 `dr-auto-sync`（自适应同步模式）
- `label-key` 与 `primary`：用于标识**主中心（primary data center）** 的标签键与取值；与之相对的另一个取值即 DR 中心
- 两个超时参数（来自 `replication-mode.dr-auto-sync`）：
  - `wait-store-timeout`：主 / DR 中心 store 下线超过该时间后，副本模式降级为 `async`
  - `wait-sync-timeout`：从 `async` 恢复时，等待各 Region 重新变 `synchronous` 的超时
- 每个 TiKV store 的 location labels、按 [placement rules](https://docs.pingcap.com/zh/tidb/stable/configure-placement-rules/) 推导出的 Role（`voter` / `learner` / `voter-follower` 等）以及 Region 数、Leader 数
- dr-auto-sync **运行时状态**：官方仅有 `sync`（双中心同步）、`sync-recover`（恢复中）、`async`（已降级异步）三种
- PD 成员列表（含 priority 并标记 `Leader`）

#### 用法
```bash
./ticheck dr-check [global flags]
```

`dr-check` 没有自己的专用 flag，仅使用以下全局 flag（与其它子命令一致）：

| Flag / TOML 键 | 说明 |
| --- | --- |
| `--config`, `-C` | 加载 TOML 配置（必须设置，指定 PD 地址） |
| `--log-level`, `-L` | 日志级别（`info` / `debug` / `warn` / `error` / `fatal`） |
| `--log-path` | 日志文件输出路径 |

#### 配置文件（`config/drcheck_config.toml`）
```toml
[log]
log-level = "Debug"
log-path = "./dev/ticheck.log"

[dr-config]
pd-address = "127.0.0.1:2379"   # 任一可用的 PD 地址（HTTP，端口默认 2379）
```

> `dr-check` 只需要 PD 的 HTTP 端口即可工作，不需要 TiDB 用户名/密码、TiKV、备份目录等。

#### 示例
```bash
# 一键查看主中心 / DR 中心拓扑与 dr-auto-sync 状态
./ticheck dr-check --config=./config/drcheck_config.toml

# 同时调整日志级别
./ticheck dr-check --config=./config/drcheck_config.toml --log-level=info
```

## 配置文件使用

对于复杂的检查任务，可以创建配置文件。系统提供了多个标准配置模板：

- `./config/argscheck_config.example.toml`
- `./config/dbcheck_config.example.toml`
- `./config/cdccheck_config.example.toml`
- `./config/drcheck_config.example.toml`

### 配置文件示例

数据库检查的 TOML 配置示例 (`config/dbcheck_config.example.toml`)：
```toml
[tidb]
host = "localhost"
port = 4000
user = "root"
password = ""
ssl-ca = ""
ssl-cert = ""
ssl-key = ""

[pd]
pd_address = "localhost:2379"

[check]
cluster_name = "my_cluster"
skip_db_tls = false
skip_versatile_test = true
log_level = "info"
show_slowest = true
concurrency = 10
```

## 高级用法示例

### 完整健康检查流程

以下是一个完整的集群健康检查流程，涵盖从参数检查到具体组件验证：

```bash
#!/bin/bash
# 完整的集群健康检查脚本示例

# 1. 参数有效性验证：将 TiDB / PD / TiKV 配置与期望值做比对
echo "Step 1: Validating arguments..."
./ticheck args-check \
    --config=./config/argscheck.toml.example \
    || { echo "Argument validation failed"; exit 1; }

# 2. 数据库健康检查：连 TiDB 跑用户自定义 SQL
echo "Step 2: Checking database health..."
./ticheck db-check \
    --config=./config/dbcheck.toml.example \
    || { echo "Database check failed"; exit 1; }

# 3. CDC 连接验证（可选，需要 TiCDC 集群）
echo "Step 3: Checking CDC connections (if applicable)..."
./ticheck cdc-check \
    --config=./config/cdccheck.toml.example \
    || { echo "CDC check failed"; exit 1; }

# 4. dr-auto-sync 自适应同步模式下的主中心 / DR 中心拓扑查看
echo "Step 4: Viewing dr-auto-sync topology..."
./ticheck dr-check \
    --config=./config/drcheck.toml.example \
    || { echo "dr-auto-sync check failed"; exit 1; }

echo "All health checks passed successfully!"
```

### 性能监控集成

可以在 CI/CD 流程中使用 ticheck 进行定期的集群健康监控：

```bash
# 在定时任务中运行健康检查并将结果保存
./ticheck db-check \
    --config=/etc/ticheck/dbcheck.toml \
    --log-level=info > /var/log/ticheck-$(date +%Y%m%d-%H%M%S).log 2>&1
```

> 数据库用户名/密码等敏感信息**写在 TOML 中，且 chmod 600 收口权限**，不建议
> 出现在命令行历史中。

## 故障排除

### 常见错误与解决方案

1. **连接超时错误**:
   - 检查网络连通性，使用 `telnet` 验证端口可达性
   - 确认防火墙没有阻止相应的端口访问

2. **认证失败**:
   - 确认提供的用户名和密码正确无误
   - 检查账户权限是否足以执行所需操作

3. **配置文件读取失败**:
   - 检查配置文件是否存在且格式正确
   - 确认使用的配置文件路径正确

4. **权限不足**:
   - 验证数据库用户是否具有足够的权限进行检查
   - 确认操作系统对某些文件或目录有正确的读写权限

## 最佳实践

1. **定期健康检查**:
   - 建议设置定时任务每周/每日执行基础检查
   - 关键业务上线前必须执行完整的健康检查

2. **配置文件管理**:
   - 不同环境（开发、测试、生产）使用独立的配置文件
   - ticheck 的连接信息（TiDB 用户 / 密码、PD 地址、CDC 地址）**必须**写在
     TOML 中；请把配置文件权限收紧到 `chmod 600`，并避免纳入版本库

3. **日志保存**:
   - 定期保存检查日志，便于问题追溯
   - 配置日志轮转避免磁盘空间耗尽

## 命令行选项说明

所有子命令共用以下全局 flag（Cobra 自动提供 `--help` / `-h`，根命令附加 `--version` / `-V`）：

| Flag | 说明 |
| --- | --- |
| `--config`, `-C` | TOML 配置文件路径（必填） |
| `--log-level`, `-L` | 日志级别：`debug` / `info` / `warn` / `error` / `fatal` |
| `--log-path` | 日志文件输出路径，覆盖 TOML 中 `[log] log-path` |
| `--output` | 仅 `args-check` 支持，取值 `print`（默认）或 `file` |

> 之前的旧文档把 `db-check / cdc-check / args-check` 等子命令描述成支持
> `--host` / `--tidb_host` / `--mysql_user` / `--upstream-port` / `--backup_dir`
> 等专用 flag —— 这些 flag 实际上**并不存在**。当前实现下，连接信息等业务
> 参数一律写在 TOML 中通过 `--config` 传入。

业务相关的字段（如 `[db-config].host`、`[cdc-config].cdc-address` 等）请直接
参考 `config/*.toml.example`。