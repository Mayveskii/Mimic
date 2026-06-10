package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/Mayveskii/Mimic/internal/cost"
	"github.com/Mayveskii/Mimic/internal/event"
	"github.com/Mayveskii/Mimic/internal/hunt"
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

	// 4. Hunt + Pipeline
	hunter := hunt.NewHunter(*meshDir)
	pipeline := orchestrator.NewPipeline(hunter)

	result, err := pipeline.Run(context.Background(), ctx, *intent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pipeline failed: %v\n", err)
		_ = eventStream.Append(event.EventError, "system", err.Error(), nil)
		os.Exit(1)
	}

	// 5. Collect proof
	proof, err := mgr.Finalize(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Finalize failed: %v\n", err)
		os.Exit(1)
	}

	// 6. Record cost
	_ = costTracker.Record(cost.Entry{
		PersonaID: *modelID,
		Model:     ctx.Config.Models.Local,
		TokensIn:  len(*intent),
		TokensOut: len(result.Output),
		SessionID: ctx.ID,
	})

	// 7. Output
	fmt.Printf("=== Result ===\n%s\n\n", result.Output)
	fmt.Printf("=== Proof ===\n")
	fmt.Printf("Git Status:\n%s\n\n", proof.GitStatus)
	fmt.Printf("Git Log:\n%s\n\n", proof.GitLog)
	if proof.GitDiff != "" {
		fmt.Printf("Git Diff:\n%s\n\n", proof.GitDiff)
	}
	fmt.Printf("=== Cost ===\n%s\n", costTracker.Total())
}
