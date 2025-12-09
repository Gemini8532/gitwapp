# Gitwapp: Local Git Management Web Application

Gitwapp is a client-server application designed to manage local Git repositories through a web interface and a command-line interface (CLI). The backend is built with Go, and the frontend is a React/TypeScript Single Page Application.

## Getting Started

### Prerequisites

-   **Go**: Version 1.24 or higher
-   **Node.js & npm**: Required for building the frontend assets.

### Build the Application

First, ensure you have the necessary Go modules and frontend dependencies.

```bash
# Go dependencies are handled automatically by 'go build'
# Frontend dependencies
cd frontend
npm install
cd ..
```

To build the entire application (backend and frontend assets):

```bash
make build
# This will build the frontend and then the Go backend binary.
# The executable will be found at ./bin/server
```

You can also build the backend specifically:

```bash
make build-backend
```

### Configuration

The application uses a `.env` file for configuration. A template is provided at `templates/env`.
You can generate your `.env` file using the Makefile:

```bash
make init-env
# This generates a .env file with default port 8080 and app name "My Go App".
# You can customize these:
# make init-env PORT=3000 APP_NAME="My Custom App"
```
The `.env` file contains settings like `APP_PORT` and `VITE_APP_NAME`.

## Usage

### Starting the Server

The application runs as a single Go binary. To start the HTTP server:

```bash
./bin/server serve
```

By default, the server will attempt to run on the port specified in your `.env` file (default: `8080`).

**Process Control:**
-   If a server instance is already running, starting a new one will automatically attempt to kill the previous instance.
-   To explicitly stop a running server instance without starting a new one:
    ```bash
    ./bin/server stop
    ```

### Command-Line Interface (CLI)

The CLI allows you to manage users and repositories by interacting with the running server's internal API.

#### Repository Management

```bash
# Add a new local Git repository to be tracked by the application
./bin/server repo add /path/to/your/repo

# List all tracked repositories
./bin/server repo list

# Remove a tracked repository by its ID
./bin/server repo remove <repository_id>
```

#### User Management

```bash
# Add a new user account
./bin/server user add <username> <password>

# Remove a user account by its ID
./bin/server user remove <user_id>

# Update a user's password (not yet implemented)
# ./bin/server user passwd <username>
```

### Web API

The server exposes two API interfaces:

-   **Internal CLI API (`/internal/api`)**: Used by the CLI commands, bound to `localhost` only.
-   **Public Web API (`/api`)**: Protected by JWT authentication, accessible via a browser or external clients.
    -   `POST /api/login`: Authenticate and receive a JWT token.
    -   `GET /api/repos`: List tracked repositories.
    -   `GET /api/repos/{id}/status`: Get detailed Git status for a repository.
    -   `POST /api/repos/{id}/stage`: Stage files in a repository.
    -   `POST /api/repos/{id}/commit`: Commit staged changes.
    -   `POST /api/repos/{id}/push`: Push changes to the remote.
    -   `POST /api/repos/{id}/pull`: Pull changes from the remote.

Further details on the API and architecture can be found in `Design.md` and `Arch.md`.
