package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Gemini8532/gitwapp/frontend"
	"github.com/Gemini8532/gitwapp/internal/api"
	"github.com/Gemini8532/gitwapp/internal/middleware"
	"github.com/gorilla/mux"
)

const (
	defaultPort    = "8080"
	defaultDataDir = "data"
)

func main() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	port := serveCmd.String("port", defaultPort, "Port to listen on")
	dataDir := serveCmd.String("data-dir", defaultDataDir, "Directory for data storage")

	repoCmd := flag.NewFlagSet("repo", flag.ExitOnError)

	userCmd := flag.NewFlagSet("user", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("Usage: gitwapp <command> [arguments]")
		fmt.Println("Commands: serve, repo, user")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		serveCmd.Parse(os.Args[2:])
		runServer(*port, *dataDir)
	case "repo":
		repoCmd.Parse(os.Args[2:])
		handleRepoCommand(os.Args[2:])
	case "user":
		userCmd.Parse(os.Args[2:])
		handleUserCommand(os.Args[2:])
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

func runServer(port, dataDir string) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal(err)
	}

	store := api.NewStorage(dataDir+"/users.json", dataDir+"/repositories.json")
	handler := api.NewHandler(store)

	r := mux.NewRouter()

	// Internal API (Localhost only)
	internal := r.PathPrefix("/internal/api").Subrouter()
	internal.Use(middleware.LocalhostOnlyMiddleware)
	internal.HandleFunc("/repos", handler.InternalAddRepoHandler).Methods("POST")
	internal.HandleFunc("/users", handler.InternalAddUserHandler).Methods("POST")

	// Public API
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/login", handler.LoginHandler).Methods("POST")

	// Protected API
	protected := apiRouter.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/repos", handler.ListReposHandler).Methods("GET")
	protected.HandleFunc("/repos/{id}/status", handler.GetRepoStatusHandler).Methods("GET")

	// Frontend Static Files
	// We need to serve the 'dist' folder from the embedded filesystem
	distFS := frontend.GetDistFS()
	// fs.Sub to get the "dist" subtree
	strippedFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatal(err)
	}

	spaHandler := spaHandler{staticFS: strippedFS}
	r.PathPrefix("/").Handler(spaHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port, // Bind to all interfaces for now, or "127.0.0.1:"+port for strict local
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(srv.ListenAndServe())
}

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the correct mode.
type spaHandler struct {
	staticFS fs.FS
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path := r.URL.Path

	// check if the file exists in the static directory
	f, err := h.staticFS.Open(path[1:]) // fs.FS uses relative paths without leading slash
	if err != nil {
		// file does not exist, serve index.html
		index, err := h.staticFS.Open("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer index.Close()
		stat, _ := index.Stat()
		http.ServeContent(w, r, "index.html", stat.ModTime(), index.(io.ReadSeeker))
		return
	}
	defer f.Close()

	stat, _ := f.Stat()
	// verify if it is a file, otherwise serve index.html
	if stat.IsDir() {
		// serve index.html
		index, err := h.staticFS.Open("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer index.Close()
		stat, _ := index.Stat()
		http.ServeContent(w, r, "index.html", stat.ModTime(), index.(io.ReadSeeker))
		return
	}

	// Otherwise, use http.FileServer to serve the static file
	http.FileServer(http.FS(h.staticFS)).ServeHTTP(w, r)
}


// CLI Command Handlers (Client Logic)

func handleRepoCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: repo <add|remove|list> [args]")
		return
	}

	client := &http.Client{}
	baseURL := "http://localhost:" + defaultPort + "/internal/api" // Assuming default port for CLI

	switch args[0] {
	case "add":
		if len(args) < 2 {
			fmt.Println("Usage: repo add <path>")
			return
		}
		path := args[1]
		reqBody, _ := json.Marshal(map[string]string{"path": path})
		resp, err := client.Post(baseURL+"/repos", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			fmt.Println("Failed to add repo. Status:", resp.Status)
			// Read body for error message
			// ...
			return
		}
		fmt.Println("Repository added successfully.")

	case "list":
		// ...
	}
}

func handleUserCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: user <add|remove|passwd> [args]")
		return
	}

	client := &http.Client{}
	baseURL := "http://localhost:" + defaultPort + "/internal/api"

	switch args[0] {
	case "add":
		if len(args) < 3 {
			fmt.Println("Usage: user add <username> <password>")
			return
		}
		username := args[1]
		password := args[2]

		reqBody, _ := json.Marshal(map[string]string{
			"username": username,
			"password": password,
		})

		resp, err := client.Post(baseURL+"/users", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			fmt.Println("Failed to add user. Status:", resp.Status)
			return
		}
		fmt.Println("User added successfully.")
	}
}
