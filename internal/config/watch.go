package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// Watcher monitors repo config and reloads it on change.
// It also watches persona/primer/waiver files if configured.
type Watcher struct {
	repoPath string
	loader   func(string) (*RepoConfig, error)
	mu       sync.RWMutex
	cfg      *RepoConfig
	onUpdate func(*RepoConfig)
	watcher  *fsnotify.Watcher
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewWatcher creates a config watcher for the given repo.
// The onUpdate callback is invoked whenever config or watched files change.
func NewWatcher(repoPath string, onUpdate func(*RepoConfig)) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		repoPath: repoPath,
		loader:   LoadRepoConfig,
		onUpdate: onUpdate,
		watcher:  fsWatcher,
		stopCh:   make(chan struct{}),
	}

	cfg, err := w.loader(repoPath)
	if err != nil {
		fsWatcher.Close()
		return nil, err
	}
	w.cfg = cfg

	if err := w.addWatches(cfg); err != nil {
		fsWatcher.Close()
		return nil, err
	}

	w.wg.Add(1)
	go w.loop()
	return w, nil
}

// Config returns the current config snapshot.
func (w *Watcher) Config() *RepoConfig {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.cfg
}

// Stop shuts down the watcher.
func (w *Watcher) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	_ = w.watcher.Close()
}

func (w *Watcher) addWatches(cfg *RepoConfig) error {
	repoName := filepath.Base(w.repoPath)
	configDir := filepath.Join(w.repoPath, ".mimic", repoName)
	configFile := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configFile); err == nil {
		if err := w.watcher.Add(configFile); err != nil {
			return err
		}
	}
	// Watch the directory so new files are noticed.
	if _, err := os.Stat(configDir); err == nil {
		if err := w.watcher.Add(configDir); err != nil {
			return err
		}
	}

	for _, path := range cfg.Personas {
		_ = w.watcher.Add(filepath.Join(w.repoPath, path))
	}
	for _, path := range cfg.Primers {
		_ = w.watcher.Add(filepath.Join(w.repoPath, path))
	}
	for _, path := range cfg.Waivers {
		_ = w.watcher.Add(filepath.Join(w.repoPath, path))
	}
	return nil
}

func (w *Watcher) loop() {
	defer w.wg.Done()
	for {
		select {
		case <-w.stopCh:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				w.reload()
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("config watcher error: %v", err)
		}
	}
}

func (w *Watcher) reload() {
	cfg, err := w.loader(w.repoPath)
	if err != nil {
		log.Printf("config reload failed: %v", err)
		return
	}

	w.mu.Lock()
	w.cfg = cfg
	w.mu.Unlock()

	if w.onUpdate != nil {
		go w.onUpdate(cfg)
	}
}
