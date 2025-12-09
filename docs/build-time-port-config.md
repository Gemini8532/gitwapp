# Build-Time Port Configuration

## Changes Made

Removed `.env` file dependency and implemented build-time port configuration using Go's `-ldflags`.

## Implementation

### Default Port Variable
```go
// Default port - can be overridden at build time with:
// go build -ldflags "-X main.defaultPort=8084"
var defaultPort = "8080"
```

### Port Priority (from highest to lowest)
1. **Environment variable** `APP_PORT` (runtime override)
2. **Command-line flag** `--port` (for serve command)
3. **Build-time variable** `defaultPort` (baked into binary)
4. **Fallback** `"8080"` (if defaultPort not set at build time)

## Building the Binary

### With Custom Default Port
```bash
go build -ldflags "-X main.defaultPort=8084" -o bin/server ./cmd/server
```

### With Standard Port (8080)
```bash
go build -o bin/server ./cmd/server
```

## Usage Examples

### Running from cmd/server directory
```bash
cd cmd/server
go run . serve              # Uses defaultPort (8080 or build-time value)
go run . serve --port 9000  # Uses 9000
go run . repo list          # Connects to defaultPort
```

### Using environment variable
```bash
APP_PORT=8084 ./bin/server serve
APP_PORT=8084 ./bin/server repo list
```

### Using command-line flag
```bash
./bin/server serve --port 9000
```

## Benefits

1. ✅ **No .env dependency** - Works anywhere without external files
2. ✅ **Build-time configuration** - Port baked into binary during build
3. ✅ **Works with `go run`** - No path issues when running from cmd/server
4. ✅ **Flexible overrides** - Can still override with env vars or flags
5. ✅ **Consistent behavior** - Server and CLI use same default port

## Removed Dependencies

- Removed `github.com/joho/godotenv` import
- Removed `.env` file loading logic
- No more "Failed to load .env file" warnings
