# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RWMod Monitor is a CLI tool that monitors filesystem changes in first-level subdirectories and automatically backs them up to S3-compatible storage. It uses a delayed trigger mechanism (5 minutes default) to avoid creating backups during active file modifications.

## Build and Run Commands

```bash
# Build the project
go build -o rwmod-monitor

# Build on Windows (using build.bat)
build.bat

# Build on Unix/Linux (using build.sh)
./build.sh

# Run the application
./rwmod-monitor

# Get dependencies
go get
```

## Architecture Overview

### Component Flow

1. **cmd/root.go** - Entry point and orchestration
   - Initializes all components with dependency injection
   - Sets up fsnotify watcher for first-level subdirectories only
   - Performs initial backup of all existing subdirectories on startup
   - Event handler routes filesystem events to the tracker

2. **internal/tracker/tracker.go** - Delayed trigger system
   - Manages per-directory timers (default: 5 minute delay)
   - Resets timer on each file change within a directory
   - Triggers archive creation when timer expires
   - Uses callbacks (ArchiveFunc, EnqueueFunc) for dependency injection

3. **internal/archiver/archiver.go** - Archive creation
   - Creates `.rwmod` files (which are zip archives)
   - Naming format: `{dirname}-{unix_timestamp}.rwmod`
   - Preserves internal directory structure relative to the monitored subdirectory

4. **internal/queue/queue.go** - Upload queue with retry logic
   - Worker-based upload queue (buffered channel, size 100)
   - Retry mechanism: max 3 attempts by default
   - Failed uploads moved to `{monitor_dir}/.failed_uploads/`
   - Graceful shutdown with WaitGroup

5. **internal/uploader/s3.go** - S3 upload implementation
   - Uses AWS SDK v2 with custom endpoint support
   - Uploads archives using filename as S3 key
   - Supports any S3-compatible storage

6. **internal/config/config.go** - Configuration management
   - Auto-creates `config.json` template on first run (next to executable)
   - Validates configuration before starting monitoring
   - Config location: same directory as the executable binary

### Key Design Patterns

**Dependency Injection via Callbacks**: The tracker accepts `ArchiveFunc` and `EnqueueFunc` callbacks, allowing the main coordinator (cmd/root.go) to wire up components without tight coupling.

**Single-Level Directory Monitoring**: The system only watches first-level subdirectories of the configured `monitor_dir`. Hidden directories (starting with `.`) are automatically ignored. This is enforced in:
- cmd/root.go:120-134 (`getFirstLevelDirs` function)
- cmd/root.go:136-138 (`isHiddenDir` function)

**Timer-Based Batching**: Each monitored subdirectory has its own timer in the tracker. File changes reset the timer, preventing backups during active editing sessions. Timers are thread-safe using mutex protection.

## Configuration

The application expects `config.json` in the same directory as the executable. The config structure:

```go
type Config struct {
    MonitorDir    string   // Root directory to monitor
    DelayMinutes  int      // Delay before creating backup (default: 5)
    MaxRetries    int      // Upload retry attempts (default: 3)
    S3            S3Config // S3 configuration
}
```

Config validation ensures placeholder values (e.g., "your-access-key-id") are replaced before the application runs.

## Testing Considerations

When adding tests:
- Mock the fsnotify watcher for unit testing event handlers
- Use the callback pattern in tracker for testing without real archiving
- Test timer behavior with shorter durations to avoid slow tests
- Mock S3 uploader using the `queue.Uploader` interface

## Common Modification Scenarios

**Changing the backup delay**: Modify `config.DelayMinutes` and restart the application.

**Supporting nested directory monitoring**: Currently limited to first-level subdirectories. To enable recursive monitoring, modify `getFirstLevelDirs` in cmd/root.go and update `tracker.getTargetDir` logic to handle nested paths.

**Different archive formats**: Modify `internal/archiver/archiver.go` to use different compression. The `.rwmod` extension is used for compatibility but the format is standard ZIP.

**Custom S3 key naming**: Modify `internal/uploader/s3.go:53` where `fileName` is set as the S3 key.
