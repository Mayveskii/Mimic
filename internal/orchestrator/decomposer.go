package orchestrator

import (
	"fmt"
	"strings"

	"github.com/Mayveskii/Mimic/internal/cgo"
)

// Task represents a decomposed unit of work
type Task struct {
	ID           string
	Description  string
	Domain       string       // git, build, io, system, analysis, mixed
	Safety       int          // 0=critical, 1=dangerous, 2=mutating, 3=safe
	Complexity   float32      // 0.0-1.0 estimated complexity
	Dependencies []string     // task IDs that must complete before this one
	Operations   []cgo.Packet // planned operations
	EstTokens    float32      // estimated token cost
	EstTimeMS    float32      // estimated time cost
}

// TaskGraph represents a decomposed project as DAG
type TaskGraph struct {
	Tasks          map[string]*Task
	RootID         string
	ParallelGroups [][]string // tasks that can run in parallel
}

// ProjectContext holds understanding of the project structure
type ProjectContext struct {
	RootPath    string
	Languages   []string // detected programming languages
	Frameworks  []string // detected frameworks
	EntryPoints []string // main files (main.go, index.js, etc.)
	ConfigFiles []string // package.json, go.mod, Cargo.toml, etc.
	SourceFiles []string // all source files
	TestFiles   []string // test files
	BuildSystem string   // make, npm, cargo, etc.
	GitStatus   string   // clean/dirty/branch
}

// Decomposer breaks down model intent into executable tasks
type Decomposer struct {
	maxTasks      int
	contextWindow int // tokens available for context
}

func NewDecomposer() *Decomposer {
	return &Decomposer{
		maxTasks:      50,
		contextWindow: 128000, // kimi k2.6 context window
	}
}

// Decompose analyzes intent and breaks it into task graph
func (d *Decomposer) Decompose(intent string, projectCtx *ProjectContext) (*TaskGraph, error) {
	graph := &TaskGraph{
		Tasks: make(map[string]*Task),
	}

	// Phase 1: Understand task type and complexity
	taskType, complexity := d.classifyTask(intent, projectCtx)

	// Phase 2: For complex tasks (>= 0.6), decompose into subtasks
	if complexity >= 0.6 {
		subtasks := d.breakIntoSubtasks(intent, taskType, projectCtx)
		for _, st := range subtasks {
			graph.Tasks[st.ID] = st
		}
		graph.ParallelGroups = d.findParallelGroups(graph.Tasks)
	} else {
		// Simple task — single node
		task := &Task{
			ID:          "t-1",
			Description: intent,
			Domain:      taskType,
			Complexity:  complexity,
		}
		graph.Tasks[task.ID] = task
		graph.RootID = task.ID
		graph.ParallelGroups = [][]string{{task.ID}}
	}

	return graph, nil
}

func (d *Decomposer) classifyTask(intent string, ctx *ProjectContext) (string, float32) {
	lower := strings.ToLower(intent)

	// Determine domain - check more specific domains first
	domain := "general"
	if strings.Contains(lower, "refactor") || strings.Contains(lower, "rename") || strings.Contains(lower, "extract") || strings.Contains(lower, "analyze") {
		domain = "analysis"
	} else if strings.Contains(lower, "git") || strings.Contains(lower, "commit") || strings.Contains(lower, "branch") || strings.Contains(lower, "merge") {
		domain = "git"
	} else if strings.Contains(lower, "build") || strings.Contains(lower, "compile") || strings.Contains(lower, "deploy") {
		domain = "build"
	} else if strings.Contains(lower, "test") {
		// Only classify as build if test is not combined with analysis keywords
		if strings.Contains(lower, "analyze") || strings.Contains(lower, "refactor") {
			domain = "analysis"
		} else {
			domain = "build"
		}
	} else if strings.Contains(lower, "file") || strings.Contains(lower, "dir") || strings.Contains(lower, "path") || strings.Contains(lower, "move") || strings.Contains(lower, "copy") {
		domain = "system"
	}

	// Estimate complexity based on keywords and project context
	complexity := float32(0.3) // base

	// High-scope keywords
	if strings.Contains(lower, "all") || strings.Contains(lower, "every") || strings.Contains(lower, "entire") || strings.Contains(lower, "safely") || strings.Contains(lower, "entire project") || strings.Contains(lower, "all changes") || strings.Contains(lower, "codebase") {
		complexity += 0.3
	}
	// Task chaining keywords
	if strings.Contains(lower, "and then") || strings.Contains(lower, "after that") || strings.Contains(lower, "next") || strings.Contains(lower, "run tests") {
		complexity += 0.2
	}
	// Conditional keywords
	if strings.Contains(lower, "if") || strings.Contains(lower, "unless") || strings.Contains(lower, "check") {
		complexity += 0.1
	}
	// Parallel keywords
	if strings.Contains(lower, "parallel") || strings.Contains(lower, "concurrent") {
		complexity += 0.15
	}
	// Domain-specific complexity boosters
	if strings.Contains(lower, "refactor") || strings.Contains(lower, "rewrite") {
		complexity += 0.2
	}
	if strings.Contains(lower, "analyze") && strings.Contains(lower, "dependencies") {
		complexity += 0.2
	}
	// Project size factor
	if len(ctx.SourceFiles) > 100 {
		complexity += 0.1
	}
	if len(ctx.SourceFiles) > 1000 {
		complexity += 0.2
	}

	if complexity > 1.0 {
		complexity = 1.0
	}

	return domain, complexity
}

func (d *Decomposer) breakIntoSubtasks(intent string, domain string, ctx *ProjectContext) []*Task {
	var tasks []*Task

	// Parse intent for sequential/parallel operations
	// This is a heuristic decomposition — production would use NLP

	if domain == "build" {
		tasks = append(tasks, &Task{
			ID:          "t-1",
			Description: "Clean build artifacts",
			Domain:      "build",
			Safety:      1,
			Complexity:  0.2,
			Operations:  planOps("BUILD_CLEAN", map[string]interface{}{"target": "all"}),
		})
		tasks = append(tasks, &Task{
			ID:           "t-2",
			Description:  "Compile project",
			Domain:       "build",
			Safety:       2,
			Complexity:   0.5,
			Dependencies: []string{"t-1"},
			Operations:   planOps("BUILD_COMPILE", map[string]interface{}{"target": "all"}),
		})
		tasks = append(tasks, &Task{
			ID:           "t-3",
			Description:  "Run tests",
			Domain:       "build",
			Safety:       3,
			Complexity:   0.6,
			Dependencies: []string{"t-2"},
			Operations:   planOps("BUILD_TEST", map[string]interface{}{"dir": ctx.RootPath}),
		})
	} else if domain == "git" {
		tasks = append(tasks, &Task{
			ID:          "t-1",
			Description: "Check git status",
			Domain:      "git",
			Safety:      3,
			Complexity:  0.2,
			Operations:  planOps("GIT_STATUS", nil),
		})
		tasks = append(tasks, &Task{
			ID:           "t-2",
			Description:  "Stage changes",
			Domain:       "git",
			Safety:       2,
			Complexity:   0.3,
			Dependencies: []string{"t-1"},
			Operations:   planOps("GIT_ADD", nil),
		})
		tasks = append(tasks, &Task{
			ID:           "t-3",
			Description:  "Commit changes",
			Domain:       "git",
			Safety:       0,
			Complexity:   0.4,
			Dependencies: []string{"t-2"},
			Operations:   planOps("GIT_COMMIT", map[string]interface{}{"message": " wip"}),
		})
	} else if domain == "analysis" {
		// Check if this is pure analysis or refactoring
		lower := strings.ToLower(intent)
		isRefactor := strings.Contains(lower, "refactor") || strings.Contains(lower, "rename") || strings.Contains(lower, "extract") || strings.Contains(lower, "rewrite")

		tasks = append(tasks, &Task{
			ID:          "t-1",
			Description: "Scan project structure",
			Domain:      "system",
			Safety:      3,
			Complexity:  0.3,
			Operations:  planOps("SYS_FILE_EXISTS", map[string]interface{}{"path": ctx.RootPath}),
		})
		tasks = append(tasks, &Task{
			ID:           "t-2",
			Description:  "Analyze dependencies",
			Domain:       "analysis",
			Safety:       3,
			Complexity:   0.5,
			Dependencies: []string{"t-1"},
		})
		if isRefactor {
			tasks = append(tasks, &Task{
				ID:           "t-3",
				Description:  "Execute refactoring",
				Domain:       "system",
				Safety:       2,
				Complexity:   0.7,
				Dependencies: []string{"t-2"},
			})
			tasks = append(tasks, &Task{
				ID:           "t-4",
				Description:  "Verify with tests",
				Domain:       "build",
				Safety:       3,
				Complexity:   0.4,
				Dependencies: []string{"t-3"},
				Operations:   planOps("BUILD_TEST", map[string]interface{}{"dir": ctx.RootPath}),
			})
		}
	}

	return tasks
}

func (d *Decomposer) findParallelGroups(tasks map[string]*Task) [][]string {
	// Find tasks with no dependencies or all dependencies satisfied
	var groups [][]string
	visited := make(map[string]bool)

	for len(visited) < len(tasks) {
		var group []string
		for id, task := range tasks {
			if visited[id] {
				continue
			}
			// Check all dependencies are visited
			allDepsDone := true
			for _, dep := range task.Dependencies {
				if !visited[dep] {
					allDepsDone = false
					break
				}
			}
			if allDepsDone {
				group = append(group, id)
			}
		}
		if len(group) == 0 {
			break // deadlock or done
		}
		for _, id := range group {
			visited[id] = true
		}
		groups = append(groups, group)
	}

	return groups
}

func planOps(opcode string, args map[string]interface{}) []cgo.Packet {
	if args == nil {
		args = map[string]interface{}{}
	}
	pkt, _ := cgo.PacketFromToolCall(opcode, args)
	return []cgo.Packet{pkt}
}

// ContextCompressor compresses large contexts to fit within token budget
type ContextCompressor struct {
	targetTokens int
}

func NewContextCompressor(target int) *ContextCompressor {
	return &ContextCompressor{targetTokens: target}
}

// Compress reduces context to essential information
func (c *ContextCompressor) Compress(projectCtx *ProjectContext, intent string) string {
	// Strategy: keep structure, compress content
	var parts []string

	parts = append(parts, fmt.Sprintf("Project: %s", projectCtx.RootPath))
	parts = append(parts, fmt.Sprintf("Languages: %s", strings.Join(projectCtx.Languages, ", ")))
	parts = append(parts, fmt.Sprintf("Build: %s", projectCtx.BuildSystem))
	parts = append(parts, fmt.Sprintf("Files: %d source, %d test", len(projectCtx.SourceFiles), len(projectCtx.TestFiles)))
	parts = append(parts, fmt.Sprintf("Entry: %s", strings.Join(projectCtx.EntryPoints, ", ")))
	parts = append(parts, fmt.Sprintf("Task: %s", intent))

	return strings.Join(parts, "\n")
}

// ProjectNavigator understands project structure
type ProjectNavigator struct {
	rootPath string
}

func NewProjectNavigator(root string) *ProjectNavigator {
	return &ProjectNavigator{rootPath: root}
}

// DetectProject scans directory and returns project context
func (n *ProjectNavigator) DetectProject() *ProjectContext {
	ctx := &ProjectContext{
		RootPath:    n.rootPath,
		Languages:   []string{},
		Frameworks:  []string{},
		EntryPoints: []string{},
		ConfigFiles: []string{},
		SourceFiles: []string{},
		TestFiles:   []string{},
	}

	// Heuristic detection based on file extensions
	// Production: use filesystem walk + pattern matching
	// Check for config files
	configPatterns := map[string]string{
		"go.mod": "Go", "package.json": "Node", "Cargo.toml": "Rust",
		"requirements.txt": "Python", "pom.xml": "Java", "CMakeLists.txt": "C++",
		"Makefile": "Make", "Dockerfile": "Docker",
	}

	for file, lang := range configPatterns {
		ctx.ConfigFiles = append(ctx.ConfigFiles, file)
		ctx.Languages = append(ctx.Languages, lang)
		if lang == "Go" {
			ctx.BuildSystem = "go"
		} else if lang == "Rust" {
			ctx.BuildSystem = "cargo"
		} else if file == "Makefile" {
			ctx.BuildSystem = "make"
		}
	}

	// Remove duplicates
	ctx.Languages = uniqueStrings(ctx.Languages)

	return ctx
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}
