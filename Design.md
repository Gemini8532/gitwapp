Project Outline: Local Git Management Web Application1. Architecture Overview**1.1. High-Level Design**
    1.1.1. The application follows a client-server architecture.
    1.1.2. **Reverse Proxy:** Nginx acts as the entry point, handling SSL termination and forwarding requests to the backend.
    1.1.3. **Frontend:** Single Page Application (SPA) built with React, TypeScript, and Tailwind CSS.
    1.1.4. **Backend:** RESTful API built with Golang.
    1.1.5. **Data Persistence:** Simple JSON file storage for application state (user credentials, list of tracked repository paths).
    1.1.6. **Git Engine:** The backend interacts directly with the local file system to manage Git repositories using a Go-native git implementation.
2. Backend Specification (Golang)**2.1. Binary Architecture**
    2.1.1. **Single Binary Design:** The application is built as a single Go binary that operates in different modes based on command-line arguments.
    2.1.2. **Server Mode:** `./gitwapp serve` - Starts the HTTP server with both internal and web API interfaces.
    2.1.3. **CLI Mode:** `./gitwapp <command>` - Executes administrative commands (user/repo management) by making HTTP calls to the running server's internal API.
    2.1.4. **Flag Parsing:** Uses standard library `flag` package or `cobra`/`urfave/cli` for command parsing and argument handling.
    2.1.5. **Command Structure:**
        2.1.5.1. `./gitwapp serve [--port PORT] [--data-dir PATH]` - Start server
        2.1.5.2. `./gitwapp repo add <path>` - Add repository
        2.1.5.3. `./gitwapp repo remove <id>` - Remove repository
        2.1.5.4. `./gitwapp repo list` - List repositories
        2.1.5.5. `./gitwapp user add <username> <password>` - Create user
        2.1.5.6. `./gitwapp user remove <id>` - Delete user
        2.1.5.7. `./gitwapp user passwd <username>` - Update password

**2.2. Core Framework & Libraries**
    2.2.1. **HTTP Server:** `gorilla/mux` (recommended for robust routing and middleware support) with standard `net/http`.
    2.2.2. **Git Implementation:** `go-git/go-git` (v5).
        2.2.2.1. Provides pure Go implementation for git operations without relying on system `git` binaries, ensuring portability.
        2.2.2.2. Fallback: Wrapper around `os/exec` if specific edge-case git hooks or obscure config support is required.
    2.2.3. **Data Storage:** Standard library `encoding/json` for reading/writing JSON files.
    2.2.4. **Authentication:** `golang-jwt/jwt` for stateless JSON Web Token authentication.
    2.2.5. **HTTP Client:** Standard library `net/http` for CLI mode to communicate with internal API.

**2.3. Data Persistence (JSON File Storage)**
    2.3.1. **File: `data/users.json`**
        2.3.1.1. Array of user objects:
        2.3.1.2. `id` (String, UUID)
        2.3.1.3. `username` (String, Unique)
        2.3.1.4. `password_hash` (String, bcrypt)
    2.3.2. **File: `data/repositories.json`**
        2.3.2.1. Array of repository objects:
        2.3.2.2. `id` (String, UUID)
        2.3.2.3. `name` (String, Display name)
        2.3.2.4. `path` (String, Absolute local system path, Unique)
        2.3.2.5. `created_at` (String, ISO 8601 timestamp)
        2.3.2.6. `user_id` (String, owner reference)
    2.3.3. **Write Access Control:** Only the server process (running in serve mode) modifies these JSON files. All configuration changes must go through the server's API endpoints.
    2.3.4. **CLI Operation:** When the binary runs in CLI mode, it makes HTTP requests to `http://localhost:<port>/internal/api` to perform administrative tasks rather than directly modifying JSON files.
    2.3.5. **Backup Strategy:** Periodic backups of JSON files recommended for data safety.


**2.4. API Interfaces**

The server exposes two distinct API interfaces:

**2.4.1. Internal CLI API (Localhost Only)**
    2.4.1.1. **Base Path:** `/internal/api`
    2.4.1.2. **Access Control:** Bound to `localhost` only (127.0.0.1), not accessible remotely.
    2.4.1.3. **Authentication:** None required (relies on localhost binding for security).
    2.4.1.4. **Client:** The same binary running in CLI mode acts as the HTTP client for this API.
    2.4.1.5. **Endpoints:**
        2.4.1.5.1. **Configuration Management**
            2.4.1.5.1.1. `POST /internal/api/repos` - Add a repository to track (path validation included).
            2.4.1.5.1.2. `DELETE /internal/api/repos/:id` - Remove a repository from tracking.
            2.4.1.5.1.3. `GET /internal/api/repos` - List all tracked repositories.
        2.4.1.5.2. **User Management**
            2.4.1.5.2.1. `POST /internal/api/users` - Create a new user account.
            2.4.1.5.2.2. `DELETE /internal/api/users/:id` - Delete a user account.
            2.4.1.5.2.3. `PUT /internal/api/users/:id/password` - Update user password.

**2.4.2. Web API (Public, JWT-Protected)**
    2.4.2.1. **Base Path:** `/api`
    2.4.2.2. **Access Control:** Accessible remotely via Nginx reverse proxy.
    2.4.2.3. **Authentication:** JWT required for all endpoints except `/api/login`.
    2.4.2.4. **Endpoints:**
        2.4.2.4.1. **Auth**
            2.4.2.4.1.1. `POST /api/login` - Validates credentials, returns JWT.
        2.4.2.4.2. **Repositories (Read-Only CRUD)**
            2.4.2.4.2.1. `GET /api/repos` - Returns list of tracked repos with high-level status summary.
        2.4.2.4.3. **Repository Operations (Git Logic)**
            2.4.2.4.3.1. `GET /api/repos/:id/status` - Detailed status (Modified, Staged, Ahead, Behind).
            2.4.2.4.3.2. `POST /api/repos/:id/stage` - Stage files (`git add`).
            2.4.2.4.3.3. `POST /api/repos/:id/commit` - Commit staged changes.
            2.4.2.4.3.4. `POST /api/repos/:id/push` - Push to remote (`origin`).
            2.4.2.4.3.5. `POST /api/repos/:id/pull` - Pull from remote (`origin`).
            2.4.2.4.3.6. `POST /api/repos/:id/fetch` - Update remote refs.
3. Frontend Specification (React + TypeScript)**3.1. Technology Stack**
    3.1.1. **Build Tool:** Vite.
    3.1.2. **Styling:** Tailwind CSS (utility-first styling).
    3.1.3. **State Management:** React Query (TanStack Query) for server state synchronization; React Context for auth state.
    3.1.4. **Routing:** React Router v6.
    3.1.5. **Icons:** Lucide-React or Heroicons.

**3.2. Page 1: Login**
    3.2.1. **Layout:** Centered card on a neutral background.
    3.2.2. **Inputs:** Username, Password.
    3.2.3. **Action:** "Sign In" button performing POST to `/api/login`.
    3.2.4. **Behavior:** On success, store JWT in HTTP-only cookie or local storage and redirect to Dashboard.

**3.3. Page 2: Dashboard (Repository List)**
    3.3.1. **Layout:** Main navigation bar (User profile, Logout) + Grid/List view of repositories.
    3.3.2. **Repository Management:** Repositories can only be added/removed via the CLI tool. The dashboard is read-only for repository configuration.
    3.3.3. **Repository Card Component:**
        3.3.3.1. **Header:** Repository Name (derived from directory name).
        3.3.3.2. **Path:** Visual indicator of filesystem location.
        3.3.3.3. **Status Indicators (Badges):**
            3.3.3.3.1. *Clean:* No local changes.
            3.3.3.3.2. *Dirty:* Uncommitted changes present (Yellow).
            3.3.3.3.3. *Ahead:* Local commits not pushed (Blue Arrow Up).
            3.3.3.3.4. *Behind:* Remote changes not pulled (Orange Arrow Down).
            3.3.3.3.5. *Diverged:* Both ahead and behind (Red Alert).

**3.4. Page 3: Repository Detail View**
    3.4.1. **Header Control Bar:**
        3.4.1.1. Branch selector (Dropdown showing current branch).
        3.4.1.2. "Fetch" button to update remote status.
        3.4.1.3. "Pull" button (disabled if up-to-date).
        3.4.1.4. "Push" button (disabled if nothing to push).
    3.4.2. **Staging Area (Two-Pane or List Layout):**
        3.4.2.1. **Working Directory Changes:** List of modified/untracked files. Action: "Stage" (+ button).
        3.4.2.2. **Staged Changes (Index):** List of files ready to commit. Action: "Unstage" (- button).
    3.4.3. **Commit Section:**
        3.4.3.1. Text area for commit message.
        3.4.3.2. "Commit" button (active only if Staged Changes > 0 and message is not empty).
    3.4.4. **History (Optional/MVP+):** List of recent commits.
4. Technical Implementation Details**4.1. Git Status Logic (Backend Calculation)**
    4.1.1. **Clean/Dirty:** calculated by checking the `Worktree` status in `go-git`.
    4.1.2. **Ahead/Behind:** calculated by comparing `HEAD` reference hash vs `refs/remotes/origin/master` (or main) reference hash.
    4.1.3. **Authentication for Remote Operations:**
        4.1.3.1. The app will rely on the host system's existing SSH keys (e.g., `~/.ssh/id_rsa`) or git credential helper.
        4.1.3.2. `go-git` SSH auth method will load keys from the standard Ubuntu location.

**4.2. Error Handling**
    4.2.1. **Merge Conflicts:** If a pull results in a conflict, the backend returns a 409 Conflict status. The frontend displays a "Manual Intervention Required" alert, instructing the user to resolve conflicts in their IDE/terminal (simplifying MVP scope).
    4.2.2. **Invalid Path:** Backend validates path existence before adding to repositories.json.

**4.3. Security Considerations**
    4.3.1. **Remote Access & Reverse Proxy:**
        4.3.1.1. **TLS Termination:** Nginx handles SSL/TLS termination; communication between Nginx and the Go backend remains HTTP.
        4.3.1.2. **Header Handling:** The Go backend must be configured to trust and parse `X-Forwarded-For`, `X-Real-IP`, and `X-Forwarded-Proto` headers for correct request origin identification.
    4.3.2. **Authentication Enforcement:** As the app is exposed remotely, JWT middleware is strictly enforced on all non-public routes. Rate limiting should be applied at the Nginx level to the `/api/login` endpoint.
    4.3.3. **Path Traversal:** Validate that repository paths are absolute and the user has OS-level read/write permissions.
    4.3.4. **CORS & CSRF:**
        4.3.4.1. If the Frontend and Backend are served on the same domain via Nginx, CORS issues are minimized.
        4.3.4.2. **Cookie Security:** Auth cookies must be set with `SameSite=Strict` and `Secure` attributes (relying on Nginx HTTPS).

