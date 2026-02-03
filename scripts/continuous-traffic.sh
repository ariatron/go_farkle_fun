#!/bin/bash

# Continuous Traffic Generator for Farkle Observability
# Runs K6 load tests in the background to generate traces/metrics/logs

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PID_FILE="$PROJECT_ROOT/.continuous-traffic.pid"
LOG_FILE="$PROJECT_ROOT/continuous-traffic.log"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

check_dependencies() {
    if ! command -v k6 &> /dev/null; then
        print_error "k6 is not installed"
        echo "Install k6: https://k6.io/docs/getting-started/installation/"
        exit 1
    fi
}

start_traffic() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            print_warning "Continuous traffic is already running (PID: $PID)"
            echo "Use './scripts/continuous-traffic.sh stop' to stop it first"
            exit 1
        else
            # Stale PID file, remove it
            rm "$PID_FILE"
        fi
    fi

    print_status "Starting continuous traffic generator..."

    # Start K6 in the background
    nohup k6 run "$PROJECT_ROOT/tests/k6/continuous-traffic.js" > "$LOG_FILE" 2>&1 &
    K6_PID=$!

    echo "$K6_PID" > "$PID_FILE"

    sleep 2

    if ps -p "$K6_PID" > /dev/null 2>&1; then
        print_status "Continuous traffic started (PID: $K6_PID)"
        print_status "Log file: $LOG_FILE"
        echo ""
        echo "Commands:"
        echo "  Stop:   ./scripts/continuous-traffic.sh stop"
        echo "  Status: ./scripts/continuous-traffic.sh status"
        echo "  Logs:   tail -f $LOG_FILE"
    else
        print_error "Failed to start continuous traffic"
        rm -f "$PID_FILE"
        exit 1
    fi
}

stop_traffic() {
    if [ ! -f "$PID_FILE" ]; then
        print_warning "No continuous traffic running"
        exit 0
    fi

    PID=$(cat "$PID_FILE")

    if ps -p "$PID" > /dev/null 2>&1; then
        print_status "Stopping continuous traffic (PID: $PID)..."
        kill "$PID"

        # Wait for graceful shutdown
        sleep 2

        # Force kill if still running
        if ps -p "$PID" > /dev/null 2>&1; then
            kill -9 "$PID" 2>/dev/null || true
        fi

        rm "$PID_FILE"
        print_status "Continuous traffic stopped"
    else
        print_warning "Process not running, cleaning up PID file"
        rm "$PID_FILE"
    fi
}

status_traffic() {
    if [ ! -f "$PID_FILE" ]; then
        echo "Status: Not running"
        exit 0
    fi

    PID=$(cat "$PID_FILE")

    if ps -p "$PID" > /dev/null 2>&1; then
        echo "Status: Running (PID: $PID)"
        echo "Log file: $LOG_FILE"
        echo ""
        echo "Recent activity:"
        tail -n 5 "$LOG_FILE" 2>/dev/null || echo "No logs yet"
    else
        echo "Status: Not running (stale PID file)"
        rm "$PID_FILE"
    fi
}

restart_traffic() {
    print_status "Restarting continuous traffic..."
    stop_traffic
    sleep 1
    start_traffic
}

show_help() {
    cat << EOF
Continuous Traffic Generator for Farkle Observability

Usage: ./scripts/continuous-traffic.sh [COMMAND]

Commands:
    start       Start continuous traffic generation
    stop        Stop continuous traffic generation
    restart     Restart continuous traffic generation
    status      Show current status
    logs        Tail the log file
    help        Show this help message

Examples:
    ./scripts/continuous-traffic.sh start
    ./scripts/continuous-traffic.sh status
    tail -f $LOG_FILE

This script generates continuous background traffic to your Farkle game,
creating traces, metrics, and logs for Grafana Cloud observability.

EOF
}

# Main script logic
case "${1:-help}" in
    start)
        check_dependencies
        start_traffic
        ;;
    stop)
        stop_traffic
        ;;
    restart)
        check_dependencies
        restart_traffic
        ;;
    status)
        status_traffic
        ;;
    logs)
        if [ -f "$LOG_FILE" ]; then
            tail -f "$LOG_FILE"
        else
            print_warning "No log file found at $LOG_FILE"
        fi
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
