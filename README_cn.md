# ticheck

**ticheck** 是一个用 Go 编写的全面的 TiDB 集群健康检查工具。它可以对 TiDB 集群的 CDC、数据库、容灾等组件进行检查，以确保它们正常运行。

## 功能

- **数据库检查**: 验证数据库连接性、性能和配置
- **CDC 检查**: 监控变更数据捕获功能
- **容灾检查**: 验证灾难恢复能力
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

```bash
# 检查数据库连接性和配置
./ticheck db-check --host=localhost --port=4000

# 验证 CDC 组件
./ticheck cdc-check --upstream-host=localhost --upstream-port=4000

# 验证灾难恢复设置
./ticheck dr-check --host=localhost --backup-dir=/backups

# 带参数验证的完整组件检查
./ticheck args-check
```

## 命令

### 常用标志位
- `--host`: TiDB 主机地址
- `--port`: TiDB 端口号
- `--user`: 数据库用户
- `--password`: 数据库密码
- `--log-level`: 设置日志级别 (debug, info, warn, error)

### 可用命令

#### `db-check`
检查数据库连接、性能和配置。

#### `cdc-check`
检查变更数据捕获进程和复制健康状态。

#### `dr-check`
验证灾难恢复设置和备份/恢复流程。

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