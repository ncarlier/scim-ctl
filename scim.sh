#!/bin/bash

# SCIM CTL Helper Script
# Simplifies execution of scim-ctl commands with common operations and configuration management

set -e

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCIM_CTL_BINARY="${SCRIPT_DIR}/scim-ctl"
CONFIG_FILE="${HOME}/.scim-ctl.yaml"
ENV_FILE="${SCRIPT_DIR}/.env"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Check if scim-ctl binary exists
check_binary() {
    if [[ ! -f "$SCIM_CTL_BINARY" ]]; then
        log_error "scim-ctl binary not found at $SCIM_CTL_BINARY"
        log_info "Run 'go build -o scim-ctl ./cmd/scim-ctl' to build it first"
        exit 1
    fi
}

# Load environment variables from .env file if it exists
load_env() {
    if [[ -f "$ENV_FILE" ]]; then
        log_info "Loading environment from $ENV_FILE"
        # shellcheck disable=SC1090
        source "$ENV_FILE"
    fi
}

# Show current configuration
show_config() {
    echo -e "${BLUE}Current SCIM Configuration:${NC}"
    echo "  Target: ${SCIM_CTL_TARGET:-'<not set>'}"
    echo "  OIDC Issuer: ${SCIM_CTL_OIDC_ISSUER:-'<not set>'}"
    echo "  OIDC Client ID: ${SCIM_CTL_OIDC_CLIENT_ID:-'<not set>'}"
    echo "  OIDC Client Secret: ${SCIM_CTL_OIDC_CLIENT_SECRET:+<set>}"
    if [[ -z "$SCIM_CTL_OIDC_CLIENT_SECRET" ]]; then
        echo "  OIDC Client Secret: <not set>"
    fi
    echo "  Config File: $CONFIG_FILE"
    echo "  Env File: $ENV_FILE"
}

# Setup configuration interactively
setup_config() {
    echo -e "${BLUE}SCIM CTL Configuration Setup${NC}"
    echo "This will create/update your .env file with SCIM configuration."
    echo

    read -p "SCIM Target URL (e.g., https://example.com/scim/v2): " target
    read -p "OIDC Issuer URL (e.g., https://auth.example.com): " issuer
    read -p "OIDC Client ID: " client_id
    read -s -p "OIDC Client Secret (optional): " client_secret
    echo

    # Create .env file
    cat > "$ENV_FILE" << EOF
# SCIM CTL Configuration
export SCIM_CTL_TARGET="$target"
export SCIM_CTL_OIDC_ISSUER="$issuer"
export SCIM_CTL_OIDC_CLIENT_ID="$client_id"
EOF

    if [[ -n "$client_secret" ]]; then
        echo "export SCIM_CTL_OIDC_CLIENT_SECRET=\"$client_secret\"" >> "$ENV_FILE"
    fi

    log_success "Configuration saved to $ENV_FILE"
    log_info "Run 'source $ENV_FILE' or restart this script to use the new configuration"
}

# Execute scim-ctl with proper error handling
run_scim_ctl() {
    local verbose=""
    if [[ "${VERBOSE:-false}" == "true" ]]; then
        verbose="-v"
    fi
    
    log_info "Executing: scim-ctl $verbose $*"
    "$SCIM_CTL_BINARY" $verbose "$@"
}

# Common operations
list_schemas() {
    log_info "Retrieving SCIM schemas..."
    run_scim_ctl schemas
}

create_user() {
    local data="$1"
    if [[ -z "$data" ]]; then
        log_error "User data is required"
        echo "Usage: $0 create-user '<json_data>'"
        echo "Example: $0 create-user '{\"userName\": \"jdoe\", \"emails\": [{\"value\": \"jdoe@example.com\"}]}'"
        exit 1
    fi
    
    log_info "Creating user..."
    run_scim_ctl create -r user -d "$data"
}

get_user() {
    local user_id="$1"
    if [[ -z "$user_id" ]]; then
        log_error "User ID is required"
        echo "Usage: $0 get-user <user_id>"
        exit 1
    fi
    
    log_info "Retrieving user: $user_id"
    run_scim_ctl get -r user --id "$user_id"
}

search_users() {
    local filter="$1"
    local start_index="${2:-1}"
    local count="${3:-10}"
    
    log_info "Searching users..."
    if [[ -n "$filter" ]]; then
        run_scim_ctl search -r user -q "$filter" -s "$start_index" -i "$count"
    else
        run_scim_ctl search -r user -s "$start_index" -i "$count"
    fi
}

update_user() {
    local user_id="$1"
    local data="$2"
    if [[ -z "$user_id" || -z "$data" ]]; then
        log_error "User ID and data are required"
        echo "Usage: $0 update-user <user_id> '<json_data>'"
        exit 1
    fi
    
    log_info "Updating user: $user_id"
    run_scim_ctl update -r user --id "$user_id" -d "$data"
}

delete_user() {
    local user_id="$1"
    if [[ -z "$user_id" ]]; then
        log_error "User ID is required"
        echo "Usage: $0 delete-user <user_id>"
        exit 1
    fi
    
    read -p "Are you sure you want to delete user $user_id? (y/N): " confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        log_info "Deleting user: $user_id"
        run_scim_ctl delete -r user --id "$user_id"
        log_success "User deleted"
    else
        log_info "Deletion cancelled"
    fi
}

# Interactive mode
interactive_mode() {
    echo -e "${BLUE}SCIM CTL Interactive Mode${NC}"
    echo "Enter commands or 'help' for assistance. Type 'quit' to exit."
    echo
    
    while true; do
        read -p "scim-ctl> " -r command args
        
        case "$command" in
            "help")
                echo "Available commands:"
                echo "  schemas                     - List SCIM schemas"
                echo "  create-user <json>         - Create a user"
                echo "  get-user <id>              - Get user by ID"
                echo "  search-users [filter]      - Search users"
                echo "  update-user <id> <json>    - Update user"
                echo "  delete-user <id>           - Delete user"
                echo "  config                     - Show current configuration"
                echo "  setup                      - Setup configuration"
                echo "  cache-info                 - Show cache information"
                echo "  cache-clear                - Clear authentication cache"
                echo "  verbose [on|off]           - Toggle verbose output"
                echo "  quit                       - Exit interactive mode"
                ;;
            "schemas")
                list_schemas
                ;;
            "create-user")
                create_user "$args"
                ;;
            "get-user")
                get_user $args
                ;;
            "search-users")
                search_users $args
                ;;
            "update-user")
                # Split args into ID and data
                user_id=$(echo "$args" | cut -d' ' -f1)
                data=$(echo "$args" | cut -d' ' -f2-)
                update_user "$user_id" "$data"
                ;;
            "delete-user")
                delete_user $args
                ;;
            "config")
                show_config
                ;;
            "setup")
                setup_config
                load_env
                ;;
            "cache-info")
                cache_info
                ;;
            "cache-clear")
                cache_clear
                ;;
            "verbose")
                case "$args" in
                    "on"|"true"|"1")
                        export VERBOSE=true
                        log_success "Verbose mode enabled"
                        ;;
                    "off"|"false"|"0")
                        export VERBOSE=false
                        log_success "Verbose mode disabled"
                        ;;
                    *)
                        echo "Verbose mode: ${VERBOSE:-false}"
                        ;;
                esac
                ;;
            "quit"|"exit"|"q")
                log_info "Goodbye!"
                break
                ;;
            "")
                # Empty command, continue
                ;;
            *)
                log_error "Unknown command: $command"
                log_info "Type 'help' for available commands"
                ;;
        esac
        echo
    done
}

# Show usage information
show_usage() {
    cat << EOF
SCIM CTL Helper Script

Usage: $0 [OPTIONS] COMMAND [ARGS...]

OPTIONS:
    -v, --verbose       Enable verbose output
    -h, --help          Show this help message

COMMANDS:
    config              Show current configuration
    setup               Interactive configuration setup
    interactive         Enter interactive mode
    
    schemas             List SCIM schemas
    create-user DATA    Create a user with JSON data
    get-user ID         Get user by ID  
    search-users [FILTER] [START] [COUNT]
                        Search users with optional filter and pagination
    update-user ID DATA Update user with JSON data
    delete-user ID      Delete user by ID (with confirmation)
    
    cache-info          Show authentication cache information
    cache-clear         Clear cached authentication tokens
    
    # Direct scim-ctl execution
    exec [ARGS...]      Execute scim-ctl directly with given arguments

EXAMPLES:
    $0 setup
    $0 config
    $0 schemas
    $0 create-user '{"userName": "jdoe", "emails": [{"value": "jdoe@example.com"}]}'
    $0 get-user 12345
    $0 search-users 'userName eq "jdoe"'
    $0 update-user 12345 '{"userName": "johndoe"}'
    $0 delete-user 12345
    $0 interactive
    $0 exec create -r group -d '{"displayName": "Admins"}'

ENVIRONMENT:
    The script looks for configuration in:
    1. Environment variables (SCIM_CTL_*)
    2. .env file in script directory
    3. ~/.scim-ctl.yaml config file

EOF
}

# Parse command line options
VERBOSE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        -*)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
        *)
            break
            ;;
    esac
done

# Show cache information
cache_info() {
    log_info "Checking authentication cache..."
    run_scim_ctl cache info
}

# Clear authentication cache
cache_clear() {
    log_info "Clearing authentication cache..."
    if run_scim_ctl cache clear; then
        log_success "Authentication cache cleared successfully"
    else
        log_error "Failed to clear cache"
        return 1
    fi
}

# Main execution
main() {
    check_binary
    load_env
    
    local command="${1:-interactive}"
    shift || true
    
    case "$command" in
        "config")
            show_config
            ;;
        "setup")
            setup_config
            ;;
        "interactive")
            interactive_mode
            ;;
        "schemas")
            list_schemas
            ;;
        "create-user")
            create_user "$1"
            ;;
        "get-user")
            get_user "$1"
            ;;
        "search-users")
            search_users "$1" "$2" "$3"
            ;;
        "update-user")
            update_user "$1" "$2"
            ;;
        "delete-user")
            delete_user "$1"
            ;;
        "cache-info")
            cache_info
            ;;
        "cache-clear")
            cache_clear
            ;;
        "exec")
            run_scim_ctl "$@"
            ;;
        *)
            log_error "Unknown command: $command"
            echo
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"