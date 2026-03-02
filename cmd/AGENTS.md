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
├── dr_check.go     # Disaster recovery checks
└── (more coming)
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Root command | root.go | Entry point for all subcommands |
| DB checks | db_check.go | Connect, Ping/health, version info |
| CDC checks | cdc_check.go | Change data capture validation |
| DR checks | dr_check.go | Backup/disaster recovery tests |
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
- Consistent usage of global flags like --host, --port, etc.

## COMMANDS
```bash
ticheck [command] [flags]    # Various check commands
```

## NOTES
- Commands share global configuration
- Commands may use --skip-* flags to skip specific checks