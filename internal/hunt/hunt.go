package hunt

import ()

// Hunter is the 4-module hunt system: Assess → Compress → Search → Rank.
// Behavior source: embryo Hunt (17 modules → 4 focused).
type Hunter struct {
	MeshDir string // path to mesh slot directory
}

// NewHunter creates a hunter with the given mesh directory.
func NewHunter(meshDir string) *Hunter {
	return &Hunter{MeshDir: meshDir}
}
