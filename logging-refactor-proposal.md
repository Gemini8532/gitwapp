# Logging Refactor Proposal

## Current Issues

The current logging implementation has several problems:

1. **Logger creation is duplicated everywhere** - The same logger creation logic (`slog.New(...)` with environment check) appears in:
   - `cmd/server/main.go` - 5 different functions (createLogger, runServer, stopServer, manageProcess, killExistingServer)
   - `internal/api/server.go` - Server.Start method
   - `internal/middleware/logging.go` - 4 different places (LoggingMiddleware, ContextLogger fallback, WithLogger)

2. **Inconsistent usage patterns**:
   - Some code creates loggers directly
   - Some code uses `middleware.ContextLogger(r.Context())` 
   - Some code calls `createLogger()`

3. **Performance overhead** - Creating new logger instances repeatedly, especially in middleware that runs on every request

4. **No logger in context in many places** - The `WithLogger` middleware exists but isn't being used, so `ContextLogger` always falls back to creating a new logger

## Proposed Solution

### 1. Single Logger Initialization in `main()`

Create the logger once at application startup and use it throughout:

```go
// cmd/server/main.go
func main() {
    // Load .env file if it exists
    if err := godotenv.Load(); err != nil {
        // It's okay if .env doesn't exist
    }

    // Initialize logger ONCE based on environment
    logger := initLogger()
    
    // Set as default logger for slog package
    slog.SetDefault(logger)

    // Rest of main logic...
}

func initLogger() *slog.Logger {
    env := os.Getenv("APP_ENV")
    if env == "production" {
        return slog.New(slog.NewJSONHandler(os.Stdout, nil))
    }
    return slog.New(slog.NewTextHandler(os.Stdout, nil))
}
```

### 2. Use slog's Built-in Context Functions

No custom wrapper needed! The `slog` package already provides context-aware functions:
- `slog.InfoContext(ctx, msg, args...)`
- `slog.ErrorContext(ctx, msg, args...)`
- `slog.WarnContext(ctx, msg, args...)`
- `slog.DebugContext(ctx, msg, args...)`

These functions automatically use the logger from the context if present, or fall back to the default logger.

### 3. Simplified Middleware

Update `internal/middleware/logging.go`:

```go
package middleware

import (
    "log/slog"
    "net/http"
    "time"
)

// LoggingMiddleware adds structured logging for HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Create a response writer that captures the status code
        lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        // Call the next handler
        next.ServeHTTP(lrw, r)
        
        // Calculate duration
        duration := time.Since(start)
        
        // Log the request details using top-level slog functions
        slog.Info("http_request",
            "method", r.Method,
            "url", r.URL.Path,
            "status", lrw.statusCode,
            "duration", duration.Milliseconds(),
            "remote_addr", r.RemoteAddr,
            "user_agent", r.UserAgent(),
        )
    })
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}
```

Remove the entire `ContextLogger` and `WithLogger` functions - they're not needed!

### 4. Update Handler Usage

In handlers (e.g., `internal/api/handlers_repo.go`), replace:

```go
logger := middleware.ContextLogger(r.Context())
logger.Info("Adding repository", "path", req.Path)
```

With:

```go
ctx := r.Context()
slog.InfoContext(ctx, "Adding repository", "path", req.Path)
slog.ErrorContext(ctx, "Failed to add", "error", err)
```

### 5. Update CLI and Server Code

In `cmd/server/main.go` and `cmd/server/cli.go`, replace all logger creation with direct `slog` calls:

```go
// Where context is available (use slog context functions):
slog.InfoContext(ctx, "message", "key", value)
slog.ErrorContext(ctx, "error occurred", "error", err)

// Where context is NOT available (use slog top-level functions):
slog.Info("Starting server", "port", port)
slog.Error("Failed to start", "error", err)
```

## Benefits

1. **Single source of truth** - Logger created once in `main()`
2. **Cleaner syntax** - Use `slog.Info()`, `slog.InfoContext()` directly
3. **Better performance** - No repeated logger creation
4. **Simpler code** - Remove all the duplicated logger creation logic AND custom wrapper
5. **Standard library only** - Uses `slog.SetDefault()` and built-in `slog` functions
6. **Consistent patterns** - `slog.*Context(ctx, ...)` when context available, `slog.*()` when not
7. **No custom code needed** - Everything is in the standard library!

## Migration Steps

1. Update `main()` to call `initLogger()` and `slog.SetDefault()`
2. Simplify `internal/middleware/logging.go` to use `slog.Info()` directly
3. Remove `createLogger()` function from `cmd/server/main.go`
4. Replace all `middleware.ContextLogger()` calls with `slog.*Context()`
5. Replace all direct logger creation with `slog.*` calls
6. Remove `internal/middleware/logging.go` functions: `ContextLogger` and `WithLogger`
7. **No need to create `internal/log/log.go`** - we don't need it!
8. Update imports across the codebase

## Example Usage Patterns

### In HTTP Handlers (with context)
```go
func (s *Server) handleAddRepo(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    slog.InfoContext(ctx, "Adding repository", "path", req.Path)
    
    // Validate path...
    if err != nil {
        slog.ErrorContext(ctx, "Failed to add repository", "error", err)
        http.Error(w, "Failed to add", http.StatusInternalServerError)
        return
    }
    
    slog.InfoContext(ctx, "Repository added successfully", "id", newRepo.ID)
    // ...
}
```

### In CLI Commands (no context)
```go
func handleRepoCommand() {
    if err := runRepoCommand(os.Args, getBaseURL(), os.Stdout); err != nil {
        slog.Error("Error executing repo command", "error", err)
        os.Exit(1)
    }
}
```

### In Server Startup (no context)
```go
func runServer() {
    // Parse flags...
    
    slog.Info("Starting server", "port", *port)
    
    if err := server.Start(*port); err != nil {
        slog.Error("Server failed to start", "error", err)
        os.Exit(1)
    }
}
```

### In Process Management (no context)
```go
func manageProcess(pidPath string) error {
    if err := killExistingServer(pidPath); err != nil {
        slog.Error("Failed to kill existing server", "error", err)
        return err
    }
    
    pid := os.Getpid()
    slog.Info("Writing PID to file", "pid", pid, "path", pidPath)
    // ...
}
```
