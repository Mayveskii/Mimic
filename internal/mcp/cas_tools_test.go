package mcp

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/cas"
	"github.com/Mayveskii/Mimic/internal/mesh"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	out, err := exec.Command("git", "init", tmpDir).CombinedOutput()
	if err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}
	return tmpDir
}

func newTestCASTools(t *testing.T) (*Server, *cas.Store, *cas.Assembler, string) {
	t.Helper()
	repoPath := initTestRepo(t)
	store, err := cas.NewStore(repoPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	assembler := cas.NewAssembler(store)
	server := NewServer(&dummyTransport{}, repoPath, "", "")
	RegisterCASTools(server, store, assembler)
	return server, store, assembler, repoPath
}

type dummyTransport struct{}

func (d *dummyTransport) Read() ([]byte, error) { return nil, nil }
func (d *dummyTransport) Write([]byte) error    { return nil }
func (d *dummyTransport) Close() error          { return nil }

func shaFromResult(t *testing.T, resp map[string]interface{}) string {
	t.Helper()
	if isErr, _ := resp["isError"].(bool); isErr {
		t.Fatalf("unexpected error response: %v", resp)
	}
	content, ok := resp["content"].([]map[string]string)
	if !ok || len(content) == 0 {
		t.Fatalf("unexpected response content: %v", resp)
	}
	var payload struct {
		TreeSHA   string `json:"tree_sha"`
		CommitSHA string `json:"commit_sha"`
		SHA       string `json:"sha"`
	}
	if err := json.Unmarshal([]byte(content[0]["text"]), &payload); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if payload.TreeSHA != "" {
		return payload.TreeSHA
	}
	if payload.CommitSHA != "" {
		return payload.CommitSHA
	}
	return payload.SHA
}

func TestCASTools_SchemasRegistered(t *testing.T) {
	server, _, _, _ := newTestCASTools(t)

	want := []string{
		"apply_pattern",
		"graft_module",
		"edit_via_patch",
		"write_file",
		"commit_transition",
		"read_blob",
		"read_tree",
	}

	found := make(map[string]bool)
	for _, tool := range server.tools {
		found[tool.Name] = true
	}
	for _, name := range want {
		if !found[name] {
			t.Fatalf("tool %q not registered", name)
		}
	}
}

func TestCASTools_WriteFileAndReadTree(t *testing.T) {
	server, store, _, _ := newTestCASTools(t)

	resp := server.dispatchCASTools("write_file", map[string]interface{}{
		"path":    "hello.txt",
		"content": "world",
	})
	treeSHA := shaFromResult(t, resp)
	if treeSHA == "" {
		t.Fatalf("expected non-empty tree SHA")
	}

	entries, err := store.ReadTree(treeSHA)
	if err != nil {
		t.Fatalf("ReadTree: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "hello.txt" {
		t.Fatalf("expected path hello.txt, got %q", entries[0].Path)
	}
}

func TestCASTools_ReadBlob(t *testing.T) {
	server, store, _, _ := newTestCASTools(t)

	blobSHA, err := store.StoreBlob([]byte("hello blob"))
	if err != nil {
		t.Fatalf("StoreBlob: %v", err)
	}

	resp := server.dispatchCASTools("read_blob", map[string]interface{}{"sha": blobSHA})
	if isErr, _ := resp["isError"].(bool); isErr {
		t.Fatalf("read_blob failed: %v", resp)
	}
	content, ok := resp["content"].([]map[string]string)
	if !ok || len(content) == 0 {
		t.Fatalf("unexpected response: %v", resp)
	}

	var payload struct {
		SHA     string `json:"sha"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(content[0]["text"]), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.SHA != blobSHA {
		t.Fatalf("sha mismatch: %q vs %q", payload.SHA, blobSHA)
	}
	if payload.Content != "hello blob" {
		t.Fatalf("content mismatch: %q", payload.Content)
	}
}

func TestCASTools_ApplyPattern(t *testing.T) {
	server, store, _, _ := newTestCASTools(t)

	pattern := &mesh.TextSlot{
		ID:      "demo",
		Domain:  "test",
		Actions: []string{"Hello $name!"},
	}
	patternSHA, err := store.RegisterPattern(pattern)
	if err != nil {
		t.Fatalf("RegisterPattern: %v", err)
	}

	resp := server.dispatchCASTools("apply_pattern", map[string]interface{}{
		"pattern_sha": patternSHA,
		"target_path": "greeting.txt",
		"params": map[string]interface{}{
			"name": "Mimic",
		},
	})
	treeSHA := shaFromResult(t, resp)

	entries, err := store.ReadTree(treeSHA)
	if err != nil {
		t.Fatalf("ReadTree: %v", err)
	}
	if len(entries) != 1 || entries[0].Path != "greeting.txt" {
		t.Fatalf("unexpected entries: %+v", entries)
	}

	data, err := store.ReadBlob(entries[0].SHA)
	if err != nil {
		t.Fatalf("ReadBlob: %v", err)
	}
	if strings.TrimSpace(string(data)) != "Hello Mimic!" {
		t.Fatalf("unexpected content: %q", data)
	}
}

func TestCASTools_GraftModule(t *testing.T) {
	server, store, _, _ := newTestCASTools(t)

	baseTree, err := store.StoreTree([]cas.TreeEntry{
		{Mode: "100644", Type: "blob", SHA: mustBlob(t, store, "a"), Path: "a.txt"},
	})
	if err != nil {
		t.Fatalf("StoreTree: %v", err)
	}

	resp := server.dispatchCASTools("graft_module", map[string]interface{}{
		"source_tree_sha": baseTree,
		"target_path":     "pkg",
	})
	newTree := shaFromResult(t, resp)

	entries, err := store.ReadTree(newTree)
	if err != nil {
		t.Fatalf("ReadTree: %v", err)
	}
	if len(entries) != 1 || entries[0].Path != "pkg" || entries[0].Type != "tree" {
		t.Fatalf("unexpected top entries: %+v", entries)
	}

	subEntries, err := store.ReadTree(entries[0].SHA)
	if err != nil {
		t.Fatalf("ReadTree sub: %v", err)
	}
	if len(subEntries) != 1 || subEntries[0].Path != "a.txt" {
		t.Fatalf("unexpected sub entries: %+v", subEntries)
	}
}

func TestCASTools_EditViaPatch(t *testing.T) {
	server, store, assembler, _ := newTestCASTools(t)

	baseTree, err := assembler.Assemble(context.Background(), cas.EmptyTreeSHA, []cas.AssemblyStep{
		cas.WriteFileStep{Path: "doc.txt", Content: []byte("line one\nline two\nline three\n")},
	})
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	// edit_via_patch starts from the empty tree, so seed the file as a blob first
	// and build a tree, then use it as the implicit current tree by writing the
	// same content into a fresh tree and patching. Because the tool uses the
	// empty tree, we first write the base content, then patch in a second call.
	_ = baseTree
	_ = store

	patch := `--- doc.txt
+++ doc.txt
@@ -1,3 +1,3 @@
 line one
-line two
+line two patched
 line three
`
	resp := server.dispatchCASTools("edit_via_patch", map[string]interface{}{
		"file_path": "doc.txt",
		"patch":     patch,
		"tree_sha":  baseTree,
	})
	treeSHA := shaFromResult(t, resp)

	entries, err := store.ReadTree(treeSHA)
	if err != nil {
		t.Fatalf("ReadTree: %v", err)
	}
	data, err := store.ReadBlob(entries[0].SHA)
	if err != nil {
		t.Fatalf("ReadBlob: %v", err)
	}
	if !strings.Contains(string(data), "line two patched") {
		t.Fatalf("patch not applied: %q", data)
	}
}

func TestCASTools_CommitTransition(t *testing.T) {
	server, store, _, repoPath := newTestCASTools(t)

	blobSHA, err := store.StoreBlob([]byte("content"))
	if err != nil {
		t.Fatalf("StoreBlob: %v", err)
	}
	treeSHA, err := store.StoreTree([]cas.TreeEntry{
		{Mode: "100644", Type: "blob", SHA: blobSHA, Path: "file.txt"},
	})
	if err != nil {
		t.Fatalf("StoreTree: %v", err)
	}

	resp := server.dispatchCASTools("commit_transition", map[string]interface{}{
		"tree_sha":   treeSHA,
		"parent_sha": "",
		"message":    "initial commit",
	})
	commitSHA := shaFromResult(t, resp)
	if commitSHA == "" {
		t.Fatalf("expected non-empty commit SHA")
	}

	// Verify the commit object exists.
	out, err := exec.Command("git", "-C", repoPath, "cat-file", "-t", commitSHA).CombinedOutput()
	if err != nil {
		t.Fatalf("cat-file failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "commit" {
		t.Fatalf("expected commit object, got %q", out)
	}
}

func mustBlob(t *testing.T, store *cas.Store, content string) string {
	t.Helper()
	sha, err := store.StoreBlob([]byte(content))
	if err != nil {
		t.Fatalf("StoreBlob: %v", err)
	}
	return sha
}
