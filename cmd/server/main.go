package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Gemini8532/gitwapp/internal/api"
	"github.com/Gemini8532/gitwapp/internal/config"
)

// defaultPort is the default port for the server to listen on.
// It can be overridden at build time with:
// go build -ldflags "-X main.defaultPort=8084"
var defaultPort = "8080"

// Build information, set at build time via ldflags.
var (
	version   = "dev"
	buildDate = "unknown"
	gitCommit = "unknown"
)

// main is the entry point of the application. It parses command-line
// arguments and executes the corresponding command.
func main() {
	// Initialize logger ONCE based on environment
	logger := initLogger()
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		runServer()
	case "stop":
		stopServer()
	case "repo":
		handleRepoCommand()
	case "user":
		handleUserCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// initLogger creates a new logger based on the APP_ENV environment variable.
// It returns a JSON logger for "production" and a text logger otherwise.
func initLogger() *slog.Logger {
	env := os.Getenv("APP_ENV")
	if env == "production" {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

// runServer starts the GitWapp server. It parses command-line flags,
// initializes the configuration store, manages the server process,
// and starts the API server.
func runServer() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	port := serveCmd.String("port", defaultPort, "Port to listen on")

	serveCmd.Parse(os.Args[2:])

	// Allow overriding port from environment variable (after flag parsing)
	if envPort := os.Getenv("APP_PORT"); envPort != "" {
		*port = envPort
	}

	store, err := config.NewStore()
	if err != nil {
		slog.Error("Failed to initialize config store", "error", err)
		os.Exit(1)
	}

	if err := manageProcess(store.GetPIDFilePath()); err != nil {
		slog.Warn("Failed to manage process", "error", err)
	}

	buildInfo := api.BuildInfo{
		Version:   version,
		BuildDate: buildDate,
		GitCommit: gitCommit,
	}
	server := api.NewServer(store, buildInfo)
	if err := server.Start(*port); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

// stopServer stops a running GitWapp server process. It reads the PID
// from the PID file and sends a termination signal to the process.
func stopServer() {
	store, err := config.NewStore()
	if err != nil {
		slog.Error("Failed to initialize config store", "error", err)
		os.Exit(1)
	}

	pidPath := store.GetPIDFilePath()
	if err := killExistingServer(pidPath); err != nil {
		slog.Error("Error stopping server", "error", err)
		os.Exit(1)
	}

	if err := os.Remove(pidPath); err != nil && !os.IsNotExist(err) {
		slog.Warn("Failed to remove PID file", "error", err)
	} else {
		slog.Info("Server stopped")
	}
}

// manageProcess ensures that only one instance of the server is running.
// It kills any existing server process and writes the current process ID
// to the PID file.
func manageProcess(pidPath string) error {
	if err := killExistingServer(pidPath); err != nil {
		slog.Error("Failed to kill existing server", "error", err)
		return err
	}

	pid := os.Getpid()
	slog.Info("Writing PID to file", "pid", pid, "path", pidPath)

	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0644); err != nil {
		slog.Error("Failed to write PID to file", "error", err, "pid", pid, "path", pidPath)
		return err
	}

	return nil
}

// killExistingServer stops a running server process by reading its PID
// from the specified PID file and sending it a termination signal.
func killExistingServer(pidPath string) error {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("No existing PID file found, no process to kill")
			return nil
		}
		slog.Error("Failed to read PID file", "error", err)
		return err
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		slog.Error("Invalid PID in file", "error", err, "content", pidStr)
		return fmt.Errorf("invalid PID in file: %w", err)
	}

	slog.Info("Found existing process PID", "pid", pid)

	process, err := os.FindProcess(pid)
	if err != nil {
		// FindProcess should always succeed on Unix, but just in case
		slog.Error("Failed to find process", "pid", pid, "error", err)
		return err
	}

	// On Unix, FindProcess always succeeds. Sending Signal 0 checks existence.
	if err := process.Signal(syscall.Signal(0)); err == nil {
		slog.Info("Killing previous instance", "pid", pid)
		if err := process.Signal(syscall.SIGTERM); err != nil {
			slog.Error("Failed to kill process", "pid", pid, "error", err)
			return fmt.Errorf("failed to kill process: %w", err)
		}
		// Wait for it to exit
		time.Sleep(100 * time.Millisecond)
		slog.Info("Previous instance killed", "pid", pid)
	} else {
		slog.Info("Process not found or already stopped", "pid", pid)
	}

	return nil
}

// printUsage prints the command-line usage instructions to the console.
func printUsage() {
	fmt.Println("Usage: gitwapp <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  serve    Start the HTTP server")
	fmt.Println("  stop     Stop the running HTTP server")
	fmt.Println("  repo     Manage repositories")
	fmt.Println("  user     Manage users")
}
