# Gitwapp: Local Git Management Web Application

Gitwapp is a client-server application designed to manage local Git repositories through a web interface and a command-line interface (CLI). The backend is built with Go using the Gorilla router, and the frontend is a React/TypeScript Single Page Application.

## Features

- **Single Binary**: One executable that serves both as HTTP server and CLI tool
- **Build-Time Configuration**: Port and settings baked into the binary during build
- **Dual API**: Internal localhost-only API for CLI, public JWT-protected API for web
- **Simple Storage**: JSON file-based configuration (no database required)
- **Comprehensive CLI**: Full help system with `--help` flags
- **Modern Logging**: Uses Go's standard `slog` package throughout

## Getting Started

### Prerequisites

- **Go**: Version 1.21 or higher
- **Node.js & npm**: Required for building the frontend assets
- **Make**: For using the build system

### Quick Start

1. **Initialize configuration**:
   ```bash
   make init-env
   # Generates .env with default port 8084
   # Customize: make init-env PORT=3000 APP_NAME="My App"
   ```

2. **Build everything**:
   ```bash
   make all
   # Builds frontend and backend with port from .env baked in
   ```

3. **Start the server**:
   ```bash
   ./bin/server serve
   ```

### Build Options

```bash
# Build just the backend (port from .env is baked into binary)
make build-backend

# Build just the frontend
make build-frontend

# Build everything
make all

# Run in production mode
make run-prod
```

### Configuration

The `.env` file is the source of truth for configuration:
- **Build time**: Port is read from `.env` and baked into the binary via `-ldflags`
- **Runtime**: Can still override with `APP_PORT` environment variable or `--port` flag

**Priority (highest to lowest)**:
1. `APP_PORT` environment variable
2. `--port` command-line flag (for serve command)
3. Build-time default (from `.env`)

## Usage

### Starting the Server

```bash
# Start with port from build
./bin/server serve

# Override port at runtime
./bin/server serve --port 9000

# Or with environment variable
APP_PORT=9000 ./bin/server serve
```

**Process Management**:
- Automatically kills previous instance when starting
- PID file stored in config directory
- Explicit stop command available:
  ```bash
  ./bin/server stop
  ```

### Command-Line Interface

The CLI communicates with the running server's internal API (localhost-only).

#### Getting Help

```bash
# General help
./bin/server

# Command-specific help
./bin/server repo help
./bin/server user help
```

#### Repository Management

```bash
# Add a repository
./bin/server repo add /path/to/your/repo

# List all repositories
./bin/server repo list

# Remove a repository
./bin/server repo remove <repository_id>

# Show help
./bin/server repo help
```

#### User Management

```bash
# Add a user
./bin/server user add <username> <password>

# List all users
./bin/server user list

# Remove a user
./bin/server user remove <user_id>

# Show help
./bin/server user help
```

**Note**: User password updates are not yet implemented.

### Web API

The server exposes two API interfaces:

#### Internal CLI API (`/internal/api`)
- **Access**: Localhost only (127.0.0.1)
- **Authentication**: None required
- **Purpose**: CLI commands
- **Endpoints**:
  - `GET /internal/api/repos` - List repositories
  - `POST /internal/api/repos` - Add repository
  - `DELETE /internal/api/repos/{id}` - Remove repository
  - `GET /internal/api/users` - List users
  - `POST /internal/api/users` - Add user
  - `DELETE /internal/api/users/{id}` - Remove user

#### Public Web API (`/api`)
- **Access**: Remote access via Nginx reverse proxy
- **Authentication**: JWT required (except `/api/login`)
- **Purpose**: Web frontend
- **Endpoints**:
  - `POST /api/login` - Authenticate and receive JWT
  - `GET /api/repos` - List tracked repositories
  - `GET /api/repos/{id}/status` - Get Git status
  - `POST /api/repos/{id}/stage` - Stage files
  - `POST /api/repos/{id}/commit` - Commit changes
  - `POST /api/repos/{id}/push` - Push to remote
  - `POST /api/repos/{id}/pull` - Pull from remote

## Architecture

- **Backend**: Go with Gorilla Mux router
- **Frontend**: React + TypeScript + Vite + Tailwind CSS
- **Storage**: JSON files in `~/.config/gitwapp/`
  - `users.json` - User credentials (bcrypt hashed)
  - `repositories.json` - Tracked repository paths
- **Logging**: Standard library `slog` with structured logging
- **Git Operations**: `go-git/go-git` (pure Go implementation)

## Documentation

Detailed documentation is available in the `docs/` directory:

- **[Design.md](docs/Design.md)** - Architecture and API specification
- **[Arch.md](docs/Arch.md)** - Implementation details and structure
- **[build-time-port-config.md](docs/build-time-port-config.md)** - Port configuration details
- **[logging-refactor-summary.md](docs/logging-refactor-summary.md)** - Logging implementation
- **[cli-help-added.md](docs/cli-help-added.md)** - CLI help system

## Development

```bash
# Run from source (uses defaultPort from code)
cd cmd/server
go run . serve

# Run with custom port
APP_PORT=8084 go run ./cmd/server serve

# Build with custom port
go build -ldflags "-X main.defaultPort=8084" -o bin/server ./cmd/server
```

## License

Free for all.