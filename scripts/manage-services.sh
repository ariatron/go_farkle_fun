#!/bin/bash
# Farkle Observability Services Management Script
# Manages launchd services for Alloy, Farkle Server, and K6 Load Tests

PLIST_DIR="$HOME/Library/LaunchAgents"
ALLOY_PLIST="$PLIST_DIR/com.farkle.alloy.plist"
SERVER_PLIST="$PLIST_DIR/com.farkle.server.plist"
K6_PLIST="$PLIST_DIR/com.farkle.k6.plist"

show_status() {
    echo "=== Farkle Services Status ==="
    echo ""

    echo "Grafana Alloy:"
    if launchctl list | grep -q "com.farkle.alloy"; then
        echo "  ✅ Running (PID: $(launchctl list | grep com.farkle.alloy | awk '{print $1}'))"
    else
        echo "  ❌ Not running"
    fi

    echo ""
    echo "Farkle Server:"
    if launchctl list | grep -q "com.farkle.server"; then
        echo "  ✅ Running (PID: $(launchctl list | grep com.farkle.server | awk '{print $1}'))"
        curl -s http://localhost:8080/health > /dev/null && echo "  ✅ Health check: OK" || echo "  ⚠️  Health check: Failed"
    else
        echo "  ❌ Not running"
    fi

    echo ""
    echo "K6 Load Tests:"
    if launchctl list | grep -q "com.farkle.k6"; then
        echo "  ✅ Running (PID: $(launchctl list | grep com.farkle.k6 | awk '{print $1}'))"
    else
        echo "  ❌ Not running (or disabled)"
    fi

    echo ""
    echo "Access Points:"
    echo "  • Game UI: http://localhost:8080"
    echo "  • Alloy UI: http://localhost:12345"
    echo "  • Metrics: http://localhost:8080/metrics"
    echo ""
    echo "Logs:"
    echo "  • Alloy: tail -f /tmp/alloy.log"
    echo "  • Server: tail -f /tmp/farkle-server.log"
    echo "  • K6: tail -f /tmp/k6-load-test.log"
}

start_all() {
    echo "Starting all services..."
    launchctl bootstrap gui/$(id -u) "$ALLOY_PLIST" 2>/dev/null && echo "✅ Alloy started"
    launchctl bootstrap gui/$(id -u) "$SERVER_PLIST" 2>/dev/null && echo "✅ Server started"
    echo "⏳ Waiting for services to initialize..."
    sleep 3
    show_status
}

stop_all() {
    echo "Stopping all services..."
    launchctl bootout gui/$(id -u)/com.farkle.alloy 2>/dev/null && echo "✅ Alloy stopped"
    launchctl bootout gui/$(id -u)/com.farkle.server 2>/dev/null && echo "✅ Server stopped"
    launchctl bootout gui/$(id -u)/com.farkle.k6 2>/dev/null && echo "✅ K6 stopped"
}

restart_all() {
    echo "Restarting all services..."
    stop_all
    sleep 2
    start_all
}

enable_k6() {
    echo "Enabling K6 load tests..."
    launchctl bootstrap gui/$(id -u) "$K6_PLIST" 2>/dev/null && echo "✅ K6 load tests enabled and started"
}

disable_k6() {
    echo "Disabling K6 load tests..."
    launchctl bootout gui/$(id -u)/com.farkle.k6 2>/dev/null && echo "✅ K6 load tests stopped and disabled"
}

case "$1" in
    status)
        show_status
        ;;
    start)
        start_all
        ;;
    stop)
        stop_all
        ;;
    restart)
        restart_all
        ;;
    enable-k6)
        enable_k6
        ;;
    disable-k6)
        disable_k6
        ;;
    *)
        echo "Farkle Observability Services Manager"
        echo ""
        echo "Usage: $0 {status|start|stop|restart|enable-k6|disable-k6}"
        echo ""
        echo "Commands:"
        echo "  status      - Show status of all services"
        echo "  start       - Start Alloy and Farkle Server"
        echo "  stop        - Stop all services"
        echo "  restart     - Restart all services"
        echo "  enable-k6   - Enable and start K6 load tests"
        echo "  disable-k6  - Disable and stop K6 load tests"
        echo ""
        echo "Services run automatically at login after initial setup."
        exit 1
        ;;
esac
