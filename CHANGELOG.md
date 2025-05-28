# Changelog

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