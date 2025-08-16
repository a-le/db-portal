# db-portal

[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://go.dev/dl/)
[![License](https://img.shields.io/github/license/a-le/db-portal)](https://github.com/a-le/db-portal/blob/main/LICENSE)


`db-portal`
## Project Description
**db-portal** is a cross-platform SQL editor with data dictionnary browsing and light ETL features.  
Several DB vendors are supported, as well as CSV, JSON, and XLSX file formats.

db-portal is designed for both solo and multi-user use.  
Whether you need a simple tool for personal database management or a multi-user solution, db-portal aims to provides a flexible and efficient way to interact.

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
- Supports the following DB vendors: ClickHouse, MySQL/MariaDB, MSSQL, PostgreSQL, SQLite and more to come.
- Write SQL queries in a syntax-highlighted minimalist editor
- View query results in a smart HTML table
- Browse data dictionaries (tables, columns, views, procedures)

- ETL features
  - Use a GUI for ETL operations. Set data sources as origin and destination, click submit and voilà ! 
  - Data sources supported as source or destination: DB table, DB query, .json (2 formats supported), .xlsx, .csv 

- Solo or multi-user support
  - Solo: Simply add data sources (DSN), assign them to your user, and start using them.
  - Multi-user: Add users and DSN, then assign DSN to specific users for controlled access.
  - Regular users can only access the data sources and connections assigned to them, whereas admins have unrestricted access to all resources.

- Configurable
  - Modify server configuration easily using a YAML file
  - Manage users and data sources (only database DSN at the moment)
  - Customize the data dictionary UI by editing SQL commands in `conf/commands.yaml`

- Implements industry-standard authentication and security practices
  - Server based with HTTPS support
  - Secure authentication via JWT

- Developer friendly
  - No CGO required for building from source
  - Instantly see changes to `.js` (`.js` files are combined and minified on the fly) 

- Light and efficient
  - Minimal CPU and memory usage
  - Custom JavaScript and CSS using a lightweight virtual DOM library (Mithril.js)

- Cross-platform support: Windows, Linux, Mac OS and other OSes supported by Go
- **see [CHANGELOG.md](https://raw.githubusercontent.com/a-le/db-portal/main/CHANGELOG.md) for latest features added to rolling release**


## App maturity
- > **Warning:** Not recommended for direct internet exposure unless you fully understand the security implications and have performed your own review and hardening.


## Quick Installation
<!--
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
-->
1. For the time being, you should **build the app yourself** from source.
  With Go installed, `go build` in the source directory is all you need !

2. Run executable: `db-portal --set-master-password=your_password`  
<sub>--set-master-password argument is only needed on the first run, or if you need to reset password.</sub>

3. **Open your browser and navigate to** [http://localhost:3000](http://localhost:3000)

4. **Log in with the `admin` user with the password set at step 2**  

---


## Roadmap / upcoming / ideas
- JS codebase reorganization and quality improvements
- Improve integration of ETL / Data copy features
- Add DuckDB support
- Use DuckDB for ETL task of reading XLSX files ?
- Support base folders as data source for files
- Replace CodeMirror by Prism (syntax highligthning) + custom js/mithril editor.
- use github actions for CI
- Load and save query/script files
- Enhance data dictionary functionality
- Support SQL scripts via CLI tools (psql, sqli etc...)
- Act as a http DB proxy for other apps
- add tests
- Split the project into 2 separate repositories: server (Go backend) and client (web frontend) ?


## Built With
- Go language
- Open source libraries  (see [go.mod](https://raw.githubusercontent.com/a-le/db-portal/main/go.mod) for a complete list of dependencies)
- [MithrilJS](https://mithril.js.org/) *a JavaScript framework for building fast and modular applications*
- [CodeMirror](https://codemirror.net/) *a powerful code editor component*
- Custom CSS for styling

## Architecture Notes
- Use RESTful APIs.
- User authentication via JSON Web Tokens (JWT).
- Configuration files auto-reload.
- User queries always use a new, clean connection to the database.
- UI queries will use a connection from the pool if supported.
- Use SQLite for data persistence

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

## Manage users and database connections
Common tasks now have a GUI.
For operations that are not yet supported, 
db-portal uses internally a SQLite database with a few tables.
As shipped, the default `admin` user is allowed to the `SQLite db-portal` data source.
To modify to your needs, you simply have to execute SQL queries.
