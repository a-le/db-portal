# goDatabaseAdmin

**Description**: Query your databases through a sleek web interface and browse data dictionaries.

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
- Built with love, leveraging incredible technologies: Go language, [MithrilJS](https://mithril.js.org/), and more.

## Quick Installation
1. Download the executable from `/builds` along with these folders: `/conf`, `/web`, `/sampledb`.
2. Modify the configuration files as needed.
3. Run the executable from the command prompt.
4. Open your browser and navigate to `localhost:3000`.
5. Log in with the `demo` user (password: `demo`).

Alternatively, clone the full repository and build your own executable.

## Roadmap
- Support for SQL scripts and multi-statement queries
- Improve data dictionnary

## Configuration

server.yaml
```yaml
# main configuration file
# ! restart app if you change this file !

# server address
addr: "localhost:3000"  # host:port to listen on. Default is "localhost:3000"

# login file
htpasswd-file: "./conf/.htpasswd"  # use /hash/{string} url to gen a bcrypt hash of a given string.

# JWT - signing and the verifying key
jwt-secret-key: "5Fy&f#cc7&lLhJr_+@"  # you should replace with your own random string
env-jwt-secret-key: # environment variable that will take precedence over jwt-secret-key if set

# DB
db-timeout: 10  # seconds, will abort any queries that take longer than this. Default is 10
max-resultset-length: 500  # maximum number of rows in a resultset. This applies only to the UI, not to file export. Default is 500

# HTTPS support
# use mkcert https://github.com/FiloSottile/mkcert for easy self-signed certificates. 
cert-file:
key-file:
```

connections.yaml
```yaml
# example
# pagila:
#   db-type: postgresql
#   dsn: # postgresql://postgres:password@localhost:5433/pagila
#   env-dsn: POSTGRES_PAGILA_DSN

# demo
Chinook-Sqlite:
  db-type: sqlite3
  dsn: ./sampledb/Chinook_Sqlite_AutoIncrementPKs.sqlite
  env-dsn: # will take precedence over dsn if set

```

users.yaml
```yaml
demo: {
  connections: ["Chinook-Sqlite"]
}

```


.htpasswd  
you can get a suitable bcrypt hash (with salt) at /hash/replace_with_your_password
```code
demo:$2a$04$6dGMCRe9V2wXXnNRfM4twOZN2Le9kRd8TjI9FY4XVP4TSR8UpPdoS

```