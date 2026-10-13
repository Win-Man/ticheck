# ticheck

**English Version** | [中文版](./README_cn.md)

# Introduction

ticheck is a comprehensive TiDB cluster health check tool written in Go. It performs checks on CDC, DB, the dr-auto-sync (dual data centers in one city) topology and other components of TiDB clusters to ensure they're functioning properly.

## Features

- **Database Checks**: Verify database connectivity, performance and configuration
- **CDC Checks**: Monitor Change Data Capture functionality
- **dr-auto-sync Topology View**: One-shot visualization of the [Synchronous Replication for Dual Data Centers](https://docs.pingcap.com/tidb/stable/two-data-centers-in-one-city-deployment/) topology — replication mode, `label-key`/`primary` (primary data center), per-store Role from placement rules (voter / learner / voter-follower …), dr-auto-sync runtime state (`sync` / `sync-recover` / `async`), PD members, `wait-store-timeout` and `wait-sync-timeout`, all gathered live from PD HTTP APIs
- **Argument Validation**: Comprehensive flag and argument verification

## Prerequisites

- Go 1.19+
- Access to TiDB cluster components

## Installation and Usage

For detailed installation and usage instructions, please refer to the complete guides:

- [English Guide](./README_en.md)
- [Chinese Guide](./README_cn.md)

Quick start commands (all subcommands share the same global flags; per-command
settings live in the matching TOML file):

```bash
# Connect to TiDB and run user-defined SQLs from db-check-items
./ticheck db-check --config=./config/dbcheck.toml.example

# Query TiCDC captures / changefeeds via the CDC HTTP API
./ticheck cdc-check --config=./config/cdccheck.toml.example

# View the dr-auto-sync (dual data centers) topology and runtime state
./ticheck dr-check --config=./config/drcheck.toml.example

# Compare TiDB / PD / TiKV configs and TiDB variables against expected values
./ticheck args-check --config=./config/argscheck.toml.example
```

## License

Apache License 2.0 - see the [LICENSE](./LICENSE) file for details.

For more information, please refer to our complete documentation:

- [English Version](./README_en.md)
- [中文版](./README_cn.md)