<h1 align="center">HJSON Server</h1>

<p align="center"><em>A lightweight, file-backed REST API server powered by HJSON—built for local development, frontend prototyping, API mocking, and integration testing.</em></p>

<p align="center">⚡ Lightweight · 💾 File-backed · 🔐 Configurable Auth · 🌱 JSON/HJSON Compatible · 🔄 Live Reload · 🔒 Optional HTTPS</p>

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Install and Build](#install-and-build)
  - [Run the Server](#run-the-server)
- [Configuration](#configuration)
  - [Configuration Priority](#configuration-priority)
  - [CLI Flags](#cli-flags)
  - [Environment Variables](#environment-variables)
- [API Reference](#api-reference)
  - [Resource Routing](#resource-routing)
  - [GET—Retrieve Data](#getretrieve-data)
  - [POST—Append to an Array](#postappend-to-an-array)
  - [PUT—Replace a Value](#putreplace-a-value)
  - [PATCH—Update Object Fields](#patchupdate-object-fields)
  - [DELETE—Remove Data](#deleteremove-data)
  - [Query Parameters](#query-parameters)
- [Authentication](#authentication)
  - [Protected Routes](#protected-routes)
  - [Bearer Authentication](#bearer-authentication)
  - [Basic Authentication](#basic-authentication)
  - [API Key Authentication](#api-key-authentication)
  - [Cookie Authentication](#cookie-authentication)
- [Browser and HTTPS Support](#browser-and-https-support)
  - [CORS](#cors)
  - [HTTPS with Caddy](#https-with-caddy)
- [Persistence and Live Reload](#persistence-and-live-reload)
- [Logging](#logging)
  - [Log Levels](#log-levels)
  - [Verbose Mode](#verbose-mode)
- [Use Cases](#use-cases)
- [License](#license)

---

## Overview

**HJSON Server** turns a local HJSON-compatible data file into a fully featured REST API. It supports CRUD operations, nested resources, filtering, sorting, pagination, authentication, live reload, structured logging, latency simulation, browser-friendly CORS, and optional HTTPS through Caddy.

No database setup or schemas required. Just point the server at a file and start building.

---

## Features

- **REST API from HJSON** — Use `GET`, `POST`, `PUT`, `PATCH`, and `DELETE` against file-backed data.
- **Flexible file input** — Work with `.hjson`, `.json`, `.txt`, or other HJSON-compatible files.
- **Deep nesting** — Traverse object keys, `id` lookups, array indexes, and nested resources.
- **Query support** — Filter, sort, paginate, offset, and simulate latency.
- **Automatic persistence** — Write mutating requests back to the configured data file.
- **Live reload** — Reload external data-file edits without restarting the server.
- **Four authentication strategies** — Choose Bearer, Basic, API key, or Cookie authentication.
- **Protected routes** — Configure which routes require authentication through `auth.hjson`.
- **Browser cookie issuance** — Configure a route that responds with `Set-Cookie` for browser testing.
- **CLI and `.env` configuration** — Let CLI flags override environment values and built-in defaults.
- **Localhost CORS support** — Connect browser frontends running on arbitrary `localhost` or `127.0.0.1` ports, including credentialed requests.
- **Optional HTTPS with Caddy** — Proxy HTTPS traffic to the Go application for local secure-context testing.
- **Structured logging** — Configure log levels and optional verbose runtime metadata.
- **URL normalization** — Normalize malformed paths before routing.

---

## Getting Started

### Prerequisites

- Go
- Optional: Caddy for local HTTPS

---

### Install and Build

```bash
git clone https://github.com/your-github-username/hjson-server.git
cd hjson-server

go mod tidy
go build
```

---

### Run the Server

```bash
# macOS / Linux
./hjson-server

# Windows
.\\hjson-server.exe
```

Run with custom options:

```bash
./hjson-server --db=data.hjson --port=5000 --live-reload --log-level=debug --verbose
```

---

## Configuration

HJSON Server supports CLI flags, environment variables, and `.env` files.

---

### Configuration Priority

Configuration values are resolved in this order, from highest to lowest priority:

1. CLI flags
2. Environment variables or `.env`
3. Built-in defaults

For example, if `.env` contains:

```env
HJSON_PORT=3000
```

Then this command runs the server on port `8080` because the explicit CLI flag takes priority:

```bash
./hjson-server --port=8080
```

---

### CLI Flags

| Flag | Description | Default |
| --- | --- | --- |
| `--db` | Path to the HJSON-compatible data file. | `data.hjson` |
| `--database` | Alias for `--db`. | `data.hjson` |
| `--port` | Port used by the Go HTTP server. | `3000` |
| `--live-reload` | Reload the data file when it changes externally. | `false` |
| `--log-level` | Set the log level to `debug`, `info`, `warn`, or `error`. | `info` |
| `--verbose` | Enable additional runtime logging metadata. | `false` |

---

### Environment Variables

Example `.env` file:

```env
HJSON_DB_PATH=data.hjson
HJSON_PORT=3000
HJSON_LIVE_RELOAD=false
HJSON_LOG_LEVEL=info
HJSON_VERBOSE=false
```

| Variable | Description |
| --- | --- |
| `HJSON_DB_PATH` | Data-file path. |
| `HJSON_PORT` | Server port. |
| `HJSON_LIVE_RELOAD` | Enable or disable live reload. |
| `HJSON_LOG_LEVEL` | Logging level. |
| `HJSON_VERBOSE` | Enable or disable verbose logging. |

---

## API Reference

### Resource Routing

The API structure mirrors the configured data file:

```http
GET /                → entire data file
GET /books           → collection
GET /books/1         → item by id, with index fallback
GET /users/42/posts  → nested resource
```

Resource lookups support:

- Object keys
- `id`-based array matching
- Array-index fallback
- Deep chaining of keys, IDs, and indexes

---

### GET—Retrieve Data

Retrieve the entire data file, a collection, an individual item, or a nested resource:

```http
GET /
GET /books
GET /books/1
GET /users/42/posts
```

---

### POST—Append to an Array

```http
POST /books
Content-Type: application/json
```

```json
{
  "title": "New Book",
  "author": "Jane Doe",
  "year": 2026
}
```

The target path must resolve to an array.

---

### PUT—Replace a Value

```http
PUT /books/1
Content-Type: application/json
```

```json
{
  "title": "Updated Book",
  "author": "Jane Doe",
  "year": 2026
}
```

`PUT` can replace objects, arrays, or primitive values.

---

### PATCH—Update Object Fields

```http
PATCH /books/1
Content-Type: application/json
```

```json
{
  "author": "New Author"
}
```

`PATCH` shallow-merges fields into an existing object.

---

### DELETE—Remove Data

```http
DELETE /books/1
DELETE /books?author=Unknown
DELETE /tags?value=fiction
```

Filter-based deletes use the same matching rules as filtered `GET` requests.

---

### Query Parameters

| Parameter | Description |
| --- | --- |
| `?key=value` | Filter objects by field and value. |
| `?value=foo` | Match primitive values or matching object values. |
| `?sort=key` | Sort in ascending order. |
| `?sort=-key` | Sort in descending order. |
| `?page=N&limit=M` | Use page-based pagination. |
| `?offset=K&limit=M` | Use offset-based pagination. |
| `?delay=2s` | Add artificial response latency. |

Examples:

```http
GET /books?author=Asimov
GET /users?role=admin
GET /products?inStock=true
GET /books?sort=title
GET /books?sort=-year
GET /books?page=2&limit=10
GET /books?offset=20&limit=10
GET /books?delay=2s
```

---

## Authentication

Authentication is configured through a separate `auth.hjson` file. It controls:

- Whether authentication is enabled
- Which authentication strategy is active
- Which routes are protected
- The credentials and settings for each strategy
- Cookie issuance behavior for browser clients

Example:

```hjson
{
    enabled: true

    // Supported: basic, bearer, api-key, cookie
    type: "cookie"

    protectedRoutes: [
        "/settings"
    ]

    basic: {
        username: "admin"
        password: "password"
    }

    bearer: {
        token: "my-local-api-key"
    }

    apiKey: {
        header: "X-API-Key"
        value: "my-local-api-key"
    }

    cookie: {
        issueRoute: "/auth/cookie"
        name: "session"
        value: "my-local-session"
    }
}
```

Only the strategy selected by `type` is used for protected requests.

> [!WARNING]
> `auth.hjson` is intended for development configuration. Do not commit real production credentials or secrets.

---

### Protected Routes

Configure the routes that require authentication:

```hjson
protectedRoutes: [
    "/settings"
    "/admin"
]
```

Unprotected routes continue normally.

---

### Bearer Authentication

Configuration:

```hjson
type: "bearer"

bearer: {
    token: "my-local-api-key"
}
```

Client request:

```http
Authorization: Bearer my-local-api-key
```

---

### Basic Authentication

Configuration:

```hjson
type: "basic"

basic: {
    username: "admin"
    password: "password"
}
```

Client request:

```http
Authorization: Basic <base64(username:password)>
```

---

### API Key Authentication

Configuration:

```hjson
type: "api-key"

apiKey: {
    header: "X-API-Key"
    value: "my-local-api-key"
}
```

Client request:

```http
X-API-Key: my-local-api-key
```

The API-key header name is configurable.

---

### Cookie Authentication

Cookie authentication supports both validating an incoming cookie and issuing that cookie to a browser.

Configuration:

```hjson
type: "cookie"

cookie: {
    issueRoute: "/auth/cookie"
    name: "session"
    value: "my-local-session"
}
```

Issue the cookie:

```http
POST /auth/cookie
```

The server responds with a `Set-Cookie` header. A browser can then include that cookie in later protected requests.

```js
fetch("http://localhost:3000/auth/cookie", {
  method: "POST",
  credentials: "include"
});
```

Then access a protected route:

```js
fetch("http://localhost:3000/settings", {
  credentials: "include"
});
```

---

## Browser and HTTPS Support

### CORS

HJSON Server supports browser frontends running on arbitrary local development ports. For example:

```text
Frontend:     http://localhost:5173
HJSON Server: http://localhost:3000
```

The browser automatically sends:

```http
Origin: http://localhost:5173
```

If the origin is local, the server responds with the exact requesting origin:

```http
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Credentials: true
```

This allows credentialed browser requests without hardcoding a specific frontend port. Supported local origins can include:

```text
http://localhost:5173
http://localhost:5500
http://localhost:8080
http://127.0.0.1:4200
```

The server also handles browser `OPTIONS` preflight requests before they reach the normal request dispatcher.

> [!NOTE]
> CORS is a browser security mechanism. It does not prevent `curl`, Postman, Go clients, or other backend services from sending requests to the server.

---

### HTTPS with Caddy

HJSON Server can run behind a Caddy reverse proxy when local development requires HTTPS:

```text
Browser or client → HTTPS → Caddy → HJSON Server on localhost:3000
```

Add the HTTP and HTTPS ports to `.env`:

```env
HJSON_PORT=3000
HJSON_HTTPS_PORT=3443
```

Create a `Caddyfile` in the project root:

```caddyfile
localhost:{$HJSON_HTTPS_PORT:3443} {
    tls internal
    reverse_proxy localhost:{$HJSON_PORT:3000}
}
```

The values after `:` are defaults. Caddy uses port `3443` for HTTPS and forwards requests to HJSON Server on port `3000` when the corresponding variables are not set.

Start HJSON Server:

```bash
./hjson-server --port=3000
```

In a second terminal, start Caddy and explicitly load the variables from `.env`:

```bash
caddy run --envfile .env
```

> [!IMPORTANT]
> Caddy does not load the project's `.env` file unless it is passed through `--envfile` or the variables are already defined in the shell environment.

The API is now available through HTTPS:

```text
https://localhost:3443/books
```

Because `tls internal` uses Caddy's local certificate authority, the browser may initially display a certificate warning. Trust Caddy's local CA with:

```bash
caddy trust
```

Local HTTPS is useful when testing:

- Browser features that require a secure context
- HTTPS-only application behavior
- Encrypted local traffic
- Clients that expect an HTTPS backend

---

## Persistence and Live Reload

All mutating operations are persisted to the configured data file:

- `POST`
- `PUT`
- `PATCH`
- `DELETE`

Enable live reload with:

```bash
./hjson-server --live-reload
```

External writes to the data file are then reloaded into memory without restarting the server.

Writes initiated by HJSON Server itself are ignored by the live-reload path so the server does not react to its own persistence operations.

---

## Logging

HJSON Server uses Charmbracelet Logger for structured logging.

### Log Levels

Set a log level when starting the server:

```bash
./hjson-server --log-level=debug
```

| Level | Description |
| --- | --- |
| `debug` | Detailed diagnostic information. |
| `info` | Normal application activity. |
| `warn` | Unexpected or potentially problematic states. |
| `error` | Request or application failures. |

---

### Verbose Mode

```bash
./hjson-server --verbose
```

Verbose mode can add metadata such as:

- Uptime
- Process ID
- Goroutine count
- Caller information

Authentication secrets such as bearer tokens, passwords, API keys, and cookie values should never be written to logs.

---

## Use Cases

- **Frontend prototyping** — Build against a realistic local REST backend without setting up a database.
- **API mocking** — Simulate CRUD, nested resources, filtering, pagination, authentication, and latency.
- **Authentication testing** — Exercise Bearer, Basic, API-key, and cookie-based clients.
- **Browser development** — Connect frontend development servers running on arbitrary localhost ports.
- **HTTPS development** — Test local secure-context behavior through Caddy.
- **Integration testing** — Replace unavailable or incomplete backend services with deterministic local data.
- **Teaching and workshops** — Demonstrate REST, authentication, headers, persistence, and client/server interaction with minimal setup.
- **Quick local data store** — Use an HJSON file as a lightweight backing store when a database would be unnecessary.

---

## License

This project is licensed under the **MIT License**.
