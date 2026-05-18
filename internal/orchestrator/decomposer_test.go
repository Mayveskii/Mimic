package orchestrator

import (
	"testing"

	"github.com/Mayveskii/Mimic/internal/cgo"
)

func TestNewDecomposer(t *testing.T) {
	d := NewDecomposer()
	if d == nil {
		t.Fatal("NewDecomposer() returned nil")
	}
	if d.maxTasks != 50 {
		t.Errorf("expected maxTasks=50, got %d", d.maxTasks)
	}
	if d.contextWindow != 128000 {
		t.Errorf("expected contextWindow=128000, got %d", d.contextWindow)
	}
}

func TestDecomposer_ClassifyTask(t *testing.T) {
	d := NewDecomposer()
	tests := []struct {
		intent        string
		wantDomain    string
		minComplexity float32
	}{
		{"run git commit", "git", 0.2},
		{"check if file exists", "system", 0.2},
		{"build the project", "build", 0.2},
		{"compile and run tests", "build", 0.3},
		{"refactor all functions", "analysis", 0.3},
		{"do something", "general", 0.1},
	}

	ctx := &ProjectContext{RootPath: "/tmp/test"}
	for _, tt := range tests {
		domain, complexity := d.classifyTask(tt.intent, ctx)
		if domain != tt.wantDomain {
			t.Errorf("classifyTask(%q) domain=%q, want %q", tt.intent, domain, tt.wantDomain)
		}
		if complexity < tt.minComplexity {
			t.Errorf("classifyTask(%q) complexity=%.2f, want >= %.2f", tt.intent, complexity, tt.minComplexity)
		}
	}
}

func TestDecomposer_Decompose_Simple(t *testing.T) {
	d := NewDecomposer()
	ctx := &ProjectContext{
		RootPath:    "/tmp/test",
		SourceFiles: []string{"main.go"},
	}

	// Simple task should NOT be decomposed
	graph, err := d.Decompose("check file exists", ctx)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	if len(graph.Tasks) != 1 {
		t.Errorf("simple task: expected 1 task, got %d", len(graph.Tasks))
	}
	if len(graph.ParallelGroups) != 1 {
		t.Errorf("simple task: expected 1 parallel group, got %d", len(graph.ParallelGroups))
	}
}

func TestDecomposer_Decompose_ComplexBuild(t *testing.T) {
	d := NewDecomposer()
	ctx := &ProjectContext{
		RootPath:    "/tmp/test",
		SourceFiles: make([]string, 200), // > 100 files = complexity bump
		BuildSystem: "make",
	}

	graph, err := d.Decompose("build and test the entire project", ctx)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Complex build should decompose into multiple tasks
	if len(graph.Tasks) < 2 {
		t.Errorf("complex build: expected >=2 tasks, got %d", len(graph.Tasks))
	}

	// Check sequential dependencies (clean → compile → test)
	hasCompile := false
	hasTest := false
	for _, task := range graph.Tasks {
		if task.Domain == "build" && task.Description == "Compile project" {
			hasCompile = true
		}
		if task.Domain == "build" && task.Description == "Run tests" {
			hasTest = true
		}
	}
	if !hasCompile {
		t.Error("complex build: missing 'Compile project' task")
	}
	if !hasTest {
		t.Error("complex build: missing 'Run tests' task")
	}
}

func TestDecomposer_Decompose_ComplexGit(t *testing.T) {
	d := NewDecomposer()
	ctx := &ProjectContext{
		RootPath:    "/tmp/test",
		SourceFiles: make([]string, 50),
	}

	graph, err := d.Decompose("commit all changes safely", ctx)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Should have status, add, commit
	if len(graph.Tasks) < 2 {
		t.Errorf("git workflow: expected >=2 tasks, got %d", len(graph.Tasks))
	}
}

func TestDecomposer_Decompose_ComplexAnalysis(t *testing.T) {
	d := NewDecomposer()
	ctx := &ProjectContext{
		RootPath:    "/tmp/test",
		SourceFiles: make([]string, 500), // > 1000 would add more
		BuildSystem: "go",
	}

	graph, err := d.Decompose("refactor the entire codebase and run tests", ctx)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Refactoring should produce: scan → analyze → refactor → verify
	if len(graph.Tasks) < 2 {
		t.Errorf("refactor: expected >=2 tasks, got %d", len(graph.Tasks))
	}
}

func TestDecomposer_FindParallelGroups(t *testing.T) {
	d := NewDecomposer()
	tasks := map[string]*Task{
		"a": {ID: "a", Dependencies: []string{}},
		"b": {ID: "b", Dependencies: []string{"a"}},
		"c": {ID: "c", Dependencies: []string{"a"}},
		"d": {ID: "d", Dependencies: []string{"b", "c"}},
	}

	groups := d.findParallelGroups(tasks)

	// Group 0: [a] (no deps)
	// Group 1: [b, c] (deps on a satisfied)
	// Group 2: [d] (deps on b,c satisfied)
	if len(groups) != 3 {
		t.Fatalf("expected 3 parallel groups, got %d: %v", len(groups), groups)
	}

	// Check first group has 'a'
	foundA := false
	for _, id := range groups[0] {
		if id == "a" {
			foundA = true
		}
	}
	if !foundA {
		t.Errorf("group[0] should contain 'a', got %v", groups[0])
	}

	// Check second group has b and c (in any order)
	if len(groups[1]) != 2 {
		t.Errorf("group[1] should have 2 tasks, got %d", len(groups[1]))
	}
}

func TestContextCompressor_Compress(t *testing.T) {
	c := NewContextCompressor(64000)
	ctx := &ProjectContext{
		RootPath:    "/home/user/project",
		Languages:   []string{"Go", "Rust"},
		BuildSystem: "cargo",
		SourceFiles: make([]string, 150),
		TestFiles:   make([]string, 30),
		EntryPoints: []string{"cmd/main.go"},
	}

	compressed := c.Compress(ctx, "build and test")

	// Should contain key info, be concise
	if compressed == "" {
		t.Error("Compress returned empty string")
	}
	if len(compressed) > 500 {
		t.Errorf("compressed too long: %d chars (should be < 500)", len(compressed))
	}
	if !contains(compressed, "Go") || !contains(compressed, "Rust") {
		t.Error("compressed should mention languages")
	}
	if !contains(compressed, "150 source") {
		t.Error("compressed should mention file counts")
	}
}

func TestProjectNavigator_DetectProject(t *testing.T) {
	n := NewProjectNavigator("/tmp")
	ctx := n.DetectProject()

	if ctx == nil {
		t.Fatal("DetectProject returned nil")
	}
	if ctx.RootPath != "/tmp" {
		t.Errorf("RootPath=%q, want /tmp", ctx.RootPath)
	}
	// Should detect config files
	if len(ctx.ConfigFiles) == 0 {
		t.Error("expected some config files detected")
	}
}

func TestNewOrchestrator_WithDecomposer(t *testing.T) {
	o := NewOrchestrator(10000, 60000)
	if o.decomposer == nil {
		t.Error("expected decomposer to be initialized")
	}
	if o.compressor == nil {
		t.Error("expected compressor to be initialized")
	}
}

func TestOrchestrator_RunComplex_Build(t *testing.T) {
	o := NewOrchestrator(10000, 60000)
	ctx := &ProjectContext{
		RootPath:    "/tmp/test",
		SourceFiles: make([]string, 200),
		BuildSystem: "make",
	}

	result, err := o.RunComplex("build and test the entire project", ctx)
	if err != nil {
		t.Fatalf("RunComplex failed: %v", err)
	}

	// Should have phases from task execution + RESPOND
	if len(result.Phases) < 2 {
		t.Errorf("expected >=2 phases, got %d", len(result.Phases))
	}

	// Check final output has decomposition info
	output, ok := result.FinalOutput.(map[string]interface{})
	if !ok {
		t.Fatal("expected FinalOutput to be map[string]interface{}")
	}
	if output["decomposition"] != true {
		t.Error("expected decomposition flag in response")
	}
	if _, ok := output["task_count"]; !ok {
		t.Error("expected task_count in response")
	}
}

func TestOrchestrator_RunComplex_EmptyProject(t *testing.T) {
	o := NewOrchestrator(10000, 60000)
	ctx := &ProjectContext{
		RootPath:    "/tmp",
		SourceFiles: []string{},
	}

	result, err := o.RunComplex("check if file exists", ctx)
	if err != nil {
		t.Fatalf("RunComplex failed: %v", err)
	}

	// Simple task on empty project == single task
	if len(result.Phases) < 1 {
		t.Errorf("expected >=1 phases, got %d", len(result.Phases))
	}
}

func TestOrchestrator_RunSimple_StillWorks(t *testing.T) {
	// Ensure C-core is initialized
	_ = cgo.Init()
	defer func() {
		// Don't actually shutdown if other tests need it
	}()

	// Ensure Run() still works for simple single-task calls
	o := NewOrchestrator(10000, 60000)
	result, err := o.Run("SYS_FILE_EXISTS", map[string]interface{}{"path": "/tmp"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.Phases) != 6 {
		t.Errorf("simple Run: expected 6 phases, got %d", len(result.Phases))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
