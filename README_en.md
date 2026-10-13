# ticheck

**ticheck** is a comprehensive TiDB cluster health check tool written in Go. It performs checks on CDC, DB, the dr-auto-sync (dual data centers in one city) topology and other components of TiDB clusters to ensure they're functioning properly.

## Features

- **Database Checks**: Verify database connectivity, performance and configuration
- **CDC Checks**: Monitor Change Data Capture functionality
- **dr-auto-sync Topology View**: One-shot visualization of the [Synchronous Replication for Dual Data Centers](https://docs.pingcap.com/tidb/stable/two-data-centers-in-one-city-deployment/) topology — replication mode, `label-key`/`primary` (primary data center), per-store Role from placement rules (voter / learner / voter-follower …), dr-auto-sync runtime state (`sync` / `sync-recover` / `async`), PD members, `wait-store-timeout` and `wait-sync-timeout`, all gathered live from PD HTTP APIs
- **Argument Validation**: Comprehensive flag and argument verification

## Prerequisites

- Go 1.19+
- Access to TiDB cluster components

## Installation

### From Source

```bash
git clone https://github.com/YOUR_USERNAME/ticheck.git
cd ticheck
make build
```

### Using Go Install

```bash
go install github.com/YOUR_USERNAME/ticheck@latest
```

## Quick Start

```bash
# Check database connectivity and configuration
./ticheck db-check --host=localhost --port=4000

# Verify CDC components
./ticheck cdc-check --upstream-host=localhost --upstream-port=4000

# View the dr-auto-sync (dual data centers) topology and runtime state
./ticheck dr-check --config=./config/drcheck_config.toml

# Check all components with arguments validation
./ticheck args-check
```

## Commands

### Common Flags
- `--host`: TiDB host address
- `--port`: TiDB port number
- `--user`: Database user
- `--password`: Database password
- `--log-level`: Set log level (debug, info, warn, error)

### Available Commands

#### `db-check`
Check database connectivity, performance and configurations.

#### `cdc-check` 
Check Change Data Capture processes and replication health.

#### `dr-check`
View the dr-auto-sync (dual data centers in one city) topology and current state in one shot.

It reads PD HTTP APIs (`/pd/api/v1/config`, `/pd/api/v1/stores`,
`/pd/api/v1/replication_mode/status`, `/pd/api/v1/members`,
`/pd/api/v1/config/rules`) and prints:

- `replication-mode` (`dr-auto-sync`, `majority`, etc.)
- The `label-key` / `primary` that identifies the primary data center
- `wait-store-timeout` and `wait-sync-timeout` from `replication-mode.dr-auto-sync`
- A table of every TiKV store, with its location labels, role
  (`voter` / `learner` / etc., derived from placement rules),
  region count and leader count
- dr-auto-sync runtime state (`sync`, `sync-recover`, `async` …)
- PD members with priority and a `Leader` marker

#### `args-check`
Validate the arguments provided to the tool.

## Configuration

Configuration can be loaded from TOML files. Example configuration files are available:

- `config/argscheck_config.example.toml`
- `config/dbcheck_config.example.toml`
- `config/drcheck_config.example.toml`
- `config/cdccheck_config.example.toml`

## Architecture

- **Command Structure**: Built with Cobra CLI framework
- **Database Layer**: Uses GORM ORM with MySQL adapter
- **Process Management**: Custom progress indicators
- **Linux Operations**: Remote execution utilities for distributed systems

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](./LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Commit your changes (`git commit -m 'feat: add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

---
Copyright © 2023-present TiDB Community