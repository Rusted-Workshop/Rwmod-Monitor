package tracker

import (
	"log"
	"path/filepath"
	"sync"
	"time"
)

type ArchiveFunc func(string) (string, error)
type EnqueueFunc func(string)

type Tracker struct {
	monitorDir   string
	delayMinutes int
	timers       map[string]*time.Timer
	mutex        sync.Mutex
	archiveFunc  ArchiveFunc
	enqueueFunc  EnqueueFunc
}

func NewTracker(monitorDir string, delayMinutes int, archiveFunc ArchiveFunc, enqueueFunc EnqueueFunc) *Tracker {
	return &Tracker{
		monitorDir:   monitorDir,
		delayMinutes: delayMinutes,
		timers:       make(map[string]*time.Timer),
		archiveFunc:  archiveFunc,
		enqueueFunc:  enqueueFunc,
	}
}

func (t *Tracker) OnFileChange(filePath string) {
	targetDir := t.getTargetDir(filePath)
	if targetDir == "" {
		return
	}

	t.resetTimer(targetDir)
}

func (t *Tracker) getTargetDir(filePath string) string {
	relPath, err := filepath.Rel(t.monitorDir, filePath)
	if err != nil {
		return ""
	}

	parts := filepath.SplitList(relPath)
	if len(parts) == 0 {
		parts = []string{filepath.Dir(relPath)}
	}

	if parts[0] == "." || parts[0] == ".." {
		return ""
	}

	return filepath.Join(t.monitorDir, parts[0])
}

func (t *Tracker) resetTimer(targetDir string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if timer, exists := t.timers[targetDir]; exists {
		timer.Stop()
	}

	t.timers[targetDir] = time.AfterFunc(
		time.Duration(t.delayMinutes)*time.Minute,
		func() {
			t.handleTimeout(targetDir)
		},
	)

	log.Printf("Reset timer for directory: %s", targetDir)
}

func (t *Tracker) handleTimeout(targetDir string) {
	log.Printf("Archiving directory: %s", targetDir)

	archivePath, err := t.archiveFunc(targetDir)
	if err != nil {
		log.Printf("Failed to archive %s: %v", targetDir, err)
		return
	}

	log.Printf("Created archive: %s", archivePath)
	t.enqueueFunc(archivePath)

	t.mutex.Lock()
	delete(t.timers, targetDir)
	t.mutex.Unlock()
}

func (t *Tracker) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, timer := range t.timers {
		timer.Stop()
	}
}
