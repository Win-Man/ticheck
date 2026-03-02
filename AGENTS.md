# PROJECT KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** 2194a09 
**Branch:** main

## OVERVIEW
This is tikv-check (ticheck) - a TiDB cluster health check tool written in Go. It performs checks on CDC, DB, DR and other components of TiDB clusters.

## STRUCTURE
```
./
├── cmd/                  # CLI command definitions
├── pkg/                  # Shared utility packages  
├── config/               # Config files (.toml examples)
├── database/             # Database connection utilities
├── service/              # Version/service info
├── main.go               # Entry point
├── Makefile              # Build system
├── go.mod/go.sum         # Dependencies
└── README.md             # Overview
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| CLI Commands | ./cmd/ | args_check, db_check, dr_check, cdc_check |
| DB Utilities | ./database/ | MySQL conn, gorm utilities |
| Config Loading | ./config/ | TOML-based configuration |
| Core Logic | ./pkg/ | API and processing functions |

## CODE MAP
| Symbol | Type | Location | Refs | Role |
|--------|------|----------|------|------|
| main.main | func | main.go | 1 | Entry point |
| cmd.Execute | func | cmd/root.go | main | Root CLI runner |

## CONVENTIONS
- Uses Cobra CLI framework for command structure
- TOML for configuration files
- Standard Go testing with go test
- Makefile for build processes

## ANTI-PATTERNS (THIS PROJECT)
- Package.json in root (only for editor plugin, not functional)
- Compiled binaries in bin/ (should be in .gitignore)

## UNIQUE STYLES
- Uses -race flag in go run commands
- Includes build time/version via linker flags

## COMMANDS
```bash
make build                    # Build binary
make arm64/amd64             # Build specific arch
./ticheck [subcommand]       # Run checks
```

## NOTES
- Built for TiDB ecosystem
- Includes progress bar utilities
- SSH remote exec for TiKV clusters