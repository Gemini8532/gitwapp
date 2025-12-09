package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Gemini8532/gitwapp/internal/api"
	"github.com/Gemini8532/gitwapp/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist, we'll rely on defaults or env vars
	}

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

func runServer() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	port := serveCmd.String("port", "8080", "Port to listen on")
	
	// Allow overriding port from env
	if envPort := os.Getenv("APP_PORT"); envPort != "" {
		*port = envPort
	}

	serveCmd.Parse(os.Args[2:])

	store, err := config.NewStore()
	if err != nil {
		log.Fatalf("Failed to initialize config store: %v", err)
	}

	if err := manageProcess(store.GetPIDFilePath()); err != nil {
		log.Printf("Warning: Failed to manage process: %v", err)
	}

	server := api.NewServer(store)
	log.Fatal(server.Start(*port))
}

func stopServer() {
	store, err := config.NewStore()
	if err != nil {
		log.Fatalf("Failed to initialize config store: %v", err)
	}

	pidPath := store.GetPIDFilePath()
	if err := killExistingServer(pidPath); err != nil {
		fmt.Printf("Error stopping server: %v\n", err)
		os.Exit(1)
	}

	if err := os.Remove(pidPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove PID file: %v\n", err)
	} else {
		fmt.Println("Server stopped.")
	}
}

func manageProcess(pidPath string) error {
	if err := killExistingServer(pidPath); err != nil {
		return err
	}
	return os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0644)
}

func killExistingServer(pidPath string) error {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid PID in file: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		// FindProcess should always succeed on Unix, but just in case
		return nil
	}

	// On Unix, FindProcess always succeeds. Sending Signal 0 checks existence.
	if err := process.Signal(syscall.Signal(0)); err == nil {
		fmt.Printf("Killing previous instance (PID: %d)...\n", pid)
		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		// Wait for it to exit
		time.Sleep(100 * time.Millisecond)
	}
	
	return nil
}

func printUsage() {
	fmt.Println("Usage: gitwapp <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  serve    Start the HTTP server")
	fmt.Println("  stop     Stop the running HTTP server")
	fmt.Println("  repo     Manage repositories")
	fmt.Println("  user     Manage users")
}
