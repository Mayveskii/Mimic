package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Mayveskii/Mimic/internal/config"
	"github.com/Mayveskii/Mimic/internal/cost"
	"github.com/Mayveskii/Mimic/internal/event"
	"github.com/Mayveskii/Mimic/internal/graphify"
	"github.com/Mayveskii/Mimic/internal/hunt"
	"github.com/Mayveskii/Mimic/internal/mcp"
	"github.com/Mayveskii/Mimic/internal/mesh"
	"github.com/Mayveskii/Mimic/internal/model"
	"github.com/Mayveskii/Mimic/internal/orchestrator"
	"github.com/Mayveskii/Mimic/internal/session"
)

func main() {
	var (
		repoPath = flag.String("repo", ".", "Path to target repository")
		modelID  = flag.String("model", "qwen", "Model ID (qwen, kimi, minimax)")
		intent   = flag.String("intent", "", "Intent for the model (e.g. 'fix race condition')")
		meshDir  = flag.String("mesh", "", "Path to mesh slots directory")
	)
	flag.Parse()

	if *intent == "" {
		fmt.Fprintln(os.Stderr, "Usage: playground -repo=/path/to/rtk -model=qwen -intent='fix race condition'")
		os.Exit(1)
	}

	fmt.Printf("v16Q Playground\n")
	fmt.Printf("  repo:  %s\n", *repoPath)
	fmt.Printf("  model: %s\n", *modelID)
	fmt.Printf("  intent: %s\n", *intent)
	fmt.Println()

	// Load repo config and set up hot-reload watcher.
	cfg, err := config.LoadRepoConfig(*repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config load failed: %v\n", err)
		os.Exit(1)
	}
	watcher, err := config.NewWatcher(*repoPath, func(updated *config.RepoConfig) {
		cfg = updated
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config watcher failed: %v\n", err)
		os.Exit(1)
	}
	defer watcher.Stop()

	// 1. Session Manager
	mgr := session.NewManager(*repoPath)
	ctx, err := mgr.Init(*modelID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Session init failed: %v\n", err)
		os.Exit(1)
	}
	defer mgr.Destroy(ctx)

	fmt.Printf("Session: %s\n", ctx.ID)
	fmt.Printf("Worktree: %s\n", ctx.WorktreePath)
	fmt.Printf("Budget: %s\n", ctx.Budget)
	fmt.Println()

	// 2. Event Stream
	eventStream := event.NewStream(ctx.WorktreePath)
	_ = eventStream.Append(event.EventObservation, "system", fmt.Sprintf("intent: %s", *intent), nil)

	// 3. Cost Tracker
	costTracker := cost.NewTracker(ctx.WorktreePath)

	// 4. Model provider and cascade.
	gonkaCfg, ok := cfg.Providers["gonkagate"]
	if !ok {
		fmt.Fprintf(os.Stderr, "GonkaGate provider not configured\n")
		os.Exit(1)
	}
	provider, err := model.NewGonkaGateProvider(gonkaCfg.Endpoint, gonkaCfg.EnvKey, gonkaCfg.TimeoutMs, gonkaCfg.RetryMax)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GonkaGate provider failed: %v\n", err)
		os.Exit(1)
	}
	cascade := model.NewCascade(provider, cfg.Models)
	mgr = mgr.WithCaller(cascade)

	// 5. Context assembler with mesh and graphify.
	assembler := orchestrator.NewContextAssembler(cfg)
	if *meshDir != "" {
		store, err := mesh.NewStore(filepath.Join(*meshDir, "mesh.db"))
		if err == nil {
			defer store.Close()
			assembler.Mesh = store
		}
	}
	graph := graphify.NewGraph()
	if err := graph.BuildFromRepo(*repoPath); err == nil && len(graph.Nodes) > 0 {
		graph.ComputePageRank(10, 0.85)
		assembler.Graphify = graph
	}
	mgr = mgr.WithAssembler(assembler)

	// 6. Hunt + Pipeline
	hunter := hunt.NewHunter(*meshDir)
	tools := mcp.ToModelTools(mcp.DefaultSchemas)
	pipeline := orchestrator.NewPipeline(hunter, cascade, tools, assembler)

	result, err := pipeline.Run(context.Background(), ctx, *intent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pipeline failed: %v\n", err)
		_ = eventStream.Append(event.EventError, "system", err.Error(), nil)
		os.Exit(1)
	}

	// 7. Collect proof
	proof, err := mgr.Finalize(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Finalize failed: %v\n", err)
		os.Exit(1)
	}

	// 8. Record cost
	_ = costTracker.Record(cost.Entry{
		PersonaID: *modelID,
		Model:     ctx.Config.Models.Local,
		TokensIn:  len(*intent),
		TokensOut: len(result.Output),
		SessionID: ctx.ID,
	})

	// 9. Output
	fmt.Printf("=== Result ===\n%s\n\n", result.Output)
	fmt.Printf("=== Proof ===\n")
	fmt.Printf("Git Status:\n%s\n\n", proof.GitStatus)
	fmt.Printf("Git Log:\n%s\n\n", proof.GitLog)
	if proof.GitDiff != "" {
		fmt.Printf("Git Diff:\n%s\n\n", proof.GitDiff)
	}
	fmt.Printf("=== Cost ===\n%s\n", costTracker.Total())
}
