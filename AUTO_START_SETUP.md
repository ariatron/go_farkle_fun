# Farkle Auto-Start Setup

Your Farkle observability stack is now configured to **automatically start at login** using macOS launchd.

## What's Auto-Starting

When you log in to your Mac, these services will automatically start:

1. **✅ Grafana Alloy** - Collects metrics, traces, and logs to send to Grafana Cloud
2. **✅ Farkle Server** - Runs on http://localhost:8080 in single-player mode
3. **❌ K6 Load Tests** - Disabled by default (optional, see below)

## Service Management

Use the provided management script to control services:

```bash
# Show status of all services
./scripts/manage-services.sh status

# Start services manually (if stopped)
./scripts/manage-services.sh start

# Stop all services
./scripts/manage-services.sh stop

# Restart all services
./scripts/manage-services.sh restart

# Enable K6 load tests (optional)
./scripts/manage-services.sh enable-k6

# Disable K6 load tests
./scripts/manage-services.sh disable-k6
```

## LaunchAgent Configuration Files

The following files in `~/Library/LaunchAgents/` control auto-start:

- `com.farkle.alloy.plist` - Grafana Alloy
- `com.farkle.server.plist` - Farkle Server
- `com.farkle.k6.plist` - K6 Load Tests (disabled by default)

## Manual launchctl Commands

If you need to use launchctl directly:

```bash
# Load/start services
launchctl load ~/Library/LaunchAgents/com.farkle.alloy.plist
launchctl load ~/Library/LaunchAgents/com.farkle.server.plist

# Unload/stop services
launchctl unload ~/Library/LaunchAgents/com.farkle.alloy.plist
launchctl unload ~/Library/LaunchAgents/com.farkle.server.plist

# List running services
launchctl list | grep farkle
```

## Logs

All services log to `/tmp/`:

- **Alloy**: `/tmp/alloy.log` and `/tmp/alloy.error.log`
- **Server**: `/tmp/farkle-server.log` and `/tmp/farkle-server.error.log`
- **K6**: `/tmp/k6-load-test.log` and `/tmp/k6-load-test.error.log`

View logs in real-time:

```bash
tail -f /tmp/alloy.log
tail -f /tmp/farkle-server.log
tail -f /tmp/k6-load-test.log
```

## Access Points

After login, your services will be available at:

- **Game UI**: http://localhost:8080
- **Alloy UI**: http://localhost:12345
- **Metrics**: http://localhost:8080/metrics
- **Health**: http://localhost:8080/health

## Configuration

### Farkle Server Settings

Edit `~/Library/LaunchAgents/com.farkle.server.plist` to change:

- **Game mode**: Change `GAME_MODE` from `single` to `multi` for multiplayer
- **Environment variables**: Add/modify under `<key>EnvironmentVariables</key>`

After editing, reload:

```bash
launchctl unload ~/Library/LaunchAgents/com.farkle.server.plist
launchctl load ~/Library/LaunchAgents/com.farkle.server.plist
```

### Alloy Configuration

Alloy configuration is in `/Users/ariatron/farkle-multiplayer/alloy-config.alloy`

After editing alloy-config.alloy, restart Alloy:

```bash
launchctl unload ~/Library/LaunchAgents/com.farkle.alloy.plist
launchctl load ~/Library/LaunchAgents/com.farkle.alloy.plist
```

## Troubleshooting

### Services won't start

Check error logs:

```bash
cat /tmp/alloy.error.log
cat /tmp/farkle-server.error.log
```

### Services not auto-starting after login

Verify plist files are loaded:

```bash
launchctl list | grep farkle
```

If not listed, manually load them:

```bash
launchctl load ~/Library/LaunchAgents/com.farkle.alloy.plist
launchctl load ~/Library/LaunchAgents/com.farkle.server.plist
```

### Port already in use

If port 8080 or 4318 is already in use:

```bash
# Find what's using port 8080
lsof -ti :8080

# Kill the process (if needed)
lsof -ti :8080 | xargs kill
```

### Disable auto-start

To prevent services from starting at login:

```bash
launchctl unload ~/Library/LaunchAgents/com.farkle.alloy.plist
launchctl unload ~/Library/LaunchAgents/com.farkle.server.plist
```

To re-enable:

```bash
launchctl load ~/Library/LaunchAgents/com.farkle.alloy.plist
launchctl load ~/Library/LaunchAgents/com.farkle.server.plist
```

## K6 Load Tests (Optional)

K6 load tests are **disabled by default** to avoid continuous CPU usage. Enable only when you want continuous traffic for testing:

```bash
# Enable K6 load tests
./scripts/manage-services.sh enable-k6

# Disable K6 load tests
./scripts/manage-services.sh disable-k6
```

When enabled, K6 will run the `single-player-traffic.js` test continuously (30-minute duration, loops).

## Performance Considerations

- **Alloy**: Lightweight (~50-100 MB RAM)
- **Farkle Server**: Minimal (~50 MB RAM)
- **K6 Load Tests**: Higher CPU usage when enabled (~5-10% CPU)

If you notice performance issues, consider:
1. Disabling K6 load tests
2. Adjusting Alloy scrape intervals in `alloy-config.alloy`
3. Stopping services when not needed: `./scripts/manage-services.sh stop`

## Next Boot

After your next restart, everything will automatically start when you log in. No manual intervention needed! 🎉

You can verify this by:
1. Restarting your Mac
2. Logging in
3. Waiting ~5 seconds
4. Running: `./scripts/manage-services.sh status`
