# goDatabaseAdmin 
`go-database-admin`

**Description**: Query your SQL databases through a minimalist web interface, browse data dictionaries and perform admin tasks efficiently.

## Demo
![Loading animation](.github/demo.gif)

## Features
- Multi-database support (PostgreSQL, MySQL/MariaDB, MSSQL, Firebird, and SQLite)
- Export query results to `.csv` or `.xlsx` files
- Install locally (single-user) or on a server (multi-user)
- User authentication with HTTP Basic Auth and JSON Web Tokens (JWT)
- HTTPS support
- RESTful API access
- Cross-platform support: Windows, Linux, and other OSes supported by Go
- The admin part is coming soon... stay tuned !

## Quick Installation
1. Download the executable from [Releases](../../releases) along with these folders: `/conf`, `/web`, `/sampledb`.
2. Modify the configuration files as needed.
3. Run the executable from the command prompt.
4. Open your browser and navigate to `localhost:3000`.
5. Log in with the `demo` user (password: `demo`).

Alternatively, clone the full repository and build your own executable.

## Roadmap
- Save workspace settings to browser local storage
- Support for DSN without specifying a database for MySQL, MSSQL, and PostgreSQL
- Support SQL scripts and multi-statement queries
- Enhance data dictionary functionality
- Enable database backup and restore
- Load and save query/script files

## Objectives
- Deliver the full power of SQL and command-line DB utilities provided by database distributions, all in a clean and fast web interface.

## Built With
- Go language (see `go.mod` for a complete list of dependencies)
- [MithrilJS](https://mithril.js.org/) (a JavaScript framework for building fast and modular applications)
- [CodeMirror](https://codemirror.net/) (a powerful code editor component)
- Custom CSS for styling

## Configuration

server.yaml
```yaml
# main configuration file
# ! restart app if you change this file !

# server address
addr: "localhost:3000"  # host:port to listen on. Default is "localhost:3000"

# login file
htpasswd-file: "./.htpasswd"  # default "./.htpasswd" will look for the file in conf directory. Use absolute path otherwise.

# databases
max-resultset-length: 500  # maximum number of rows in a resultset. This applies only to the UI, not to file export. Default is 500

# HTTPS support
# use mkcert https://github.com/FiloSottile/mkcert for easy self-signed certificates. 
cert-file:
key-file:
```


connections.yaml  
*Add as many connections as you like.*
```yaml
# example
# pagila:                                                     # that's the name you'll see in the UI
#   db-type: postgresql                                       # valid values: firebird, mysql, mssql, postgresql, sqlite3
#   dsn: postgresql://user:password@localhost:5433/pagila     # DSN, all format supported. Database should be set in the DSN. 
#   env-dsn: POSTGRES_PAGILA_DSN                              # Environment variable name. Which value will take precedence over dsn if set

# a sqlite3 sample database (https://github.com/lerocha/chinook-database)
Chinook-Sqlite:
  db-type: sqlite3
  dsn: ./sampledb/Chinook_Sqlite_AutoIncrementPKs.sqlite
  env-dsn:


```


users.yaml  
*Add as many users as you like. List connections available to user.*
```yaml
demo: {
  connections: ["Chinook-Sqlite"]
}

```


.htpasswd  
*Each user needs a entry there.  
you can get a suitable bcrypt hash (with salt) at /hash/replace_with_your_password*
```code
demo:$2a$04$6dGMCRe9V2wXXnNRfM4twOZN2Le9kRd8TjI9FY4XVP4TSR8UpPdoS

```