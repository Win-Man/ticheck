# ticheck 使用说明文档

## 概述

ticheck 是一个用于检查 TiDB 集群健康状况的命令行工具，支持对 CDC、数据库、DR（灾难恢复）和参数等方面进行全面检查。

## 功能特性

- **数据库检查**: 验证数据库连接、性能和配置
- **CDC 检查**: 验证 TiDB Change Data Capture 组件的状态  
- **DR 检查**: 检查灾难恢复配置的正确性
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

检查所有输入参数的有效性，这是使用其他功能前的建议步骤。

#### 用法
```bash
./ticheck args-check [flags]
```

#### 支持的参数
- `--tidb_host`, `-H`: TiDB 服务主机地址
- `--tidb_port`, `-P`: TiDB 服务端口
- `--mysql_user`, `-u`: MySQL 用户名
- `--mysql_password`, `-p`: MySQL 密码
- `--pd_host`: PD (Placement Driver) 主机地址
- `--pd_port`: PD 服务端口

#### 示例
```bash
./ticheck args-check --tidb_host=localhost --tidb_port=4000 --mysql_user=root --mysql_password=
```

### 数据库检查 (db-check)

验证数据库连接、性能、权限以及各项配置是否正常。

#### 用法
```bash
./ticheck db-check [flags]
```

#### 支持的主要参数
- `--tidb_host`, `-H`: TiDB 服务主机地址
- `--tidb_port`, `-P`: TiDB 服务端口
- `--mysql_user`, `-u`: MySQL 用户名
- `--mysql_password`, `-p`: MySQL 密码
- `--cluster-name`: 集群名称标识符
- `--pd_host`: PD 服务主机地址
- `--pd_port`: PD 服务端口
- `--skip_db_tls`: 跳过数据库 TLS 检查
- `--skip_versatile_test`: 跳过多功能测试
- `--log_level`: 设置日志级别（info, warn, error, debug）
- `--conf`: 配置文件路径
- `--show-slowest`: 显示最慢的查询结果
- `--concurrency`: 并发度设置（默认5）

#### 示例
```bash
# 基础连接检查
./ticheck db-check --tidb_host=localhost --tidb_port=4000 --mysql_user=root --mysql_password=

# 带集群名称的详细检查
./ticheck db-check --tidb_host=192.168.1.100 --tidb_port=4000 --mysql_user=tidb_admin --mysql_password=password --cluster-name=my_tidb_cluster

# 使用配置文件执行检查
./ticheck db-check --conf=./config/dbcheck_config.toml
```

### CDC 检查 (cdc-check)

验证 TiDB CDC (Change Data Capture) 功能和下游系统状态。

#### 用法
```bash
./ticheck cdc-check [flags]
```

#### 支持的主要参数
- `--tiup`: 使用 TiUP 配置检查
- `--tiup_user`: TiUP SSH 用户名
- `--ssh_keypath`: SSH 密钥路径
- `--conf`: 配置文件路径
- `--cluster-name`: 集群名称标识符
- `--enable-progress`: 启用进度条显示
- `--cdc_upstream_host`: CDC 上游 TiDB 主机 IP
- `--cdc_upstream_port`: CDC 上游 TiDB 端口
- `--cdc_upstream_user`: CDC 上游用户名
- `--cdc_upstream_password`: CDC 上游密码
- `--cdc_downstream_host`: CDC 下游主机 IP
- `--cdc_downstream_port`: CDC 下游端口
- `--cdc_downstream_user`: CDC 下游用户名
- `--cdc_downstream_password`: CDC 下游密码
- `--source_db_name`: 源数据库名称
- `--check_source_table_name`: 检查的源表名称
- `--target_table_names`: 目标表名称（逗号分隔）

#### 示例
```bash
# 基础 CDC 功能检查
./ticheck cdc-check --cdc_upstream_host=192.168.1.100 --cdc_upstream_port=4000 --cdc_upstream_user=root --cdc_upstream_password="" --cdc_downstream_host=192.168.1.101 --cdc_downstream_port=4000 --source_db_name=test_db --check_source_table_name=test_table --target_table_names=target_table

# 使用配置文件的 CDC 检查
./ticheck cdc-check --conf=./config/cdccheck_config.toml
```

### DR (灾难恢复) 检查 (dr-check)

验证 TiDB 灾难恢复相关的配置、备份和恢复机制。

#### 用法
```bash
./ticheck dr-check [flags]
```

#### 支持的主要参数
- `--tidb_host`, `-H`: TiDB 服务主机
- `--tidb_port`, `-P`: TiDB 服务端口
- `--mysql_user`, `-u`: MySQL 用户名
- `--mysql_password`, `-p`: MySQL 密码
- `--pd_host`: PD 服务主机
- `--pd_port`: PD 服务端口
- `--tikv_host`: TiKV 服务主机
- `--tikv_restful_port`: TiKV Restful 接口端口
- `--backup_dir`: 备份目录路径
- `--gc_value`: GC 保存天数设置值
- `--min_safe_ts_hours`: 最小安全 TS 小时数
- `--log_level`: 日志输出级别
- `--conf`: 配置文件路径
- `--dr_type`: DR 类型设置

#### 示例
```bash
# 基础 DR 配置检查
./ticheck dr-check --tidb_host=192.168.1.100 --tidb_port=4000 --pd_host=192.168.1.101 --pd_port=2379 --tikv_host=192.168.1.102 --tikv_restful_port=20180 --mysql_user=root --mysql_password=""

# 检查备份配置
./ticheck dr-check --backup_dir=/tmp/tidb_backup --gc_value=86400 --conf=./config/drcheck_config.toml
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

# 1. 参数有效性验证
echo "Step 1: Validating arguments..."
./ticheck args-check \
    --tidb_host=localhost \
    --tidb_port=4000 \
    --mysql_user=root \
    --mysql_password="" \
    || { echo "Argument validation failed"; exit 1; }

# 2. 数据库健康检查
echo "Step 2: Checking database health..."
./ticheck db-check \
    --tidb_host=localhost \
    --tidb_port=4000 \
    --mysql_user=root \
    --mysql_password="" \
    --show-slowest \
    --concurrency=10 \
    || { echo "Database check failed"; exit 1; }

# 3. CDC 连接验证
echo "Step 3: Checking CDC connections (if applicable)..."
./ticheck cdc-check \
    --cdc_upstream_host=localhost \
    --cdc_upstream_port=4000 \
    --cdc_upstream_user=root \
    --cdc_upstream_password="" \
    --source_db_name=test \
    --check_source_table_name=test_table \
    --target_table_names=cdc_test_target \
    || { echo "CDC check failed"; exit 1; }

# 4. 灾难恢复配置验证
echo "Step 4: Checking DR configurations..."
./ticheck dr-check \
    --tidb_host=localhost \
    --tidb_port=4000 \
    --pd_host=localhost \
    --pd_port=2379 \
    --mysql_user=root \
    --mysql_password="" \
    || { echo "DR check failed"; exit 1; }

echo "All health checks passed successfully!"
```

### 性能监控集成

可以在 CI/CD 流程中使用 ticheck 进行定期的集群健康监控：

```bash
# 在定时任务中运行健康检查并将结果保存
./ticheck db-check \
    --tidb_host=production-server.company.com \
    --tidb_port=4000 \
    --mysql_user=ticheck_user \
    --mysql_password="$DB_PASSWORD" \
    --log_level=info > /var/log/ticheck-$(date +%Y%m%d-%H%M%S).log 2>&1
```

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
   - 配置文件中不要包含敏感的密码信息

3. **日志保存**:
   - 定期保存检查日志，便于问题追溯
   - 配置日志轮转避免磁盘空间耗尽

## 命令行选项说明

```bash
--help, -h          显示所有可用命令和选项的帮助信息
--version, -v       显示当前程序版本
--log_file          指定日志文件输出位置
--log_level         设置日志级别: debug, info, warn, error
--config, -c        使用指定的配置文件
```

所有检查命令都遵循标准的 CLI 框架，支持简写形式参数如 `-H` 表示 `--host`。