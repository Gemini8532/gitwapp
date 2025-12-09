package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/Gemini8532/gitwapp/internal/api"
	"github.com/Gemini8532/gitwapp/pkg/models"
)

func getBaseURL() string {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	return fmt.Sprintf("http://localhost:%s/internal/api", port)
}

func handleRepoCommand() {
	if err := runRepoCommand(os.Args, getBaseURL(), os.Stdout); err != nil {
		slog.Error("Error executing repo command", "error", err)
		os.Exit(1)
	}
}

func runRepoCommand(args []string, baseURL string, out io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("Usage: gitwapp repo <add|remove|list> [args]")
	}

	subCmd := args[2]
	switch subCmd {
	case "add":
		if len(args) < 4 {
			return fmt.Errorf("Usage: gitwapp repo add <path>")
		}
		path := args[3]
		reqBody, _ := json.Marshal(api.AddRepoRequest{Path: path})
		resp, err := http.Post(baseURL+"/repos", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			return fmt.Errorf("failed to connect to server: %w", err)
		}
		return processResponse(resp, err, out)
	case "remove":
		if len(args) < 4 {
			return fmt.Errorf("Usage: gitwapp repo remove <id>")
		}
		id := args[3]
		req, _ := http.NewRequest("DELETE", baseURL+"/repos/"+id, nil)
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
		for _, r := range repos {
			fmt.Fprintf(out, "%s\t%s\t%s\n", r.ID, r.Name, r.Path)
		}
		return nil
	default:
		return fmt.Errorf("Unknown repo command: %s", subCmd)
	}
}

func handleUserCommand() {
	if err := runUserCommand(os.Args, getBaseURL(), os.Stdout); err != nil {
		slog.Error("Error executing user command", "error", err)
		os.Exit(1)
	}
}

func runUserCommand(args []string, baseURL string, out io.Writer) error {
	if len(args) < 3 {
		return fmt.Errorf("Usage: gitwapp user <add|remove|passwd> [args]")
	}

	subCmd := args[2]
	switch subCmd {
	case "add":
		if len(args) < 5 {
			return fmt.Errorf("Usage: gitwapp user add <username> <password>")
		}
		username := args[3]
		password := args[4]
		reqBody, _ := json.Marshal(api.AddUserRequest{Username: username, Password: password})
		resp, err := http.Post(baseURL+"/users", "application/json", bytes.NewBuffer(reqBody))
		return processResponse(resp, err, out)
	case "remove":
		if len(args) < 4 {
			return fmt.Errorf("Usage: gitwapp user remove <id>")
		}
		id := args[3]
		req, _ := http.NewRequest("DELETE", baseURL+"/users/"+id, nil)
		resp, err := http.DefaultClient.Do(req)
		return processResponse(resp, err, out)
	case "passwd":
		fmt.Fprintln(out, "passwd command not implemented yet")
		return nil
	default:
		return fmt.Errorf("Unknown user command: %s", subCmd)
	}
}

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
