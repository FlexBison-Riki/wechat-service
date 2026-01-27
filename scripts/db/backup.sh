#!/bin/bash
set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-wechat_service}"
BACKUP_DIR="${BACKUP_DIR:-/opt/wechat-service/backups/db}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Generate filename with timestamp
generate_filename() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    echo "${DB_NAME}_${timestamp}.sql.gz"
}

# Backup database
backup() {
    local filename=$(generate_filename)
    local filepath="$BACKUP_DIR/$filename"

    log_info "Starting database backup..."
    log_info "Database: $DB_NAME"
    log_info "Output: $filepath"

    # Perform backup with compression
    PGPASSWORD="$DB_PASSWORD" pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        --no-owner --no-privileges \
        | gzip > "$filepath"

    # Get file size
    local size=$(du -h "$filepath" | cut -f1)
    log_info "Backup complete! Size: $size"

    # Verify backup
    if gunzip -t "$filepath" 2>/dev/null; then
        log_info "Backup verification passed"
    else
        log_warn "Backup verification failed!"
        rm -f "$filepath"
        exit 1
    fi

    echo "$filepath"
}

# Restore database
restore() {
    local backup_file="$1"

    if [ -z "$backup_file" ]; then
        log_warn "Please specify a backup file to restore"
        exit 1
    fi

    if [ ! -f "$backup_file" ]; then
        log_warn "Backup file not found: $backup_file"
        exit 1
    fi

    log_info "Restoring database from: $backup_file"
    log_warn "This will overwrite all data in $DB_NAME"

    read -p "Are you sure? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        log_info "Restore cancelled"
        exit 0
    fi

    # Drop and recreate database
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "DROP DATABASE IF EXISTS ${DB_NAME}_restore;" 2>/dev/null || true
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE ${DB_NAME}_restore;" 2>/dev/null || true

    # Restore
    gunzip -c "$backup_file" | PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "${DB_NAME}_restore"

    log_info "Database restored to ${DB_NAME}_restore"
    log_info "Run this command to rename:"
    echo "  PGPASSWORD='$DB_PASSWORD' psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c \"ALTER DATABASE ${DB_NAME}_rename RENAME TO ${DB_NAME};\""
}

# List backups
list_backups() {
    echo ""
    echo "Available backups in $BACKUP_DIR:"
    echo "================================"

    ls -lah "$BACKUP_DIR"/*.sql.gz 2>/dev/null || echo "No backups found"

    echo ""
    echo "Total: $(ls -1 "$BACKUP_DIR"/*.sql.gz 2>/dev/null | wc -l) files"
    echo "Oldest: $(ls -1t "$BACKUP_DIR"/*.sql.gz 2>/dev/null | tail -n 1)"
    echo "Newest: $(ls -1t "$BACKUP_DIR"/*.sql.gz 2>/dev/null | head -n 1)"
}

# Cleanup old backups
cleanup() {
    log_info "Cleaning up backups older than $RETENTION_DAYS days..."

    local deleted=$(find "$BACKUP_DIR" -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete -printf '%f\n' 2>/dev/null | wc -l)

    if [ "$deleted" -gt 0 ]; then
        log_info "Deleted $deleted old backup(s)"
    else
        log_info "No old backups to delete"
    fi
}

# Main
main() {
    case "${1:-backup}" in
        backup)
            backup
            cleanup
            ;;
        restore)
            restore "$2"
            ;;
        list)
            list_backups
            ;;
        cleanup)
            cleanup
            ;;
        *)
            echo "Usage: $0 {backup|restore|list|cleanup}"
            echo ""
            echo "Environment variables:"
            echo "  DB_HOST          Database host (default: localhost)"
            echo "  DB_PORT          Database port (default: 5432)"
            echo "  DB_USER          Database user (default: postgres)"
            echo "  DB_PASSWORD      Database password"
            echo "  DB_NAME          Database name (default: wechat_service)"
            echo "  BACKUP_DIR       Backup directory"
            echo "  RETENTION_DAYS   Days to keep backups (default: 7)"
            exit 1
            ;;
    esac
}

main "$@"
