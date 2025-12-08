# Architecture Plan: Golang + React (Vite)

This document outlines the strategy for building a web application using a Golang backend and a React/TypeScript frontend (built with Vite and Tailwind). The architecture focuses on a unified deployment model where the Go binary embeds and serves the frontend static assets in production, while maintaining a decoupled developer experience using proxies.

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
├── /nginx                 # Nginx configuration  
│   └── nginx.conf  
├── Makefile               # Build automation  
└── go.mod                 # Go module definition (at Root)
```
Create a `.env` file in the project root. Both Vite (dev) and Go (dev/prod) will read this.

**File:** `.env`
```ini
# Backend Config  
APP_PORT=8080  
APP_ENV=development

# Frontend Config (Prefix with VITE_ to expose to client)  
VITE_APP_NAME="My Go App"
```## 3. Frontend Development (Vite & Proxy)

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

```
## 5. Production Build Strategy

**Makefile:**

```makefile
.PHONY: all build-frontend build-backend run

ifneq (,$(wildcard ./.env))  
    include .env  
    export  
endif

all: build-frontend build-backend

build-frontend:  
	cd frontend && npm install && npm run build

build-backend:  
	@mkdir -p frontend/dist  
	@touch frontend/dist/.keep  
	go build -o bin/server cmd/server/main.go

run-prod: all  
	./bin/server
```

## 6. Nginx Proxy (Production)

**File:** nginx/nginx.conf

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