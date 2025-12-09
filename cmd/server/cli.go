package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gemini8532/gitwapp/internal/api"
	"github.com/Gemini8532/gitwapp/pkg/models"
)

// getBaseURL constructs the base URL for the internal API based on the
// configured port. It prioritizes the APP_PORT environment variable and
// falls back to the default port.
func getBaseURL() string {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = defaultPort
	}
	return fmt.Sprintf("http://localhost:%s/internal/api", port)
}

// handleRepoCommand is the entry point for the "repo" command. It parses
// arguments and calls the appropriate function to handle the subcommand.
func handleRepoCommand() {
	if err := runRepoCommand(os.Args, getBaseURL(), os.Stdout); err != nil {
		slog.Error("Error executing repo command", "error", err)
		os.Exit(1)
	}
}

// runRepoCommand executes the repository-related subcommands (add, remove, list).
func runRepoCommand(args []string, baseURL string, out io.Writer) error {
	if len(args) < 3 {
		printRepoHelp(out)
		return nil
	}

	subCmd := args[2]
	switch subCmd {
	case "add":
		if len(args) < 4 {
			printRepoHelp(out)
			return nil
		}
		path := args[3]
		// Expand to absolute path
		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
		reqBody, _ := json.Marshal(api.AddRepoRequest{Path: absPath})
		resp, err := http.Post(baseURL+"/repos", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
		return processResponse(resp, err, out)
	case "remove":
		if len(args) < 4 {
			printRepoHelp(out)
			return nil
		}
		id := args[3]
		// Basic validation - ID should not be empty or look like a path
		if id == "" || id == "." || id == ".." || strings.Contains(id, "/") {
			return fmt.Errorf("invalid repository ID: %q\nUse 'gitwapp repo list' to see repository IDs", id)
		}
		// URL encode the ID to handle special characters
		encodedID := url.PathEscape(id)
		req, _ := http.NewRequest("DELETE", baseURL+"/repos/"+encodedID, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
		return processResponse(resp, err, out)
	case "list":
		resp, err := http.Get(baseURL + "/repos")
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(out, "Error: Server returned %s\n", resp.Status)
			io.Copy(out, resp.Body)
			return fmt.Errorf("server error: %s", resp.Status)
		}
		var repos []models.Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			return fmt.Errorf("error decoding response: %v", err)
		}
		if len(repos) == 0 {
			fmt.Fprintln(out, "No repositories tracked.")
			fmt.Fprintln(out, "Use 'gitwapp repo add <path>' to add a repository.")
		} else {
			for _, r := range repos {
				fmt.Fprintf(out, "%s\t%s\t%s\n", r.ID, r.Name, r.Path)
			}
		}
		return nil
	case "help", "-h", "--help":
		printRepoHelp(out)
		return nil
	default:
		fmt.Fprintf(out, "Unknown repo command: %s\n\n", subCmd)
		printRepoHelp(out)
		return nil
	}
}

// handleUserCommand is the entry point for the "user" command.
func handleUserCommand() {
	if err := runUserCommand(os.Args, getBaseURL(), os.Stdout); err != nil {
		slog.Error("Error executing user command", "error", err)
		os.Exit(1)
	}
}

// runUserCommand executes the user-related subcommands (add, remove, list).
func runUserCommand(args []string, baseURL string, out io.Writer) error {
	if len(args) < 3 {
		printUserHelp(out)
		return nil
	}

	subCmd := args[2]
	switch subCmd {
	case "add":
		if len(args) < 5 {
			printUserHelp(out)
			return nil
		}
		username := args[3]
		password := args[4]
		reqBody, _ := json.Marshal(api.AddUserRequest{Username: username, Password: password})
		resp, err := http.Post(baseURL+"/users", "application/json", bytes.NewBuffer(reqBody))
		return processResponse(resp, err, out)
	case "remove":
		if len(args) < 4 {
			printUserHelp(out)
			return nil
		}
		id := args[3]
		// Basic validation - ID should not be empty or look like a path
		if id == "" || id == "." || id == ".." || strings.Contains(id, "/") {
			return fmt.Errorf("invalid user ID: %q\nUse 'gitwapp user list' to see user IDs", id)
		}
		encodedID := url.PathEscape(id)
		req, _ := http.NewRequest("DELETE", baseURL+"/users/"+encodedID, nil)
		resp, err := http.DefaultClient.Do(req)
		return processResponse(resp, err, out)
	case "passwd":
		fmt.Fprintln(out, "passwd command not implemented yet")
		return nil
	case "list":
		resp, err := http.Get(baseURL + "/users")
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(out, "Error: Server returned %s\n", resp.Status)
			io.Copy(out, resp.Body)
			return fmt.Errorf("server error: %s", resp.Status)
		}
		var users []models.User
		if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
			return fmt.Errorf("error decoding response: %v", err)
		}
		if len(users) == 0 {
			fmt.Fprintln(out, "No users found.")
			fmt.Fprintln(out, "Use 'gitwapp user add <username> <password>' to create a user.")
		} else {
			for _, u := range users {
				fmt.Fprintf(out, "%s\t%s\n", u.ID, u.Username)
			}
		}
		return nil
	case "help", "-h", "--help":
		printUserHelp(out)
		return nil
	default:
		fmt.Fprintf(out, "Unknown user command: %s\n\n", subCmd)
		printUserHelp(out)
		return nil
	}
}

// processResponse handles the HTTP response from the API. It checks for
// errors, and if the request was successful, prints "Success" to the
// output.
func processResponse(resp *http.Response, err error, out io.Writer) error {
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	fmt.Fprintln(out, "Success")
	if resp.ContentLength > 0 || resp.Header.Get("Content-Type") == "application/json" {
		io.Copy(out, resp.Body)
		fmt.Fprintln(out)
	}
	return nil
}

// printRepoHelp prints the help message for the "repo" command.
func printRepoHelp(out io.Writer) {
	fmt.Fprintln(out, "Usage: gitwapp repo <command> [args]")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  add <path>    Add a repository to track")
	fmt.Fprintln(out, "  remove <id>   Remove a repository from tracking")
	fmt.Fprintln(out, "  list          List all tracked repositories")
	fmt.Fprintln(out, "  help          Show this help message")
}

// printUserHelp prints the help message for the "user" command.
func printUserHelp(out io.Writer) {
	fmt.Fprintln(out, "Usage: gitwapp user <command> [args]")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  add <username> <password>   Create a new user")
	fmt.Fprintln(out, "  remove <id>                 Delete a user")
	fmt.Fprintln(out, "  list                        List all users")
	fmt.Fprintln(out, "  passwd <username>           Update user password (not implemented)")
	fmt.Fprintln(out, "  help                        Show this help message")
}
