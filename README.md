# ticheck

**English Version** | [中文版](./README_cn.md)

# Introduction

ticheck is a comprehensive TiDB cluster health check tool written in Go. It performs checks on CDC, DB, DR and other components of TiDB clusters to ensure they're functioning properly.

## Features

- **Database Checks**: Verify database connectivity, performance and configuration
- **CDC Checks**: Monitor Change Data Capture functionality
- **DR Checks**: Validate Disaster Recovery capabilities 
- **Argument Validation**: Comprehensive flag and argument verification

## Prerequisites

- Go 1.19+
- Access to TiDB cluster components

## Installation and Usage

For detailed installation and usage instructions, please refer to the complete guides:

- [English Guide](./README_en.md)
- [Chinese Guide](./README_cn.md)

Quick start commands:

```bash
# Check database connectivity and configuration
./ticheck db-check --host=localhost --port=4000

# Verify CDC components
./ticheck cdc-check --upstream-host=localhost --upstream-port=4000

# Verify disaster recovery setup
./ticheck dr-check --host=localhost --backup-dir=/backups
```

## License

Apache License 2.0 - see the [LICENSE](./LICENSE) file for details.

For more information, please refer to our complete documentation:

- [English Version](./README_en.md)
- [中文版](./README_cn.md)