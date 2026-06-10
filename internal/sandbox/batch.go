package sandbox

import (
	"fmt"
	"sync"
)

// BatchResult holds the outcome of a batch provisioning operation.
type BatchResult struct {
	Worktrees map[string]string // modelID → worktreePath
	Errors    map[string]error
}

// ProvisionBatch creates worktrees for multiple models in parallel.
// Each model gets an isolated worktree from the same baseline SHA.
func (wm *WorktreeManager) ProvisionBatch(modelIDs []string) *BatchResult {
	result := &BatchResult{
		Worktrees: make(map[string]string),
		Errors:    make(map[string]error),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, modelID := range modelIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			wtPath, err := wm.Provision(id)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.Errors[id] = err
			} else {
				result.Worktrees[id] = wtPath
			}
		}(modelID)
	}

	wg.Wait()
	return result
}

// DestroyBatch removes all worktrees for the given models.
func (wm *WorktreeManager) DestroyBatch(modelIDs []string) map[string]error {
	errors := make(map[string]error)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, modelID := range modelIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			err := wm.Destroy(id)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors[id] = err
			}
		}(modelID)
	}

	wg.Wait()
	return errors
}

// String returns a human-readable batch summary.
func (r *BatchResult) String() string {
	return fmt.Sprintf("provisioned=%d errors=%d", len(r.Worktrees), len(r.Errors))
}
