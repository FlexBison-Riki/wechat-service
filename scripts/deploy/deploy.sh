#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
APP_NAME="wechat-service"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/$APP_NAME}"
BACKUP_DIR="${BACKUP_DIR:-/opt/$APP_NAME/backups}"
DOCKER_COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-docker-compose.yml}"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_warn "Not running as root. Some operations may fail."
    fi
}

# Create directories
setup_dirs() {
    log_info "Creating directories..."
    mkdir -p "$DEPLOY_DIR"
    mkdir -p "$BACKUP_DIR"
    mkdir -p "$DEPLOY_DIR/logs"
}

# Backup current deployment
backup() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/${APP_NAME}_${timestamp}.tar.gz"

    log_info "Creating backup: $backup_file"
    tar -czf "$backup_file" -C "$(dirname $DEPLOY_DIR)" "$(basename $DEPLOY_DIR)"

    # Keep only last 10 backups
    ls -1 "$BACKUP_DIR"/*.tar.gz 2>/dev/null | tail -n +11 | xargs -r rm
    log_info "Backup complete: $backup_file"
}

# Pull latest images
pull_images() {
    log_info "Pulling latest Docker images..."
    docker-compose -f "$DEPLOY_DIR/$DOCKER_COMPOSE_FILE" pull
}

# Stop services
stop_services() {
    log_info "Stopping services..."
    docker-compose -f "$DEPLOY_DIR/$DOCKER_COMPOSE_FILE" down --remove-orphans || true
}

# Start services
start_services() {
    log_info "Starting services..."
    docker-compose -f "$DEPLOY_DIR/$DOCKER_COMPOSE_FILE" up -d
}

# Health check
health_check() {
    local max_attempts=30
    local attempt=1

    log_info "Running health check..."

    while [ $attempt -le $max_attempts ]; do
        if curl -sf "http://localhost:8080/health" > /dev/null 2>&1; then
            log_info "Health check passed!"
            return 0
        fi

        log_warn "Health check attempt $attempt/$max_attempts failed. Waiting..."
        sleep 2
        ((attempt++))
    done

    log_error "Health check failed after $max_attempts attempts"
    return 1
}

# Rollback
rollback() {
    local latest_backup=$(ls -1 "$BACKUP_DIR"/*.tar.gz 2>/dev/null | sort -r | head -n 1)

    if [ -z "$latest_backup" ]; then
        log_error "No backup found to rollback to"
        exit 1
    fi

    log_warn "Rolling back to: $latest_backup"

    stop_services
    rm -rf "$DEPLOY_DIR"
    tar -xzf "$latest_backup" -C "$(dirname $DEPLOY_DIR)"
    start_services
    health_check
}

# Deploy function
deploy() {
    local action="${1:-full}"

    log_info "Starting deployment (action: $action)..."

    check_root
    setup_dirs

    case "$action" in
        full)
            backup
            pull_images
            stop_services
            start_services
            health_check
            ;;
        update)
            pull_images
            stop_services
            start_services
            health_check
            ;;
        restart)
            stop_services
            start_services
            health_check
            ;;
        backup)
            backup
            ;;
        rollback)
            rollback
            ;;
        *)
            log_error "Unknown action: $action"
            echo "Usage: $0 {full|update|restart|backup|rollback}"
            exit 1
            ;;
    esac

    log_info "Deployment complete!"
}

# Main
main() {
    deploy "$1"
}

main "$@"
