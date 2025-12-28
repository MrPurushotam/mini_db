# Mini Database

A simple, in-memory key-value store with data persistence, built with Go and the Fiber web framework. This project demonstrates a basic implementation of a database that can store string key-value pairs and recover its state from an Append Only File (AOF).

## Features

- **In-Memory Storage**: Fast key-value operations.
- **RESTful API**: Exposes endpoints for common database operations (Set, Get, Delete, GetAll, GetAllKeys, GetAllValues).
- **Append Only File (AOF) Persistence**: All write operations are logged to a file, allowing the database state to be reconstructed on startup.
- **Configurable Logging**: Structured logging with different levels (Debug, Info, Warn, Error).
- **Environment Variable Configuration**: Easy customization of port, log level, and AOF filename.
- **Concurrency Safe**: Uses RWMutex for safe concurrent access to the data store.

## Project Evolution (Learning Journey)

This project is primarily a learning exercise in Go programming, evolving through different versions:

- **v0: Basic Application**: Focused on building the fundamental in-memory key-value store with basic CRUD operations and a RESTful API.
- **v1: AOF & Global Logger**: Introduced data persistence using an Append Only File (AOF) and integrated a custom global logger for better observability.
- **v2 (Planned): Different Key-Value Pair Types**: Future plans include enhancing the store to support various key-value pair types beyond simple strings, allowing for more complex data structures.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Go (version 1.24.5 or later)

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/mrpurushotam/mini_db.git
    cd mini_database
    ```

2.  **Download dependencies:**
    ```bash
    go mod download
    ```

### Running the Application

You can run the application directly:

```bash
go run cmd/server/main.go
```

The server will start on the configured port (default: `3000`).

## API Endpoints

The API base path is `/api/v0`.

### `POST /api/v0/set`

Sets a key-value pair.

- **Request Body**: `application/json`
  ```json
  {
    "key": "mykey",
    "value": "myvalue"
  }
  ```
- **Response**: `application/json`
  ```json
  {
    "status": "success",
    "message": "ok"
  }
  ```

### `GET /api/v0/get?key={key}`

Retrieves the value for a given key.

- **Query Parameter**: `key` (string)
- **Response (Success)**: `application/json`
  ```json
  {
    "status": "success",
    "value": "myvalue"
  }
  ```
- **Response (Not Found)**: `application/json`
  ```json
  {
    "status": "error",
    "message": "Not found"
  }
  ```

### `DELETE /api/v0/delete?key={key}`

Deletes a key-value pair.

- **Query Parameter**: `key` (string)
- **Response (Success)**: `application/json`
  ```json
  {
    "message": "ok",
    "status": "success"
  }
  ```
- **Response (Not Found/Error)**: `application/json`
  ```json
  {
    "status": "error",
    "message": "Couldn't delete key value pair."
  }
  ```

### `GET /api/v0/get/all`

Retrieves all key-value pairs.

- **Response**: `application/json`
  ```json
  {
    "status": "success",
    "values": {
      "key1": "value1",
      "key2": "value2"
    }
  }
  ```

### `GET /api/v0/keys/all`

Retrieves all keys.

- **Response**: `application/json`
  ```json
  {
    "status": "success",
    "keys": ["key1", "key2"]
  }
  ```

### `GET /api/v0/values/all`

Retrieves all values.

- **Response**: `application/json`
  ```json
  {
    "status": "success",
    "values": ["value1", "value2"]
  }
  ```

### `GET /api/v0/`

Basic API status check.

- **Response**: `application/json`
  ```json
  {
    "message": "Api is running"
  }
  ```

## Configuration

The application can be configured using environment variables:

- `PORT`: The port for the server to listen on. Default: `3000`
- `LOG_LEVEL`: The minimum level for logs to be displayed. Possible values: `debug`, `info`, `warn`, `error`. Default: `info`
- `AOF_FILENAME`: The name of the file used for AOF persistence. Default: `database.aof`

Example `.env` file:

```env
PORT=8080
LOG_LEVEL=debug
AOF_FILENAME=my_database.aof
```

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go       // Main application entry point
├── internal/
│   ├── aof/              // Append Only File implementation for persistence
│   │   └── aof.go
│   ├── config.go         // Application configuration loading
│   ├── handler/          // HTTP request handlers
│   │   └── handler.go
│   ├── logger/           // Custom logging utility
│   │   └── logger.go
│   ├── routes/           // API route definitions
│   │   └── route.go
│   └── store/            // In-memory data store logic
│       └── store.go
├── database.aof          // Default AOF file (created on first run)
├── go.mod                // Go module dependencies
├── go.sum
└── readme.md             // This file
```

## Persistence

The `mini_database` uses an Append Only File (AOF) for data persistence. Every `SET` and `DELETE` operation is logged to the `database.aof` (or configured) file. When the application starts, it reads and replays all operations from this file to reconstruct the last known state of the database. This ensures that data is not lost when the application restarts.
