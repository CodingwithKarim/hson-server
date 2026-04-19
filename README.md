<h1 align="center">HSON Server</h1>
<p align="center"><em>Drop-in replacement for <a href="https://json-server.dev/" target="_blank">json-server</a> with <a href="https://hjson.github.io" target="_blank">HJSON</a> support, deep nesting, live reload, and structured logging.</em></p>

<p align="center">
  ⚡ Lightweight · 💾 File-backed · 🌱 JSON/HJSON Compatible · 🧪 Built for Mock APIs
</p>

---

**HSON Server** is a flexible HTTP server that loads data from a local .hson, .json, .txt, or any HJSON-compatible file, and instantly spins up a live REST API. Designed as a drop-in replacement for the popular <a href="https://json-server.dev/" target="_blank">json-server</a>, it supports deep nesting, advanced filtering, automatic file updates, live reload support, and styled logs powered by <a href="https://charm.sh/blog/the-charm-logger/" target="_blank">Charmbracelet</a>.

No database setup. No schemas. Just point it at a file and run the server.

---

## Table of Contents

- [Features](#features)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [API Guide](#api-guide)
- [Logging](#logging)
- [Use Cases](#use-cases)
- [License](#license)

---

## Features

🔹 **Drop-In JSON Server Alternative**  
Supports all standard HTTP verbs (<code>GET</code>, <code>POST</code>, <code>PUT</code>, <code>PATCH</code>, <code>DELETE</code>) ideal for mocking APIs and simulating backend services.

🔹 **Flexible File Input**  
Load data from <code>.hson</code>, <code>.json</code>, <code>.txt</code>, or any compatible file powered by the <a href="https://hjson.github.io" target="_blank">HJSON</a> parser.

🔹 **Deep Nesting**  
Access objects and arrays at any depth, with support for <code>id</code>-based lookups and fallback indexing when no <code>id</code> is present.

🔹 **Query Parameters**  
 Supports filtering by any object field or primitive value (<code>?value=foo</code>), sorting by any key (<code>?sort=key</code> or <code>?sort=-key</code>), and pagination using <code>?page=N&limit=M</code> or <code>?offset=K&limit=M</code>.
 
🔹 **Automatic File Persistence**  
All changes made through the API are instantly written to the original data file. No database or manual saving is required.

🔹 **Live Reload Mode**  
Optional live reload syncs data file changes immediately to memory. Great for making manual edits without restarting the server.

🔹 **Configurable via CLI**  
Customize port, data file, log level, and toggle live reload or verbose logging using command-line flags.

🔹 **Styled Logging**  
Clean, structured logs powered by <a href="https://charm.sh/blog/the-charm-logger/" target="_blank">Charmbracelet Logger</a> with support for log levels, timestamps, and verbosity option.

🔹 **Clean Middleware & Routing**  
Automatically normalizes messy or invalid paths like <code>////api////books///1///</code> into clean, valid routes like <code>/api/books/1</code>.

🔹 **Response Delay (Latency Simulation)**
Simulate network latency by adding a delay query parameter to any request (e.g. ?delay=2s). The server waits before processing, helping test loading states and slow network behavior. Supports standard Go duration formats with built-in safety and cancellation handling.

---

## Getting Started

#### 🔧 Prerequisites

- Go / Golang

#### 🛠 Installation

```bash
# Clone the repository (or download it as a ZIP from GitHub)
git clone https://github.com/your-github-username/hson-server.git

# Navigate into project directory where main.go is located
cd hson-server

# Install required dependencies
go mod tidy

# Build the executable
go build

# Run the server

# For macOS / Linux
./hson-server

# For Windows
.\hson-server.exe

# With CLI args
# Use either ./ or .\ to run executable depending on OS
./hson-server --db=data.hson --port=5000 --live-reload --log-level=debug --verbose
```

---

## Usage

After building the executable, run the server and customize behavior using CLI flags:

```bash
hson-server [flags]
```

#### 🧩 Available Flags

| Flag                   | Description                                                                                             |
|------------------------|---------------------------------------------------------------------------------------------------------|
| `--db`                 | Path to the data file (`.hson`, `.json`, `.txt`, etc). Defaults to `data.hson`.                        |
| `--port`               | Port the server will listen on. Defaults to `3000`.                                                     |
| `--live-reload`        | Enables live reload: syncs file changes to memory on-the-fly.                                           |
| `--log-level`          | Sets the log level: `debug`, `info`, `warn`, `error`, `fatal`.                                          |
| `--verbose`            | Enables verbose logging: includes uptime, PID, goroutines, etc.                                         |

---

#### ▶️ Basic Run

```bash
hson-server
```

#### ⚙️ Custom File and Port

```bash
hson-server --db=mock_data.hson --port=8080
hson-server --db="C:\Documents\mock-data.hson" --port=8080
```

#### 🔄 Enable Live Reload + Logging

```bash
hson-server --live-reload --log-level=debug --verbose
```

> 💡 On macOS/Linux, use `./hson-server`  
> 💡 On Windows, use `.\hson-server.exe`

You can mix and match CLI flags as needed.

---

## API Guide

Once the server is running, you can interact with it using standard HTTP methods. The API structure mirrors your data file (by default data.hson), with collections, nested objects, and array items are all mapped to RESTful routes.

```http
GET /                → Root (returns entire data file)
GET /books           → Collection (array)
GET /books/1         → Single item (by `id` if available, or by index fallback)
GET /users/42/posts  → Nested resource (supports deep nesting)
```

#### Lookups support:
- id-based match (if object contains `"id"`)
- fallback to index (e.g., `/books/0`)
- deep chaining of object keys, id matches, and array indexes

#### 🔎 Query Parameters

Advanced query parameters allow filtering, sorting, and pagination on any collection or array.

```http
GET /[collection]?[key]=[value]&sort=[field]&limit=[N]&offset=[K]
```

#### 🧩 Supported Parameters

| Parameter           | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| `?key=value`        | Filter results by matching any field or value.                             |
| `?value=foo`        | Match against primitive values or object fields equal to `foo`.            |
| `?sort=key`         | Sort results in ascending order by `key`.                                  |
| `?sort=-key`        | Sort results in descending order by `key`.                                 |
| `?page=N&limit=M`   | Paginate results using page-based logic (1-indexed).                       |
| `?offset=K&limit=M` | Paginate using offset-based logic (0-indexed).                             |

#### ▶️ Filtering Examples

```http
GET /books?author=Asimov
GET /users?role=admin
GET /products?inStock=true
GET /tags?0=fiction
```
---

### 📥 GET – Retrieve Data

Fetch entire collections, specific items, or nested data.

```http
GET /books
GET /books/1
GET /users/42/posts
```

Supports:
- Full object lookups by `id`
- Fallback indexing for arrays when ID prop is not found
- Deep nesting through chained keys, ID matches, and indexes.
---

### ➕ POST – Append to a Collection

Use `POST` to append any value to an array at a given path. The server accepts any type as an appended value but the URL path must point to an array.

```http
POST /books
```

#### 📦 Request Body

```json
{
  "title": "New Book",
  "author": "Jane Doe",
  "year": 2025
}
```

💡 Appends the value to the `/books` array.

---

### ✏️ PUT – Replace Value at a Path

Use `PUT` to overwrite the value at a specific path. Can replace entire arrays, objects, or primitive values.

```http
PUT /books/1
PUT /data/settings/theme
```

#### 📦 Request Body

```json
{
  "title": "Updated Book",
  "author": "Jane Doe",
  "year": 2024
}
```

💡 Fully replaces whatever is currently at the target path with the request body.

---

### 🔧 PATCH – Update Object Fields

Use `PATCH` to shallow merge fields into an existing object. Only works for maps (objects) and not for arrays or primitives.

```http
PATCH /books/1
```

#### 📦 Request Body

```json
{
  "author": "New Author"
}
```

💡 Merges fields into the object at `/books/1`.

---

### 🗑️ DELETE – Remove Data

Use `DELETE` to remove:

- A single object by ID or index
- Filtered objects (bulk delete)
- Primitive array values

```http
DELETE /books/1
DELETE /books?author=Unknown
DELETE /tags?value=fiction
```

💡 Filter-based deletes respect the same rules as GET as you can delete by any field or primitive match.

---

### 💾 Persistence Behavior

- All write operations (`POST`, `PUT`, `PATCH`, `DELETE`) are automatically persisted to the original `.hson` or `.json` file.
- `POST` appends any value (object, primitive, etc.) to an array. It only works on paths that resolve to arrays.
- `PUT` is more flexible since it overwrites the entire value at the given path (including primitives, maps, or arrays).
- `PATCH` only shallow-merges into existing **objects** (not arrays or primitives).
- ⚠️ Live-reload only applies to **manual edits** to the file. Edits made via the API do not trigger reloads (to prevent infinite write loops).

---

## Logging

HSON Server includes structured, styled logging powered by [Charmbracelet Logger](https://charm.sh/blog/the-charm-logger/), with rich metadata and dynamic verbosity options.

#### 🧪 Default Log Format

Each incoming HTTP request is logged with key metadata:

```log
[INFO] GET /books
→ method=GET path=/books duration=2ms
```

![image](https://github.com/user-attachments/assets/6fbf4767-aca1-49b5-9a81-8aba6153a81d)


#### ⚙️ Log Levels

Use the `--log-level` flag to control the minimum level of logs shown:

| Level   | Description                      |
|---------|----------------------------------|
| `debug` | Most verbose, includes all logs |
| `info`  | Default, shows normal activity  |
| `warn`  | Warnings or unexpected states   |
| `error` | Only errors and failures        |
| `fatal` | Critical errors only            |

```bash
hson-server --log-level=debug
```

---

#### 🔍 Verbose Mode

Add `--verbose` to include extended runtime metadata:

- Uptime
- Goroutine count
- Process ID (PID)

```bash
hson-server --verbose
```

Example output with verbose logging enabled:

```log
[INFO] GET /books
→ method=GET path=/books status=200 duration=3ms uptime=42s pid=1042 goroutines=6
```
![image](https://github.com/user-attachments/assets/11610138-4ed2-49cc-a1d5-2897699d5fd2)


---

#### 🧼 Clean Paths

The logger also auto-cleans malformed or messy URL paths:

```http
////api////books///1/// → /api/books/1
```

Each cleaned path is logged transparently for debugging purposes.

---

## Use Cases

HSON Server is ideal for a variety of development and testing scenarios:

- **Frontend Prototyping**
  → Mock out a backend using a local `.hson` or any HJSON compatible file. No DB or separate backend service required.

- **API Mocking for Testing**
  → Simulate REST APIs with full CRUD support, nested paths, and advanced filters using a simple data file.

- **Live API Demos**
  → Build interactive UI demos with a realistic backend feel using live data updates and structured logs without trouble of building backend service.

- **Teaching / Workshops**
  → Use simple HJSON files to teach REST API principles and JSON request/response structure without deploying backend service.

- **Stubbing Microservices**
  → Replace unavailable or unstable backend services with file based HSON server during development or integration testing if needed.

- **Quick Dev Tooling**
  → Use HSON Server as a local config data store that you can treat like a mini local database without needing something like SQLite. Shoutout SQLite though.

---

## License
This project is licensed under the **MIT License**. See [here](https://mit-license.org/) for details.
