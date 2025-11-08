package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"rwmod-monitor/internal/archiver"
	"rwmod-monitor/internal/config"
	"rwmod-monitor/internal/queue"
	"rwmod-monitor/internal/tracker"
	"rwmod-monitor/internal/uploader"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rwmod-monitor",
	Short: "A file system monitor tool",
	Long:  `RWMod Monitor is a CLI tool that monitors file system changes and backs them up to S3.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		if err := cfg.Validate(); err != nil {
			configPath, _ := config.GetConfigPath()
			log.Printf("\nConfiguration incomplete!")
			log.Printf("Please edit the config file: %s\n", configPath)
			log.Fatalf("Error: %v", err)
		}

		if _, err := os.Stat(cfg.MonitorDir); os.IsNotExist(err) {
			log.Fatalf("Monitor directory does not exist: %s", cfg.MonitorDir)
		}

		startMonitoring(cfg)
	},
}

func init() {
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type eventHandler struct {
	tracker *tracker.Tracker
	watcher *fsnotify.Watcher
}

func (h *eventHandler) handleEvents() {
	for {
		select {
		case event, ok := <-h.watcher.Events:
			if !ok {
				return
			}
			h.processEvent(event)
		case err, ok := <-h.watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (h *eventHandler) processEvent(event fsnotify.Event) {
	log.Println("event:", event)

	if event.Op&fsnotify.Write == fsnotify.Write {
		log.Println("modified file:", event.Name)
		h.tracker.OnFileChange(event.Name)
	}

	if event.Op&fsnotify.Create == fsnotify.Create {
		h.handleCreate(event.Name)
	}

	if event.Op&fsnotify.Remove == fsnotify.Remove {
		log.Println("deleted:", event.Name)
	}

	if event.Op&fsnotify.Rename == fsnotify.Rename {
		log.Println("renamed:", event.Name)
	}
}

func (h *eventHandler) handleCreate(path string) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			log.Println("created directory:", path)
		} else {
			log.Println("created file:", path)
			h.tracker.OnFileChange(path)
		}
	} else {
		log.Println("created:", path)
	}
}

func addDirectoryToWatcher(watcher *fsnotify.Watcher, dir string) error {
	err := watcher.Add(dir)
	if err != nil {
		return err
	}
	log.Printf("Watching: %s", dir)
	return nil
}

func getFirstLevelDirs(monitorDir string) ([]string, error) {
	entries, err := os.ReadDir(monitorDir)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && !isHiddenDir(entry.Name()) {
			dirPath := filepath.Join(monitorDir, entry.Name())
			dirs = append(dirs, dirPath)
		}
	}
	return dirs, nil
}

func isHiddenDir(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

func startMonitoring(cfg *config.Config) {
	s3Client, err := uploader.NewS3Uploader(
		cfg.S3.Endpoint,
		cfg.S3.AccessKey,
		cfg.S3.SecretKey,
		cfg.S3.Bucket,
	)
	if err != nil {
		log.Fatalf("Failed to create S3 uploader: %v", err)
	}

	uploadQueue, err := queue.NewQueue(s3Client, cfg.MaxRetries, cfg.MonitorDir)
	if err != nil {
		log.Fatalf("Failed to create upload queue: %v", err)
	}
	defer uploadQueue.Close()

	arch := archiver.NewArchiver(cfg.MonitorDir)

	fileTracker := tracker.NewTracker(
		cfg.MonitorDir,
		cfg.DelayMinutes,
		arch.Archive,
		uploadQueue.Enqueue,
	)
	defer fileTracker.Stop()

	if err := performInitialBackup(cfg, arch, uploadQueue); err != nil {
		log.Printf("Warning: initial backup failed: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	firstLevelDirs, err := getFirstLevelDirs(cfg.MonitorDir)
	if err != nil {
		log.Fatalf("Failed to get first level directories: %v", err)
	}

	for _, dir := range firstLevelDirs {
		if err := addDirectoryToWatcher(watcher, dir); err != nil {
			log.Printf("Warning: failed to watch %s: %v", dir, err)
		}
	}

	handler := &eventHandler{
		tracker: fileTracker,
		watcher: watcher,
	}

	go handler.handleEvents()

	log.Println("Monitoring started. Press Ctrl+C to stop...")

	done := make(chan bool)
	<-done
}

func performInitialBackup(cfg *config.Config, arch *archiver.Archiver, uploadQueue *queue.Queue) error {
	dirs, err := getFirstLevelDirs(cfg.MonitorDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		log.Printf("Creating initial backup for: %s", dir)
		archivePath, err := arch.Archive(dir)
		if err != nil {
			log.Printf("Failed to archive %s: %v", dir, err)
			continue
		}
		log.Printf("Created initial archive: %s", archivePath)
		uploadQueue.Enqueue(archivePath)
	}

	return nil
}
