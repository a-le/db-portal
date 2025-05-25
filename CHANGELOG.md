# Changelog

## v0.2.1
- **Code refactoring & cleanup**

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