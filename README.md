# db-portal

[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://go.dev/dl/)
[![License](https://img.shields.io/github/license/a-le/db-portal)](https://github.com/a-le/db-portal/blob/main/LICENSE)


`db-portal`

**Description**: 
Query all your SQL databases through a minimalist web interface, browse data dictionaries and export data.
Regroup and manage connections to DB and give your users access to them.

## Demo (old v0.2.0 version)
![Loading animation](.github/demo.gif)

## Table of Contents
- [Features](#features)
- [App Maturity](#app-maturity)
- [Quick Installation](#quick-installation)
- [Roadmap](#roadmap)
- [Built With](#built-with)
- [Architecture Notes](#architecture-notes)
- [Server Configuration](#server-configuration)
- [Configuration](#configuration)

## Features
- Query all your databases through a unified web interface in your browser
- Supports the following databases: ClickHouse, Firebird, MySQL/MariaDB, MSSQL, PostgreSQL, and SQLite
- Write SQL queries in a syntax-highlighted minimalist editor
- View query results in a smart HTML table
- Download results as `.csv`, `.xlsx`, or `.json` files, with optional gzip compression
- Supports multiple JSON formats: standard array of objects and ClickHouse JSON compact
- Browse data dictionaries (tables, columns, views, procedures, etc.)

- Implements industry-standard authentication and security practices
  - Server based with HTTPS support
  - Secure authentication via HTTP Basic Auth and JWT
  - CSRF protection, CORS configuration, secure HTTP headers, and HTTP-only cookies

- Solo or multi-user support
  - Solo: Simply add database connections (DSN), assign them to the `admin` user, and start querying.
  - Multi-user: Add users and connections, then assign connections to specific users for controlled access.

- Configurable
  - Modify server configuration easily using a YAML file
  - Manage users and database connections by executing SQL queries (no GUI or API yet)
  - Customize the data dictionary UI by editing SQL commands in `conf/commands.yaml`

- Developer friendly
  - No CGO required for building from source
  - Instantly see changes to `.js` (`.js` files are combined and minified on the fly) 

- Light and efficient
  - Minimal CPU and memory usage
  - File downloads are streamed and can be gzipped
  - Custom JavaScript and CSS using a lightweight virtual DOM library (Mithril.js)
  - Executable is only ~30 MB, including all 6 supported database drivers

- Cross-platform support: Windows, Linux, and other OSes supported by Go
- **see [CHANGELOG.md](https://raw.githubusercontent.com/a-le/db-portal/main/CHANGELOG.md) for latest features added to rolling release**


## App maturity
- > **Warning:** Not recommended for direct internet exposure unless you fully understand the security implications and have performed your own review and hardening.


## Quick Installation

1. **Run the install script**

**Linux/macOS:**  
```bash
curl -sSfL https://raw.githubusercontent.com/a-le/db-portal/main/install/install.sh -o install.sh
bash install.sh
```

**Windows (PowerShell):**  
```powershell
irm https://raw.githubusercontent.com/a-le/db-portal/main/install/install.ps1 -OutFile install.ps1
powershell -File install.ps1
```

2. **Open your browser and navigate to** [http://localhost:3000](http://localhost:3000)

3. **Log in with the `admin` user**  
   Password: `admin`

---


## Roadmap
- codebase reorganization and quality improvements
- Act as a http DB proxy for other apps
- use github actions for CI
- add tests
- file import (csv... ) to existing table or new auto-created table
- table import/export in some JSON format (with definition + data)
- Support SQL scripts via CLI tools (psql, sqli etc...)
- Load and save query/script files
- Enhance data dictionary functionality
- APIs to manage users and connections
- Split the project into 2 separate repositories: server (Go backend) and client (web frontend) ?
- Add Oracle and DuckDB support ?

## Built With
- Go language
- Open source libraries  (see [go.mod](https://raw.githubusercontent.com/a-le/db-portal/main/go.mod) for a complete list of dependencies)
- [MithrilJS](https://mithril.js.org/) *a JavaScript framework for building fast and modular applications*
- [CodeMirror](https://codemirror.net/) *a powerful code editor component*
- Custom CSS for styling

## Architecture Notes
- Use RESTful APIs.
- User authentication via HTTP(s) Basic Auth and JSON Web Tokens (JWT).
- Configuration files auto-reload.
- User queries always use a new, clean connection to the database.
- UI queries will use a connection from the pool if supported.


## Server configuration

server.yaml
```yaml
# main configuration file
# ! restart app if you change this file !
# server address
addr: "localhost:3000"  # host:port to listen on. Default is "localhost:3000"
# databases
max-resultset-length: 500  # maximum number of rows in a resultset. This applies only to the UI, not to file export. Default is 500
# HTTPS support
# use mkcert https://github.com/FiloSottile/mkcert for easy self-signed certificates. 
cert-file:
key-file:
```

## Manage users and database connections by executing SQL queries
db-portal uses a SQLite database with 3 tables to store those informations.
Default data consist of an `admin` user who can use an sqlite3 connection named `db-portal`.
To modify to your needs, you simply have to execute SQL queries.
See SQL queries below for common tasks.

Changing default password for admin user.  
```sql
update user set pwdhash = '' where name = 'admin'
```
> you can gen a pwdhash from this endpoint: hash/replace-with-your-password  
ex: https://localhost:3000/hash/ThisIsBadPassword 

Adding a user
```sql
insert into user (name, pwdhash) values ('demo', 'password-hash');

```

Adding a database connection
use DSN format accepted by Go drivers
```sql
insert into connection (name, dbtype, dsn) values ('movies', 'sqlite3', 'file path');

```

Associate an existing user with an existing database connection
```sql
insert into user_connection (user_id, connection_id) 
  select u.id, c.id 
  from   user u join connection c
  where  u.name = 'demo' and c.name = 'movies';

```