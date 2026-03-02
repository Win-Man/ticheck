# DATABASE DIRECTORY KNOWLEDGE BASE

**Generated:** 2026-03-01
**Commit:** 2194a09 
**Branch:** main

## OVERVIEW
Database connection utilities and ORM wrappers for MySQL/Gorm integration.

## STRUCTURE
```
./database/
├── db.go             # Database connection utilities 
├── db_test.go        # Database connection tests
├── gorm_mysql.go     # GORM MySQL database wrapper
├── gorm_mysql_test.go # GORM connection tests
└── (connection utils)
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Connection pool | db.go | MySQL connection pooling |
| GORM wrapper | gorm_mysql.go | GORM database wrapper |
| Connection tests | db_test.go | Basic connection tests |
| GORM tests | gorm_mysql_test.go | ORM functionality tests |

## CONVENTIONS
- Uses GORM for ORM operations
- Standard Go SQL connection interfaces
- Go testing package for unit tests

## ANTI-PATTERNS (THIS PROJECT)
- No connection closing in defers in some functions  

## UNIQUE STYLES
- Uses MySQL specifically
- Connection timeout configuration

## NOTES
- Database abstraction layer
- Integration tests with actual connections needed