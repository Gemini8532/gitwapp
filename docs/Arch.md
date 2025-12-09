# Architecture Plan: Golang + React (Vite)

This document outlines the strategy for building a web application using a Golang backend and a React/TypeScript frontend (built with Vite and Tailwind). The architecture focuses on a unified deployment model where the Go binary embeds and serves the frontend static assets in production, while maintaining a decoupled developer experience using proxies.

---

**📋 Template Repository**

This repository is a **template** that can be cloned to bootstrap new projects. After cloning:

**Option 1: Automated Setup (Recommended)**
```bash
./bin/init-template.sh <repository-name> <port>

# Example:
./bin/init-template.sh myapp 8080

# Dry run to see what would happen:
./bin/init-template.sh -n myapp 8080
```
This script will:
- Extract your GitHub username from the current remote
- Update the git remote to use the same format (https:// or git@) with your new repository name
- Update the Go module name in `go.mod`
- Update all Go import paths
- Generate the `.env` file with capitalized app name and specified port

**Option 2: Manual Setup**
1. Update the git remote to point to your new repository:
   ```bash
   git remote set-url origin <your-new-repository-url>
   ```
2. Update the Go module name in `go.mod` to match your project.
3. Run `make init-env` to generate your `.env` configuration file.

---

## 1. Project Structure

We use a monorepo-style structure. A `.env` file at the root acts as the single source of truth for dynamic configuration. To support Go's embed, we treat the frontend directory as a Go package.

**Note:** Run `go mod init <module-name>` in the root directory.

```bash
/my-project  
├── /.env                  # Configuration (PORT, ENV, VITE_APP_NAME)  
├── /cmd  
│   └── /server  
│       └── main.go        # Entry point (imports frontend package)  
├── /internal  
│   └── /api               # API handlers  
│   └── /middleware        # HTTP Middleware (Logging, Auth)  
├── /frontend              # React app & Go Embed Package  
│   ├── /public  
│   │   └── favicon.svg    # Adaptive SVG Favicon  
│   ├── /src  
│   │   └── /hooks         # React hooks (e.g., useTitle)  
│   ├── /dist              # Production build output  
│   ├── embed.go           # Go file to embed the dist folder  
│   ├── index.html         # Entry HTML (Injects Title)  
│   ├── vite.config.ts  
│   └── package.json  
├── /templates             # Configuration templates
│   ├── env                # Template for .env file
│   └── nginx.conf         # Template for nginx config
├── Makefile               # Build automation  
└── go.mod                 # Go module definition (at Root)
```

## 2. Environment Configuration

The `.env` file in the project root contains stable configuration that is **checked into version control**. Both Vite (dev) and Go (dev/prod) will read this file.

**Note:** The `.env` file is treated as stable configuration that is rarely changed. It is **not** gitignored and should be committed to the repository. This ensures consistent configuration across all environments.

**Template:** `templates/env`
```ini
# Backend Config  
APP_PORT={{PORT}}  
APP_ENV=development

# Frontend Config (Prefix with VITE_ to expose to client)  
VITE_APP_NAME={{APP_NAME}}
```

**Generated File:** `.env`
```ini
# Backend Config  
APP_PORT=8080  
APP_ENV=development

# Frontend Config (Prefix with VITE_ to expose to client)  
VITE_APP_NAME="Gen App"
```

You can generate this file using the Makefile with optional arguments:

```bash
# Generate with defaults (PORT=8080, APP_NAME="Gen App")
make init-env

# Generate with custom port
make init-env PORT=3000

# Generate with custom app name
make init-env APP_NAME="My Custom App"

# Generate with both custom values
make init-env PORT=3000 APP_NAME="My Custom App"
```

**Important:** All build targets (`build-frontend`, `build-backend`, `nginx-config`) require `.env` to exist. If it's missing, you'll get an error message prompting you to run `make init-env` first.


## 3. Frontend Development (Vite & Proxy)

### A. Proxy Configuration

We configure Vite to load the environment variable from the root `.env` file to determine the proxy target.

**File:** `frontend/vite.config.ts`

### B. Application Identity (Title & Favicon)

For boilerplate efficiency, we use an SVG favicon (which supports CSS dark mode) and inject the App Name into the HTML at build time.

**File:** `frontend/index.html`

**File:** `frontend/public/favicon.svg`

**Optional Helper:** `frontend/src/hooks/useTitle.ts`

## 4. Backend Implementation (Golang with Embed)

### A. Embedding the Assets

Since `//go:embed` cannot use relative paths like `../../`, we place a small Go file *inside* the frontend directory.

**File:** `frontend/embed.go`

### B. Server Implementation (With Structured Logging)

We use `log/slog` for structured logging. We define the `AppName` as a constant here to serve as the single source of truth for directory naming.

**File:** `cmd/server/main.go`

## 5. Production Build Strategy

**Makefile:**

```makefile
.PHONY: all build-frontend build-backend nginx-config init-env run-prod check-env

# Default values for .env generation
PORT ?= 8080
APP_NAME ?= "My Go App"

# Load .env if it exists
ifneq (,$(wildcard ./.env))  
    include .env  
    export  
endif

# Check that .env exists
check-env:
	@test -f .env || (echo "Error: .env file not found. Run 'make init-env' first." && exit 1)

all: check-env build-frontend build-backend

init-env:
	@sed -e 's/{{PORT}}/$(PORT)/g' -e 's/{{APP_NAME}}/$(APP_NAME)/g' templates/env > .env
	@echo "Generated .env with APP_PORT=$(PORT) and VITE_APP_NAME=$(APP_NAME)"

build-frontend: check-env
	cd frontend && npm install && npm run build

build-backend: check-env
	@mkdir -p frontend/dist  
	@touch frontend/dist/.keep  
	go build -o bin/server cmd/server/main.go

nginx-config: check-env
	@sed 's/{{APP_PORT}}/$(APP_PORT)/g' templates/nginx.conf

run-prod: all  
	./bin/server
```

## 6. Nginx Proxy (Production)

To ensure the Nginx configuration stays in sync with the `.env` file, we use a template file and the Makefile.

**Template:** `templates/nginx.conf`

```nginx
server {  
    listen 80;  
    server_name example.com;

    location / {  
        proxy_pass http://localhost:{{APP_PORT}};   
        proxy_http_version 1.1;  
        proxy_set_header Upgrade $http_upgrade;  
        proxy_set_header Connection 'upgrade';  
        proxy_set_header Host $host;  
        proxy_cache_bypass $http_upgrade;  
    }  
}
```

The Makefile includes a target to print the nginx config with the port from `.env`:

```makefile
nginx-config:
	@sed 's/{{APP_PORT}}/$(APP_PORT)/g' templates/nginx.conf
```

**Usage:**
```bash
# Print the nginx config to stdout
make nginx-config

# Save to a file if needed
make nginx-config > /etc/nginx/sites-available/myapp
```

**Example Output:**

```nginx
server {  
    listen 80;  
    server_name example.com;

    location / {  
        proxy_pass http://localhost:8080;   
        proxy_http_version 1.1;  
        proxy_set_header Upgrade $http_upgrade;  
        proxy_set_header Connection 'upgrade';  
        proxy_set_header Host $host;  
        proxy_cache_bypass $http_upgrade;  
    }  
}
```

## 7. Workflow Summary

### Development

1. **Terminal 1 (Go):** `go run cmd/server/main.go`  
   * Logs format: `time=... level=INFO msg=http_request ...`  
   * Data dir: `~/.config/my-web-app`  
2. **Terminal 2 (React):** `npm run dev` (in `frontend/`)  
   * Reads `VITE_APP_NAME` for title.  
   * Proxies `/api` to `localhost:8080`.

### Production

1. Set `APP_ENV=production` in `.env`.  
2. Run `make build-frontend` & `make build-backend`.  
3. Deploy binary.  
   * Binary serves static assets (with favicon and correct title).  
   * Logs JSON format.
   
## 8. Implementation Notes

The following steps and configurations were applied during the project initialization:

### Tailwind CSS v4
The project uses Tailwind CSS v4.
- **Installation:** `npm install -D tailwindcss @tailwindcss/vite postcss autoprefixer`
- **Configuration:**
  - `vite.config.ts` uses the `@tailwindcss/vite` plugin.
  - `src/index.css` imports Tailwind via `@import "tailwindcss";`.

### Build & Automation
- **Makefile:** The `Makefile` handles the build process.
  - `make build-frontend`: Installs dependencies and builds the React app to `frontend/dist`.
  - `make build-backend`: Compiles the Go binary to `bin/server`. It ensures `frontend/dist` exists (with a dummy file if needed) to satisfy `go:embed`.
  - `make run-prod`: Builds both and runs the binary.
- **Gitignore:** A root `.gitignore` is configured to exclude build artifacts like `bin/`, `frontend/dist/`, and `frontend/node_modules/`.

### Go Module
Use libraries from the standard library where possible. If a library is not available, use a well-maintained third-party library.

- **Module Name:** `github.com/Gemini8532/genapp`
- **Embedding:** `frontend/embed.go` allows the Go backend to serve the compiled frontend.




## 9. Testing Strategy

### Frontend (Jest + React Testing Library)
We use Jest for unit and integration testing of the React application.

**Dependencies:**
- `jest`, `ts-jest`, `jest-environment-jsdom`
- `@testing-library/react`, `@testing-library/dom`, `@testing-library/jest-dom`

**Configuration:**
- `frontend/jest.config.ts`: Configures Jest to handle TypeScript and JSDOM.
- `frontend/jest.setup.ts`: Imports `@testing-library/jest-dom` for custom matchers.

**Running Tests:**
```bash
cd frontend && npm test
```

### Backend (Go Testing)
Standard Go testing is used for the backend.

**Running Tests:**
```bash
go test ./...
```
