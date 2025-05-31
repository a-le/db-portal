package types

type DBVendor string

const (
	DBVendorClickHouse DBVendor = "clickhouse"
	DBVendorFirebird   DBVendor = "firebird"
	DBVendorMySQL      DBVendor = "mysql"
	DBVendorMariaDB    DBVendor = "mariadb"
	DBVendorMSSQL      DBVendor = "mssql"
	DBVendorPostgres   DBVendor = "postgresql"
	DBVendorSQLite     DBVendor = "sqlite3"
)
