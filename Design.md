Project Outline: Local Git Management Web Application1. Architecture Overview**1.1. High-Level Design**
    1.1.1. The application follows a client-server architecture.
    1.1.2. **Reverse Proxy:** Nginx acts as the entry point, handling SSL termination and forwarding requests to the backend.
    1.1.3. **Frontend:** Single Page Application (SPA) built with React, TypeScript, and Tailwind CSS.
    1.1.4. **Backend:** RESTful API built with Golang.
    1.1.5. **Database:** MySQL for persistent application state (user credentials, list of tracked repository paths).
    1.1.6. **Git Engine:** The backend interacts directly with the local file system to manage Git repositories using a Go-native git implementation.
2. Backend Specification (Golang)**2.1. Core Framework & Libraries**
    2.1.1. **HTTP Server:** `Gin` or `Echo` (recommended for robust routing and middleware support) or standard `net/http` for simplicity.
    2.1.2. **Git Implementation:** `go-git/go-git` (v5).
        2.1.2.1. Provides pure Go implementation for git operations without relying on system `git` binaries, ensuring portability.
        2.1.2.2. Fallback: Wrapper around `os/exec` if specific edge-case git hooks or obscure config support is required.
    2.1.3. **Database ORM/Driver:** `GORM` or `sqlx` connecting to MySQL.
    2.1.4. **Authentication:** `golang-jwt/jwt` for stateless JSON Web Token authentication.

**2.2. Data Persistence (MySQL Schema)**
    2.2.1. **Table: `users`**
        2.2.1.1. `id` (Primary Key, UUID/Int)
        2.2.1.2. `username` (Varchar, Unique)
        2.2.1.3. `password_hash` (Varchar, bcrypt)
    2.2.2. **Table: `repositories`**
        2.2.2.1. `id` (Primary Key, UUID/Int)
        2.2.2.2. `name` (Varchar, Display name)
        2.2.2.3. `path` (Varchar, Absolute local system path, Unique)
        2.2.2.4. `created_at` (Timestamp)
        2.2.2.5. `user_id` (Foreign Key, owner)

**2.3. API Endpoints**
    2.3.1. **Auth**
        2.3.1.1. `POST /api/login` - Validates credentials, returns JWT.
    2.3.2. **Repositories (CRUD)**
        2.3.2.1. `GET /api/repos` - Returns list of tracked repos with high-level status summary.
        2.3.2.2. `POST /api/repos` - Adds a directory path to be tracked.
        2.3.2.3. `DELETE /api/repos/:id` - Stops tracking a repository (does not delete files).
    2.3.3. **Repository Operations (Git Logic)**
        2.3.3.1. `GET /api/repos/:id/status` - Detailed status (Modified, Staged, Ahead, Behind).
        2.3.3.2. `POST /api/repos/:id/stage` - Stage files (`git add`).
        2.3.3.3. `POST /api/repos/:id/commit` - Commit staged changes.
        2.3.3.4. `POST /api/repos/:id/push` - Push to remote (`origin`).
        2.3.3.5. `POST /api/repos/:id/pull` - Pull from remote (`origin`).
        2.3.3.6. `POST /api/repos/:id/fetch` - Update remote refs.
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
    3.3.2. **Add Repository Modal:** Input field for absolute local path (e.g., `/home/user/dev/project-a`).
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
    4.2.2. **Invalid Path:** Backend validates path existence before adding to MySQL.

**4.3. Security Considerations**
    4.3.1. **Remote Access & Reverse Proxy:**
        4.3.1.1. **TLS Termination:** Nginx handles SSL/TLS termination; communication between Nginx and the Go backend remains HTTP.
        4.3.1.2. **Header Handling:** The Go backend must be configured to trust and parse `X-Forwarded-For`, `X-Real-IP`, and `X-Forwarded-Proto` headers for correct request origin identification.
    4.3.2. **Authentication Enforcement:** As the app is exposed remotely, JWT middleware is strictly enforced on all non-public routes. Rate limiting should be applied at the Nginx level to the `/api/login` endpoint.
    4.3.3. **Path Traversal:** Validate that repository paths are absolute and the user has OS-level read/write permissions.
    4.3.4. **CORS & CSRF:**
        4.3.4.1. If the Frontend and Backend are served on the same domain via Nginx, CORS issues are minimized.
        4.3.4.2. **Cookie Security:** Auth cookies must be set with `SameSite=Strict` and `Secure` attributes (relying on Nginx HTTPS).

