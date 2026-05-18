package orchestrator

import (
	"strings"
	"testing"
	"time"
)

// Benchmark test suite for task decomposition
// Goal: 100% success rate for 5 typical scenarios

func TestBenchmark_Decomposition(t *testing.T) {
	scenarios := []struct {
		name              string
		intent            string
		projectCtx        *ProjectContext
		expectedTasks     int
		expectedGroups    int
		maxPlanTimeMicros int64
	}{
		{
			name:              "Simple",
			intent:            "check if file exists",
			projectCtx:        &ProjectContext{RootPath: "/tmp", SourceFiles: []string{"main.go"}},
			expectedTasks:     1,
			expectedGroups:    1,
			maxPlanTimeMicros: 100,
		},
		{
			name:              "Build",
			intent:            "build and test entire project",
			projectCtx:        &ProjectContext{RootPath: "/tmp/proj", SourceFiles: make([]string, 200), BuildSystem: "make"},
			expectedTasks:     3,
			expectedGroups:    3,
			maxPlanTimeMicros: 100,
		},
		{
			name:              "Git",
			intent:            "commit all changes safely",
			projectCtx:        &ProjectContext{RootPath: "/tmp/repo", SourceFiles: make([]string, 50), GitStatus: "dirty"},
			expectedTasks:     3,
			expectedGroups:    3,
			maxPlanTimeMicros: 100,
		},
		{
			name:              "Refactor",
			intent:            "refactor codebase and run tests",
			projectCtx:        &ProjectContext{RootPath: "/tmp/proj", SourceFiles: make([]string, 500), BuildSystem: "go"},
			expectedTasks:     4,
			expectedGroups:    4,
			maxPlanTimeMicros: 100,
		},
		{
			name:              "Analysis",
			intent:            "analyze all dependencies",
			projectCtx:        &ProjectContext{RootPath: "/tmp/proj", SourceFiles: make([]string, 300)},
			expectedTasks:     2,
			expectedGroups:    2,
			maxPlanTimeMicros: 100,
		},
	}

	d := NewDecomposer()

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			start := time.Now()
			graph, err := d.Decompose(sc.intent, sc.projectCtx)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("decomposition failed: %v", err)
			}

			// Check task count
			if len(graph.Tasks) != sc.expectedTasks {
				t.Errorf("expected %d tasks, got %d", sc.expectedTasks, len(graph.Tasks))
			}

			// Check parallel groups
			if len(graph.ParallelGroups) != sc.expectedGroups {
				t.Errorf("expected %d parallel groups, got %d", sc.expectedGroups, len(graph.ParallelGroups))
			}

			// Check all tasks are in some group
			taskInGroup := make(map[string]bool)
			for _, group := range graph.ParallelGroups {
				for _, tid := range group {
					taskInGroup[tid] = true
				}
			}
			for tid := range graph.Tasks {
				if !taskInGroup[tid] {
					t.Errorf("task %s not in any parallel group", tid)
				}
			}

			// Check planning time
			if elapsed.Microseconds() > sc.maxPlanTimeMicros {
				t.Errorf("planning time %dμs > limit %dμs", elapsed.Microseconds(), sc.maxPlanTimeMicros)
			}

			// Log metrics
			t.Logf("tasks=%d, groups=%d, time=%dμs, complexity=%.2f",
				len(graph.Tasks), len(graph.ParallelGroups), elapsed.Microseconds(),
				graph.Tasks["t-1"].Complexity)
		})
	}
}

func TestBenchmark_ContextCompression(t *testing.T) {
	c := NewContextCompressor(64000)

	// Large project context
	ctx := &ProjectContext{
		RootPath:    "/home/user/very-large-project",
		Languages:   []string{"Go", "Rust", "Python", "TypeScript"},
		Frameworks:  []string{"React", "Gin", "Tokio"},
		EntryPoints: []string{"cmd/api/main.go", "cmd/worker/main.go", "frontend/src/index.tsx"},
		ConfigFiles: []string{"go.mod", "package.json", "Cargo.toml", "docker-compose.yml", "Makefile"},
		SourceFiles: func() []string {
			files := make([]string, 1500)
			for i := range files {
				files[i] = "src/file.go"
			}
			return files
		}(),
		TestFiles:   make([]string, 200),
		BuildSystem: "make",
	}

	start := time.Now()
	compressed := c.Compress(ctx, "build and deploy entire project with tests")
	elapsed := time.Since(start)

	// Compressed should be < 500 characters
	if len(compressed) > 500 {
		t.Errorf("compressed too long: %d chars (max 500)", len(compressed))
	}

	// Should contain essential info
	if !containsAny(compressed, []string{"Go", "Rust", "Python"}) {
		t.Error("missing language info in compressed context")
	}
	if !containsAny(compressed, []string{"1500", "source"}) {
		t.Error("missing file count in compressed context")
	}

	t.Logf("compressed_length=%d, time=%dμs", len(compressed), elapsed.Microseconds())
	t.Logf("compressed:\n%s", compressed)
}

func TestBenchmark_ComplexityClassification(t *testing.T) {
	tests := []struct {
		intent            string
		projectSize       int
		wantDecomposed    bool
		wantMinComplexity float32
	}{
		{"check file", 10, false, 0.3},
		{"build project", 50, false, 0.3},
		{"build entire project", 200, true, 0.6},
		{"commit all changes safely", 100, true, 0.6},
		{"refactor all code and test", 300, true, 0.7},
		{"compile and then test", 50, false, 0.5},
		{"parallel build and test", 50, true, 0.55},
	}

	d := NewDecomposer()
	for _, tt := range tests {
		ctx := &ProjectContext{
			RootPath:    "/tmp/test",
			SourceFiles: make([]string, tt.projectSize),
		}

		domain, complexity := d.classifyTask(tt.intent, ctx)
		t.Logf("intent=%q domain=%s complexity=%.2f", tt.intent, domain, complexity)

		if complexity < tt.wantMinComplexity {
			t.Errorf("%q: complexity %.2f < min %.2f", tt.intent, complexity, tt.wantMinComplexity)
		}

		shouldDecompose := complexity >= 0.6
		if shouldDecompose != tt.wantDecomposed {
			t.Errorf("%q: decompose=%v, want=%v (complexity=%.2f)",
				tt.intent, shouldDecompose, tt.wantDecomposed, complexity)
		}
	}
}

func BenchmarkDecompose_Simple(b *testing.B) {
	d := NewDecomposer()
	ctx := &ProjectContext{RootPath: "/tmp", SourceFiles: []string{"main.go"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Decompose("check file exists", ctx)
	}
}

func BenchmarkDecompose_Complex(b *testing.B) {
	d := NewDecomposer()
	ctx := &ProjectContext{RootPath: "/tmp/proj", SourceFiles: make([]string, 500), BuildSystem: "make"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Decompose("build and test entire project", ctx)
	}
}

func BenchmarkContextCompress(b *testing.B) {
	c := NewContextCompressor(64000)
	ctx := &ProjectContext{
		RootPath:    "/home/user/project",
		Languages:   []string{"Go", "Rust"},
		SourceFiles: make([]string, 1000),
		TestFiles:   make([]string, 200),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Compress(ctx, "build and test")
	}
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
