#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

dry_run() {
    echo -e "${BLUE}[DRY RUN]${NC} $1"
}

# Parse arguments
DRY_RUN=false
if [[ "$1" == "-n" ]] || [[ "$1" == "--dry-run" ]]; then
    DRY_RUN=true
    shift
fi

# Function to show usage
usage() {
    echo "Usage: $0 [-n|--dry-run] <repository-name> <port>"
    echo ""
    echo "Arguments:"
    echo "  repository-name  Name of the new repository (lowercase, e.g., 'myapp')"
    echo "  port            Port number for the application (e.g., '8080')"
    echo ""
    echo "Options:"
    echo "  -n, --dry-run   Show what would be done without making changes"
    echo ""
    echo "Example:"
    echo "  $0 myapp 8080"
    echo "  $0 -n myapp 8080  # dry run"
    exit 1
}

# Check arguments
if [ $# -ne 2 ]; then
    error "Expected 2 arguments, got $#"
    echo ""
    usage
fi

REPO_NAME="$1"
PORT="$2"

# Validate repository name (lowercase only)
if [[ ! "$REPO_NAME" =~ ^[a-z0-9_-]+$ ]]; then
    error "Repository name must be lowercase and contain only letters, numbers, hyphens, and underscores"
    exit 1
fi

# Validate port number
if [[ ! "$PORT" =~ ^[0-9]+$ ]]; then
    error "Port must be a number"
    exit 1
fi

# Check if we're in a git repository
if [ ! -d .git ]; then
    error "Not in a git repository. Please run this script from the repository root."
    exit 1
fi

# Get current remote and extract username
CURRENT_REMOTE=$(git remote get-url origin 2>/dev/null || echo "")
if [ -z "$CURRENT_REMOTE" ]; then
    error "No git remote 'origin' found"
    exit 1
fi

# Extract username and base URL from current remote
# Handles both git@github.com:user/repo.git and https://github.com/user/repo.git
if [[ "$CURRENT_REMOTE" =~ git@github\.com:([^/]+)/([^/]+)(\.git)?$ ]]; then
    USERNAME="${BASH_REMATCH[1]}"
    REMOTE_BASE="git@github.com:"
    REMOTE_SUFFIX=".git"
elif [[ "$CURRENT_REMOTE" =~ https://github\.com/([^/]+)/([^/]+)(\.git)?$ ]]; then
    USERNAME="${BASH_REMATCH[1]}"
    REMOTE_BASE="https://github.com/"
    REMOTE_SUFFIX="${BASH_REMATCH[3]}"  # Preserve .git suffix if present
else
    error "Could not extract username from remote: $CURRENT_REMOTE"
    exit 1
fi

# Get current module path
CURRENT_MODULE=$(grep "^module " go.mod | awk '{print $2}')

# Generate new values
GITHUB_PATH="$USERNAME/$REPO_NAME"
NEW_MODULE="github.com/$GITHUB_PATH"
NEW_REMOTE="${REMOTE_BASE}${GITHUB_PATH}${REMOTE_SUFFIX}"

# Capitalize first letter of app name
APP_NAME="$(echo ${REPO_NAME:0:1} | tr '[:lower:]' '[:upper:]')${REPO_NAME:1}"

echo "========================================="
echo "  Template Initialization"
echo "========================================="
echo ""
echo "Current configuration:"
echo "  Git remote: $CURRENT_REMOTE"
echo "  Go module:  $CURRENT_MODULE"
echo ""
echo "New configuration:"
echo "  Repository: $REPO_NAME"
echo "  Username:   $USERNAME"
echo "  Git remote: $NEW_REMOTE"
echo "  Go module:  $NEW_MODULE"
echo "  App name:   $APP_NAME"
echo "  Port:       $PORT"
echo ""

if [ "$DRY_RUN" = true ]; then
    dry_run "Would update git remote origin to: $NEW_REMOTE"
    dry_run "Would update go.mod module path to: $NEW_MODULE"
    dry_run "Would update Go import paths from '$CURRENT_MODULE' to '$NEW_MODULE'"
    dry_run "Would run: make init-env APP_NAME=\"$APP_NAME\" PORT=$PORT"
    echo ""
    info "Dry run complete. No changes were made."
    exit 0
fi

read -p "Proceed with these changes? (y/N): " CONFIRM

if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    warn "Aborted by user"
    exit 0
fi

echo ""
info "Updating git remote origin..."
git remote remove origin
git remote add origin "$NEW_REMOTE"
info "Git remote updated to: $NEW_REMOTE"

echo ""
info "Updating go.mod module path..."
sed -i "s|^module .*|module $NEW_MODULE|" go.mod
info "Go module updated to: $NEW_MODULE"

echo ""
info "Updating Go import paths in source files..."
find . -type f -name "*.go" -not -path "./vendor/*" -exec sed -i "s|$CURRENT_MODULE|$NEW_MODULE|g" {} +
info "Import paths updated"

echo ""
info "Generating .env file with new app name and port..."
if [ -f .env ]; then
    warn ".env already exists, backing up to .env.bak"
    mv .env .env.bak
fi
make init-env APP_NAME="$APP_NAME" PORT="$PORT"
info ".env created with APP_NAME=\"$APP_NAME\" and PORT=$PORT"

echo ""
echo "========================================="
echo -e "${GREEN}✓ Template initialization complete!${NC}"
echo "========================================="
echo ""
echo "Next steps:"
echo "  1. Review the changes: git status"
echo "  2. Update .env file if needed"
echo "  3. Create the repository on GitHub: https://github.com/new"
echo "  4. Push to your new repository: git push -u origin main"
echo "  5. Delete this script: rm bin/init-template.sh"
echo ""
