# CMD DIRECTORY KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** 2194a09 
**Branch:** main

## OVERVIEW
Command-line interface commands and utilities.

## STRUCTURE
```
./cmd/
├── root.go         # Root command setup with cobra
├── args_check.go   # Arguments validation & parsing
├── db_check.go     # Database connectivity checks  
├── cdc_check.go    # Change data capture checks
├── dr_check.go     # dr-auto-sync (dual data centers in one city) topology & state check
└── (more coming)
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Root command | root.go | Entry point for all subcommands |
| DB checks | db_check.go | Connect, Ping/health, version info |
| CDC checks | cdc_check.go | Change data capture validation |
| dr-auto-sync (dual data centers) topology check | dr_check.go | Reads PD HTTP APIs to print replication-mode, `label-key`/`primary` (primary data center), per-store Role from placement rules, dr-auto-sync runtime state (`sync` / `sync-recover` / `async`), PD members, `wait-store-timeout` and `wait-sync-timeout` (no backup/disaster-recovery logic) |
| Args validation | args_check.go | CLI argument processing |

## CONVENTIONS
- Uses Cobra CLI framework (cmd.Execute(), PersistentFlags)
- All commands follow args-check-cdc-db-dr naming pattern
- Global flags in init() in different files

## ANTI-PATTERNS (THIS PROJECT)
- No error handling consistency across commands
- Mixed responsiblity in some command files

## UNIQUE STYLES
- Progress bars for long operations in checks
- Consistent global flags across subcommands: `--config/-C`, `--log-level/-L`, `--log-path`; per-command options (e.g. `args-check --output`) are declared locally

## COMMANDS
```bash
ticheck [command] [flags]    # Various check commands
```

## NOTES
- Commands share the same global flag set (`--config` / `--log-level` / `--log-path`); per-command knobs (e.g. `args-check --output`) are declared in each `*_check.go`
- All per-check settings are read from TOML via `--config` rather than from individual CLI flags
- No `--skip-*` flags exist in the current implementation