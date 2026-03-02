# PKG DIRECTORY KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** 2194a09 
**Branch:** main

## OVERVIEW
Shared utility packages for API, process management and Linux-specific operations.

## STRUCTURE
```
./pkg/
├── api.go           # API handler functions
├── process.go       # Process utility functions  
├── logger/          # Logging utilities
├── log/             # Log-related utilities
├── linux/           # Linux-specific operations
└── (utils shared)
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| HTTP API | api.go | API route handlers |
| Process management | process.go | Terminal display, progress |
| Logging | logger/, log/ | Structured logging utils |
| Remote ops | linux/ | SSH command execution |

## CONVENTIONS
- Small focused Go files 
- Simple packages for specific purposes
- Unix-style tool separation

## ANTI-PATTERNS (THIS PROJECT)
- Some packages too small to be individual directories (log/)

## UNIQUE STYLES
- Progress bar implementations
- Linux-specific operation abstractions

## NOTES
- Shared utilities across commands
- Cross-platform considerations needed