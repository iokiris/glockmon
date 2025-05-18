
# GlockMon — Long Mutex Lock Monitoring for Go

GlockMon is a library for monitoring long mutex locks in Go.  
It allows you to see which locks are delayed, obtain their categories, stack traces, and statistics.

---

## Features

- Automatic collection of long lock data (`MonitoredMutex`)
- Lock categorization
- Built-in HTTP server with JSON API
- Stack trace caching
- Easy integration and minimal overhead

---

## Quick Start

See the usage example in the `example/main.go` file.

It demonstrates how to create a monitor, several mutexes with categories, and run goroutines with locks.

---

## HTTP API

By default, the server listens on `:8080`. Available endpoints:

- **GET** `/blocked` — list of current long locks. Example response:

  ```json
  [
    {
      "id": "5384828326025797004",
      "category": "Category-A",
      "wait_ms": 825,
      "timestamp": "1747527334215577812"
    },
    {
      "id": "11857636764564668017",
      "category": "Category-C",
      "wait_ms": 1807,
      "timestamp": "1747527334818556020"
    }
  ]
  ```

- **GET** `/categories` — statistics by lock categories. Example response:

  ```json
  [
    {
      "category": "Category-B",
      "count": 32,
      "average_wait_ms": 812,
      "total_wait_ms": 26007
    },
    {
      "category": "Category-A",
      "count": 29,
      "average_wait_ms": 968,
      "total_wait_ms": 28073
    },
    {
      "category": "Category-C",
      "count": 62,
      "average_wait_ms": 1243,
      "total_wait_ms": 77086
    }
  ]
  ```

- **GET** `/stacks/{id}` — stack trace text by lock ID. Example output:

  ```
  goroutine 60 [running]:
  glockmon.getStackTrace()
      /app/stacktrace.go:12 +0x53
  glockmon.(*MonitoredMutex).Lock(0xc0000b2c30)
      /app/monitored_mutex.go:60 +0x9f
  main.main.func1(0x7, 0xc0002884d0, 0xc0000b2c30, {0x6da918, 0xa})
      /app/example/main.go:54 +0x92
  created by main.main.func2 in goroutine 20
      /app/example/main.go:94 +0x317
  ```

---

## Configuration

### HTTPConfig

Holds the URL paths for HTTP monitoring endpoints:

- `Blocked` — endpoint to retrieve the list of long locks (e.g. `/blocked`)
- `Categories` — endpoint to retrieve lock statistics grouped by categories (e.g. `/categories`)
- `Stack` — endpoint prefix to retrieve the stack trace by ID (e.g. `/stacks/`)

  > Note: The trailing slash in `Stack` is important since the stack ID is appended to this path.

### MonitorConfig

Main configuration struct for the lock monitor and HTTP server with the following fields:

- `KeepRecords` (bool): Whether to keep detailed lock records in memory after unlock.
- `DefaultCategory` (string): Default category name assigned to locks if none is specified.
- `HTTPServerAddr` (string): Address and port where the HTTP server listens, e.g., `":8080"` or `"0.0.0.0:8080"`.
- `HTTPEndpoints` (HTTPConfig): Holds the URL paths for HTTP monitoring endpoints.

### Default Configuration

The `Default()` function returns a MonitorConfig instance with sensible default values:

- `KeepRecords`: **false**
- `DefaultCategory`: `"GLOBAL"`
- `HTTPServerAddr`: `"0.0.0.0:8080"`
- `HTTPEndpoints`:
  - `Blocked`: `"/blocked"`
  - `Categories`: `"/categories"`
  - `Stack`: `"/stacks/"`

---

Use this config to customize how GlockMon collects lock data and serves monitoring HTTP endpoints.
The config is set via the `MonitorConfig` struct with fields:

- `KeepRecords` — whether to keep detailed lock records in memory after the mutex unlocks
- `DefaultCategory` — default category for TrackedMutex.
- `HTTPServerAddr` — HTTP server address, e.g. `":8080"`
- `HTTPEndpoints` — endpoints paths

You can get the default config via `config.Default()`.

---
