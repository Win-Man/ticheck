# ticheck

**ticheck** is a comprehensive TiDB cluster health check tool written in Go. It performs checks on CDC, DB, DR and other components of TiDB clusters to ensure they're functioning properly.

## Features

- **Database Checks**: Verify database connectivity, performance and configuration
- **CDC Checks**: Monitor Change Data Capture functionality
- **DR Checks**: Validate Disaster Recovery capabilities 
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

# Verify disaster recovery setup
./ticheck dr-check --host=localhost --backup-dir=/backups

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
Verify Disaster Recovery setups and backup/restore procedures.

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