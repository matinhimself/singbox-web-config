package watcher

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches for configuration file changes
type Watcher struct {
	configPath string
	watcher    *fsnotify.Watcher
	onChange   func()
	stopCh     chan struct{}
}

// NewWatcher creates a new file watcher
func NewWatcher(configPath string, onChange func()) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		configPath: configPath,
		watcher:    fw,
		onChange:   onChange,
		stopCh:     make(chan struct{}),
	}

	// Watch the directory containing the config file
	// (watching the file directly doesn't work well with editors that replace files)
	dir := filepath.Dir(configPath)
	if err := fw.Add(dir); err != nil {
		fw.Close()
		return nil, fmt.Errorf("failed to watch directory: %w", err)
	}

	return w, nil
}

// Start starts watching for file changes
func (w *Watcher) Start() {
	go w.watch()
}

// Stop stops watching for file changes
func (w *Watcher) Stop() {
	close(w.stopCh)
	w.watcher.Close()
}

// watch monitors file events
func (w *Watcher) watch() {
	// Debounce rapid fire events
	var timer *time.Timer

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only trigger on events for our config file
			if filepath.Clean(event.Name) == filepath.Clean(w.configPath) {
				// Check if it's a write or create event
				if event.Op&fsnotify.Write == fsnotify.Write ||
				   event.Op&fsnotify.Create == fsnotify.Create {
					// Debounce: reset timer on each event
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(500*time.Millisecond, func() {
						log.Printf("Config file changed: %s", w.configPath)
						if w.onChange != nil {
							w.onChange()
						}
					})
				}
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)

		case <-w.stopCh:
			return
		}
	}
}
