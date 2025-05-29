# Changelog

## v0.2.2-beta
- **JSON export now writes to a temporary file instead of processing in memory**
- **CSV export no longer relies on external dependencies**
- **CSV export now writes to a temporary file instead of streaming directly**
- **Added JSON compact export compatible with ClickHouse JSONCompact format**
- **Refactored export logic for all supported formats**
- **Added gzip compression option for export downloads**

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