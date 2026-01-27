#!/bin/bash
set -e

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
TIMEOUT="${TIMEOUT:-5}"
MAX_RETRIES="${MAX_RETRIES:-3}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check if service is alive
check_alive() {
    local attempt=1

    while [ $attempt -le $MAX_RETRIES ]; do
        if curl -sf --max-time "$TIMEOUT" "$API_URL/health" > /dev/null 2>&1; then
            return 0
        fi
        log_warn "Health check attempt $attempt/$MAX_RETRIES failed"
        ((attempt++))
        sleep 1
    done

    return 1
}

# Check response time
check_response_time() {
    local start_time=$(date +%s%N)
    curl -sf --max-time "$TIMEOUT" -o /dev/null "$API_URL/health"
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))

    echo "$duration"
}

# Check dependencies
check_dependencies() {
    local status=0

    log_info "Checking dependencies..."

    # Check Redis
    if curl -sf --max-time 2 "redis-cli ping" > /dev/null 2>&1 || \
       docker exec $(docker ps -qf "name=redis") redis-cli ping > /dev/null 2>&1; then
        log_info "✓ Redis is healthy"
    else
        log_warn "✗ Redis connection failed"
        status=1
    fi

    # Check PostgreSQL
    if PGPASSWORD="${DB_PASSWORD:-postgres}" psql -h "${DB_HOST:-localhost}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-wechat_service}" -c "SELECT 1;" > /dev/null 2>&1; then
        log_info "✓ PostgreSQL is healthy"
    else
        log_warn "✗ PostgreSQL connection failed"
        status=1
    fi

    return $status
}

# Generate status report
generate_report() {
    local alive=true
    local response_time="N/A"

    log_info "Generating health report for: $API_URL"

    if check_alive; then
        response_time=$(check_response_time)
        log_info "✓ Service is alive (response time: ${response_time}ms)"
    else
        log_error "✗ Service is not responding"
        alive=false
    fi

    echo ""
    echo "================================"
    echo "         Health Report          "
    echo "================================"
    echo "Service:       $API_URL"
    echo "Status:        $([ "$alive" = true ] && echo "HEALTHY" || echo "UNHEALTHY")"
    echo "Response Time: ${response_time}ms"
    echo "Timestamp:     $(date '+%Y-%m-%d %H:%M:%S')"
    echo "================================"

    check_dependencies
}

# Main
main() {
    case "${1:-report}" in
        alive)
            check_alive
            ;;
        report)
            generate_report
            ;;
        dependencies)
            check_dependencies
            ;;
        *)
            echo "Usage: $0 {alive|report|dependencies}"
            exit 1
            ;;
    esac
}

main "$@"
