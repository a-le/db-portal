# db-portal
`db-portal`

**Description**: 
Query all your SQL databases through a minimalist web interface, browse data dictionaries and export data.
Regroup and manage connections to DB and give your users access to them.

## Demo
![Loading animation](.github/demo.gif)

## Features
- Multi-database support : Clickhouse, Firebird, MySQL/MariaDB, MSSQL, PostgreSQL, SQLite
- Install locally (single-user) or on a server (multi-user) with HTTPS support
- Cross-platform support: Windows, Linux, and other OSes supported by Go
- Export data to `.csv` or `.xlsx` files
- View query results in an smart table
- Adapt data dict UI to your needs by simply editing sql commands (see conf/commands.yaml)

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

Alternatively, you can clone the full repository and build your own executable.

## Roadmap
- codebase reorganization and quality improvements (almost done)
- csv file import
- Support SQL scripts via CLI tools
- Load and save query/script files
- Enhance data dictionary functionality
- Act as a http DB proxy
- APIs to manage users and connections
- Split the project into 2 separate repositories: server (Go backend) and client (web frontend) ?
- Oracle and DuckDB support ?

## Built With
- Go language (see `go.mod` for a complete list of dependencies)
- [MithrilJS](https://mithril.js.org/) *a JavaScript framework for building fast and modular applications*
- [CodeMirror](https://codemirror.net/) *a powerful code editor component*
- Custom CSS for styling

## Architecture Notes
- Use RESTful APIs.
- User authentication via HTTP(s) Basic Auth and JSON Web Tokens (JWT).
- Configuration files auto-reload.
- User queries always use a new, clean connection to the database.
- UI queries will use a connection from the pool if supported.
- Dev: no build step for JavaScript: a new `main.min.js` is automatically built on any `*.js` change.
- Dev: no CGO dependencies

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

## Configuration
db-portal use a sqlite DB to store its conf.
To modify conf to your needs, you simply have to issue queries to the embedded sqlite3 DB `db-portal`.

Changing default password for admin user.  
you can gen a pwdhash from this url: http://localhost:3000/hash/replace-with-your-password
```sql
update user set pwdhash = '' where name = 'admin'

```

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