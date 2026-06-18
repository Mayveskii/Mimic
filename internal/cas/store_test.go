package cas

import (
	"bytes"
	"os/exec"
	"testing"
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

func TestStoreBlobDeduplication(t *testing.T) {
	store, err := NewStore(initTestRepo(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	data := []byte("hello, GAP-SHA-1")
	sha1, err := store.StoreBlob(data)
	if err != nil {
		t.Fatalf("StoreBlob first: %v", err)
	}
	sha2, err := store.StoreBlob(data)
	if err != nil {
		t.Fatalf("StoreBlob second: %v", err)
	}

	if sha1 == "" {
		t.Fatal("expected non-empty SHA")
	}
	if sha1 != sha2 {
		t.Fatalf("expected same SHA for duplicate data, got %q and %q", sha1, sha2)
	}
}

func TestStoreTreeRoundTrip(t *testing.T) {
	store, err := NewStore(initTestRepo(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	aBlob, err := store.StoreBlob([]byte("alpha"))
	if err != nil {
		t.Fatalf("StoreBlob alpha: %v", err)
	}
	bBlob, err := store.StoreBlob([]byte("beta"))
	if err != nil {
		t.Fatalf("StoreBlob beta: %v", err)
	}

	want := []TreeEntry{
		{Mode: "100644", Type: "blob", SHA: aBlob, Path: "a.txt"},
		{Mode: "100644", Type: "blob", SHA: bBlob, Path: "b.txt"},
	}
	// Deliberately unordered to exercise internal sorting.
	unordered := []TreeEntry{
		{Mode: "100644", Type: "blob", SHA: bBlob, Path: "b.txt"},
		{Mode: "100644", Type: "blob", SHA: aBlob, Path: "a.txt"},
	}

	treeSHA, err := store.StoreTree(unordered)
	if err != nil {
		t.Fatalf("StoreTree: %v", err)
	}
	if treeSHA == "" {
		t.Fatal("expected non-empty tree SHA")
	}

	got, err := store.ReadTree(treeSHA)
	if err != nil {
		t.Fatalf("ReadTree: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d entries, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d mismatch: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestReadBlob(t *testing.T) {
	store, err := NewStore(initTestRepo(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	data := []byte("content-addressable storage")
	sha, err := store.StoreBlob(data)
	if err != nil {
		t.Fatalf("StoreBlob: %v", err)
	}

	read, err := store.ReadBlob(sha)
	if err != nil {
		t.Fatalf("ReadBlob: %v", err)
	}
	if !bytes.Equal(read, data) {
		t.Fatalf("ReadBlob returned %q, want %q", read, data)
	}
}

func TestStoreCommit(t *testing.T) {
	store, err := NewStore(initTestRepo(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	blob, err := store.StoreBlob([]byte("root content"))
	if err != nil {
		t.Fatalf("StoreBlob: %v", err)
	}
	tree, err := store.StoreTree([]TreeEntry{
		{Mode: "100644", Type: "blob", SHA: blob, Path: "file.txt"},
	})
	if err != nil {
		t.Fatalf("StoreTree: %v", err)
	}

	rootCommit, err := store.StoreCommit(tree, "", "root commit")
	if err != nil {
		t.Fatalf("StoreCommit root: %v", err)
	}
	if len(rootCommit) != 40 {
		t.Fatalf("expected 40-character root commit SHA, got %q", rootCommit)
	}

	childCommit, err := store.StoreCommit(tree, rootCommit, "child commit")
	if err != nil {
		t.Fatalf("StoreCommit child: %v", err)
	}
	if len(childCommit) != 40 {
		t.Fatalf("expected 40-character child commit SHA, got %q", childCommit)
	}
	if childCommit == rootCommit {
		t.Fatal("expected child commit SHA to differ from root commit SHA")
	}
}
