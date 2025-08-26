# Changelog

## v0.3.1
- **copy data: schema support for DB source/dest**
- **copy data: batch optimization for DB dest (multi values INSERTs)**
- **copy data: set header types (SQL canonical types) for jsontabular dest when source is DB**
- **some refactoring (Go) and bug fixes (Go and JS)**

## v0.3.0-beta
- **Go code refactoring, cleanup and quality improvment**
- **New ETL ("copy data" page) features that supports all data sources (DSN) and .csv/.json/.xlsx files**  
- **New "data sources" pages to manage DSN and users**
- **Drop download "compress to .gz" feature**
- **Simplify authentification by just using JWT + HTTPS**

## v0.2.2-beta
- **Code refactoring, cleanup and quality improvment**
- **JSON export now writes to a temporary file instead of processing in memory**
- **CSV export no longer relies on external dependencies**
- **CSV export now writes to a temporary file instead of streaming directly**
- **Added JSON tabular export compatible**
- **Refactored export logic for all supported formats**
- **Added gzip compression option for export downloads**
- **Security improvements:**
  - Added CSRF protection to all routes (using gorilla/csrf).
  - Hardened HTTP security headers (X-Content-Type-Options, X-Frame-Options, X-XSS-Protection).
  - Improved CORS configuration for safer cross-origin requests.
  - Replaced legacy session cookies with secure, HTTP-only cookies for JWT-based authentication.
- **Drop Firebird support**

## v0.2.1
- **Same as v0.2.1-beta, promoted to stable release**
- **Update install scripts**

## v0.2.1-beta
- **Code refactoring, cleanup and quality improvment**
- **Add DB warmup when server start**
- **Add json export**
- **upgrade to github.com/golang-jwt/jwt/v5**

## v0.2.0

- **Project renamed to `db-portal`**
- **Versioning now better follows [Semantic Versioning](https://semver.org/)**
- **Added `CHANGELOG.md` file**
- **Configuration storage migrated to SQLite**  
  (replaces `users.yaml`, `connections.yaml`, `.htpasswd`)
- **Added ClickHouse support**  
  (using [`github.com/ClickHouse/clickhouse-go/v2`](https://github.com/ClickHouse/clickhouse-go) driver)
- **Code restructuring** for improved maintainability
- **Enhanced SQL statement identification**  
  (better handling of comments and `RETURNING`/`OUTPUT` clauses for query/non-query detection)
- **Minor web UI improvements**