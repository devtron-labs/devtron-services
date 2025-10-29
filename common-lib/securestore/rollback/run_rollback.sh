#!/bin/bash

# Cluster Config Rollback Script
# This script provides a convenient way to run the cluster config rollback utility

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Database Encryption Rollback Script"
    echo "==================================="
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -d, --database NAME     Database name (default: orchestrator)"
    echo "  -t, --table NAME        Table to rollback (cluster, gitops_config, docker_artifact_store,"
    echo "                          git_provider, remote_connection_config, all) (default: all)"
    echo "  -i, --id ID             Specific record ID to rollback"
    echo "  -v, --validate          Validate rollback results"
    echo "  -h, --help              Show this help message"
    echo "  --dry-run               Show what would be done without executing"
    echo ""
    echo "Examples:"
    echo "  $0                      # Rollback all tables"
    echo "  $0 -t cluster           # Rollback cluster table only"
    echo "  $0 -t cluster -i 123    # Rollback specific cluster"
    echo "  $0 -v                   # Validate all rollback results"
    echo "  $0 -t cluster -v        # Validate cluster rollback results"
    echo "  $0 -d mydb              # Use different database"
    echo "  $0 --dry-run            # Show what would be done"
    echo ""
    echo "Supported Tables:"
    echo "  cluster                 - config column (EncryptedMap)"
    echo "  gitops_config           - token column (EncryptedString)"
    echo "  docker_artifact_store   - aws_secret_accesskey, password (EncryptedString)"
    echo "  git_provider            - password, ssh_private_key, access_token (EncryptedString)"
    echo "  remote_connection_config - ssh_password, ssh_auth_key (EncryptedString)"
    echo ""
    echo "Environment Variables:"
    echo "  PG_ADDR                 PostgreSQL address (default: 127.0.0.1)"
    echo "  PG_PORT                 PostgreSQL port (default: 5432)"
    echo "  PG_USER                 PostgreSQL username"
    echo "  PG_PASSWORD             PostgreSQL password"
    echo "  PG_DATABASE             PostgreSQL database"
    echo ""
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [[ ! -f "main.go" ]]; then
        print_error "main.go not found. Please run this script from the rollback directory."
        exit 1
    fi
    
    # Check required environment variables
    if [[ -z "$PG_USER" ]]; then
        print_warning "PG_USER environment variable is not set"
    fi
    
    if [[ -z "$PG_PASSWORD" ]]; then
        print_warning "PG_PASSWORD environment variable is not set"
    fi
    
    print_success "Prerequisites check completed"
}

# Function to run the rollback
run_rollback() {
    local database="orchestrator"
    local table="all"
    local record_id=""
    local validate=false
    local dry_run=false
    local go_args=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--database)
                database="$2"
                shift 2
                ;;
            -t|--table)
                table="$2"
                shift 2
                ;;
            -i|--id)
                record_id="$2"
                shift 2
                ;;
            -v|--validate)
                validate=true
                shift
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # Validate table option
    case $table in
        cluster|gitops_config|docker_artifact_store|git_provider|remote_connection_config|all)
            ;;
        *)
            print_error "Invalid table: $table"
            echo "Supported tables: cluster, gitops_config, docker_artifact_store, git_provider, remote_connection_config, all"
            exit 1
            ;;
    esac

    # Build Go arguments
    go_args="-database=$database -table=$table"

    if [[ -n "$record_id" ]]; then
        go_args="$go_args -id=$record_id"
    fi

    if [[ "$validate" == true ]]; then
        go_args="$go_args -validate"
    fi

    # Show what will be executed
    print_info "Configuration:"
    echo "  Database: $database"
    echo "  Table: $table"
    if [[ -n "$record_id" ]]; then
        echo "  Record ID: $record_id"
    else
        echo "  Record ID: All records"
    fi
    echo "  Validate: $validate"
    echo "  Dry run: $dry_run"
    echo ""

    if [[ "$dry_run" == true ]]; then
        print_info "Dry run mode - showing command that would be executed:"
        echo "go run *.go $go_args"
        exit 0
    fi

    # Confirm execution
    read -p "Do you want to proceed? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Operation cancelled"
        exit 0
    fi

    # Execute the rollback
    print_info "Starting rollback operation..."

    if go run *.go $go_args; then
        print_success "Rollback operation completed successfully"
    else
        print_error "Rollback operation failed"
        exit 1
    fi
}

# Main execution
main() {
    echo "Database Encryption Rollback Utility"
    echo "===================================="
    echo ""

    check_prerequisites
    echo ""
    run_rollback "$@"
}

# Run main function with all arguments
main "$@"
