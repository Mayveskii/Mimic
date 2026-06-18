package cas

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

func newTestAssembler(t *testing.T) (*Assembler, string) {
	t.Helper()
	repoPath := initTestRepo(t)
	store, err := NewStore(repoPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return NewAssembler(store), repoPath
}

func TestAssembleWriteFileIntoEmptyTree(t *testing.T) {
	asm, repoPath := newTestAssembler(t)
	ctx := context.Background()

	newTree, err := asm.Assemble(ctx, emptyTreeSHA, []AssemblyStep{
		WriteFileStep{Path: "hello.txt", Content: []byte("world")},
	})
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	entries, err := asm.store.ReadTree(newTree)
	if err != nil {
		t.Fatalf("ReadTree: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 tree entry, got %d", len(entries))
	}
	if entries[0].Path != "hello.txt" {
		t.Fatalf("expected path hello.txt, got %q", entries[0].Path)
	}

	content, err := asm.store.ReadBlob(entries[0].SHA)
	if err != nil {
		t.Fatalf("ReadBlob: %v", err)
	}
	if !bytes.Equal(content, []byte("world")) {
		t.Fatalf("unexpected content %q", content)
	}

	diff := treeDiff(t, repoPath, emptyTreeSHA, newTree)
	if !strings.Contains(diff, "+world") {
		t.Fatalf("expected diff to contain +world, got:\n%s", diff)
	}
}

func TestAssembleWriteAndApplyPattern(t *testing.T) {
	asm, repoPath := newTestAssembler(t)
	ctx := context.Background()

	pattern := &mesh.TextSlot{
		ID:      "greet",
		Domain:  "demo",
		Actions: []string{"Hello $name!", "Welcome to $place."},
	}
	patternSHA, err := asm.store.RegisterPattern(pattern)
	if err != nil {
		t.Fatalf("RegisterPattern: %v", err)
	}

	newTree, err := asm.Assemble(ctx, emptyTreeSHA, []AssemblyStep{
		WriteFileStep{Path: "base.txt", Content: []byte("base content")},
		ApplyPatternStep{
			PatternSHA: patternSHA,
			TargetPath: "greeting.txt",
			Params: map[string]string{
				"name":  "Mimic",
				"place": "GAP-SHA-3",
			},
		},
	})
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	entries, err := asm.store.ReadTree(newTree)
	if err != nil {
		t.Fatalf("ReadTree: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 tree entries, got %d", len(entries))
	}

	greeting := readTreeFile(t, asm.store, entries, "greeting.txt")
	want := "Hello Mimic!\nWelcome to GAP-SHA-3."
	if string(greeting) != want {
		t.Fatalf("greeting.txt = %q, want %q", greeting, want)
	}

	diff := treeDiff(t, repoPath, emptyTreeSHA, newTree)
	if !strings.Contains(diff, "+base content") {
		t.Fatalf("expected diff to contain +base content, got:\n%s", diff)
	}
	if !strings.Contains(diff, "+Hello Mimic!") {
		t.Fatalf("expected diff to contain interpolated greeting, got:\n%s", diff)
	}
}

func TestAssembleEditPatch(t *testing.T) {
	asm, repoPath := newTestAssembler(t)
	ctx := context.Background()

	inputTree, err := asm.Assemble(ctx, emptyTreeSHA, []AssemblyStep{
		WriteFileStep{Path: "doc.txt", Content: []byte("line one\nline two\nline three\n")},
	})
	if err != nil {
		t.Fatalf("Assemble input: %v", err)
	}

	patch := `--- doc.txt
+++ doc.txt
@@ -1,3 +1,3 @@
 line one
-line two
+line two modified
 line three
`

	outputTree, err := asm.Assemble(ctx, inputTree, []AssemblyStep{
		EditPatchStep{Path: "doc.txt", Patch: patch},
	})
	if err != nil {
		t.Fatalf("Assemble patch: %v", err)
	}

	entries, err := asm.store.ReadTree(outputTree)
	if err != nil {
		t.Fatalf("ReadTree: %v", err)
	}

	content := readTreeFile(t, asm.store, entries, "doc.txt")
	want := "line one\nline two modified\nline three\n"
	if string(content) != want {
		t.Fatalf("doc.txt = %q, want %q", content, want)
	}

	diff := treeDiff(t, repoPath, inputTree, outputTree)
	if !strings.Contains(diff, "-line two") || !strings.Contains(diff, "+line two modified") {
		t.Fatalf("expected diff to show line change, got:\n%s", diff)
	}
}

func readTreeFile(t *testing.T, store *Store, entries []TreeEntry, name string) []byte {
	t.Helper()
	for _, e := range entries {
		if e.Path == name {
			data, err := store.ReadBlob(e.SHA)
			if err != nil {
				t.Fatalf("ReadBlob %s: %v", name, err)
			}
			return data
		}
	}
	t.Fatalf("file %q not found in tree", name)
	return nil
}

func treeDiff(t *testing.T, repoPath, treeA, treeB string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repoPath, "diff", treeA, treeB).CombinedOutput()
	if err != nil {
		t.Fatalf("git diff %s %s: %v\n%s", treeA, treeB, err, out)
	}
	return string(out)
}
