// Package projectmap implements embryo-style workspace indexing.
// Creates a SQLite+FTS5 database at .mimic/projectmap.db for fast
// file/symbol/text search across the entire workspace.
package projectmap

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

const dbName = ".mimic/projectmap.db"

// ProjectMap holds the SQLite index for a workspace.
type ProjectMap struct {
	db         *sql.DB
	mu         sync.RWMutex
	root       string
	dirtyFiles map[string]bool // files changed since last commit
}

// OpenOrCreate initializes the projectmap database in the workspace.
func OpenOrCreate(root string) (*ProjectMap, error) {
	dbPath := filepath.Join(root, dbName)
	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open projectmap db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping projectmap db: %w", err)
	}

	pm := &ProjectMap{db: db, root: root, dirtyFiles: make(map[string]bool)}
	if err := pm.createSchema(); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	return pm, nil
}

func (pm *ProjectMap) createSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS files (
    id          INTEGER PRIMARY KEY,
    path        TEXT UNIQUE NOT NULL,
    content_hash TEXT NOT NULL,
    size        INTEGER NOT NULL,
    lang        TEXT,
    mtime       INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
CREATE INDEX IF NOT EXISTS idx_files_lang ON files(lang);

CREATE TABLE IF NOT EXISTS symbols (
    id       INTEGER PRIMARY KEY,
    name     TEXT NOT NULL,
    type     TEXT NOT NULL,
    file_id  INTEGER NOT NULL,
    line     INTEGER NOT NULL,
    signature TEXT,
    FOREIGN KEY(file_id) REFERENCES files(id)
);
CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name);
CREATE INDEX IF NOT EXISTS idx_symbols_type ON symbols(type);
CREATE INDEX IF NOT EXISTS idx_symbols_file ON symbols(file_id);

CREATE VIRTUAL TABLE IF NOT EXISTS fts USING fts5(
    path, content, content_rowid=rowid, tokenize='porter unicode61'
);

CREATE TABLE IF NOT EXISTS imports (
    id       INTEGER PRIMARY KEY,
    file_id  INTEGER NOT NULL,
    module   TEXT NOT NULL,
    FOREIGN KEY(file_id) REFERENCES files(id)
);
CREATE INDEX IF NOT EXISTS idx_imports_module ON imports(module);
`
	_, err := pm.db.Exec(schema)
	return err
}

// Close releases the database.
func (pm *ProjectMap) Close() error {
	return pm.db.Close()
}

// Root returns the workspace root directory.
func (pm *ProjectMap) Root() string {
	return pm.root
}

// IndexWorkspace performs a full scan and reindex of the workspace.
func (pm *ProjectMap) IndexWorkspace() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	tx, err := pm.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Truncate and rebuild
	for _, table := range []string{"imports", "symbols", "fts", "files"} {
		if _, err := tx.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("truncate %s: %w", table, err)
		}
	}

	err = filepath.WalkDir(pm.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(pm.root, path)
		if shouldSkip(rel, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if err := pm.indexFileTx(tx, path, rel); err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

// IndexFile updates or inserts a single file into the index.
func (pm *ProjectMap) IndexFile(absPath string) error {
	rel, err := filepath.Rel(pm.root, absPath)
	if err != nil {
		return err
	}
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.indexFileTx(pm.db, absPath, rel)
}

func (pm *ProjectMap) indexFileTx(tx executor, absPath, rel string) error {
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	f, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	data, err := io.ReadAll(io.TeeReader(f, h))
	if err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))

	lang := detectLang(rel)

	var fileID int64
	err = tx.QueryRow("SELECT id FROM files WHERE path = ?", rel).Scan(&fileID)
	if err == nil {
		tx.Exec("DELETE FROM symbols WHERE file_id = ?", fileID)
		tx.Exec("DELETE FROM imports WHERE file_id = ?", fileID)
		tx.Exec("DELETE FROM fts WHERE rowid = ?", fileID)
		tx.Exec("DELETE FROM files WHERE id = ?", fileID)
	}

	res, err := tx.Exec(
		"INSERT INTO files(path, content_hash, size, lang, mtime) VALUES(?,?,?,?,?)",
		rel, hash, info.Size(), lang, info.ModTime().Unix(),
	)
	if err != nil {
		return fmt.Errorf("insert file: %w", err)
	}
	fileID, _ = res.LastInsertId()

	content := string(data)
	if len(content) > 50000 {
		content = content[:50000]
	}
	if _, err := tx.Exec("INSERT INTO fts(rowid, path, content) VALUES(?,?,?)", fileID, rel, content); err != nil {
		return fmt.Errorf("insert fts: %w", err)
	}

	if lang == "go" {
		syms, imps := parseGo(data)
		for _, s := range syms {
			tx.Exec("INSERT INTO symbols(name, type, file_id, line, signature) VALUES(?,?,?,?,?)",
				s.Name, s.Type, fileID, s.Line, s.Signature)
		}
		for _, imp := range imps {
			tx.Exec("INSERT INTO imports(file_id, module) VALUES(?,?)", fileID, imp)
		}
	}

	return nil
}

// QuerySymbol searches symbols by name prefix (fast, indexed).
func (pm *ProjectMap) QuerySymbol(name string) ([]Symbol, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	rows, err := pm.db.Query(`
		SELECT s.name, s.type, f.path, s.line, s.signature
		FROM symbols s
		JOIN files f ON s.file_id = f.id
		WHERE s.name LIKE ? || '%'
		ORDER BY s.name
		LIMIT 50`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Symbol
	for rows.Next() {
		var s Symbol
		if err := rows.Scan(&s.Name, &s.Type, &s.File, &s.Line, &s.Signature); err == nil {
			out = append(out, s)
		}
	}
	return out, nil
}

// SearchText performs FTS5 full-text search over file contents.
func (pm *ProjectMap) SearchText(query string, limit int) ([]SearchResult, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	q := strings.ReplaceAll(query, `"`, `""`)
	q = `"` + q + `"`

	rows, err := pm.db.Query(`
		SELECT f.path, f.lang, f.size, snippet(fts, 2, '<b>', '</b>', '...', 32)
		FROM fts
		JOIN files f ON fts.rowid = f.id
		WHERE fts MATCH ?
		ORDER BY rank
		LIMIT ?`, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SearchResult
	for rows.Next() {
		var r SearchResult
		var snippet sql.NullString
		if err := rows.Scan(&r.Path, &r.Lang, &r.Size, &snippet); err == nil {
			r.Snippet = snippet.String
			out = append(out, r)
		}
	}
	return out, nil
}

// ListFilesByLang returns files filtered by extension/language.
func (pm *ProjectMap) ListFilesByLang(lang string, limit int) ([]string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	rows, err := pm.db.Query("SELECT path FROM files WHERE lang = ? ORDER BY path LIMIT ?", lang, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil {
			out = append(out, p)
		}
	}
	return out, nil
}

// Stats returns workspace statistics.
func (pm *ProjectMap) Stats() (map[string]interface{}, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := map[string]interface{}{"root": pm.root}

	queries := []struct {
		key string
		sql string
	}{
		{"total_files", "SELECT COUNT(*) FROM files"},
		{"total_symbols", "SELECT COUNT(*) FROM symbols"},
		{"total_imports", "SELECT COUNT(*) FROM imports"},
	}
	for _, q := range queries {
		rows, err := pm.db.Query(q.sql)
		if err != nil {
			continue
		}
		if rows.Next() {
			var v int
			rows.Scan(&v)
			stats[q.key] = v
		}
		rows.Close()
	}

	// Language breakdown
	rows, err := pm.db.Query("SELECT lang, COUNT(*) FROM files GROUP BY lang ORDER BY COUNT(*) DESC")
	if err == nil {
		langs := make([]map[string]interface{}, 0)
		for rows.Next() {
			var l string
			var c int
			if rows.Scan(&l, &c) == nil {
				langs = append(langs, map[string]interface{}{"lang": l, "count": c})
			}
		}
		stats["languages"] = langs
		rows.Close()
	}

	return stats, nil
}

// executor abstracts *sql.DB and *sql.Tx.
type executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Symbol represents a code symbol found in the workspace.
type Symbol struct {
	Name      string
	Type      string // func, type, var, const, interface, method
	File      string
	Line      int
	Signature string
}

// SearchResult is a single FTS5 match.
type SearchResult struct {
	Path    string
	Lang    string
	Size    int
	Snippet string
}

// ---- helpers ---------------------------------------------------------------

func shouldSkip(rel string, isDir bool) bool {
	parts := strings.Split(rel, string(filepath.Separator))
	for _, p := range parts {
		switch p {
		case ".git", ".github", ".mimic", "vendor", "node_modules", "bin", "dist", "build",
			".cache", ".opencode", ".vscode", "tmp", "temp":
			return true
		}
	}
	if !isDir {
		ext := strings.ToLower(filepath.Ext(rel))
		switch ext {
		case ".exe", ".dll", ".so", ".dylib", ".o", ".a", ".bin",
			".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg",
			".woff", ".woff2", ".ttf", ".eot",
			".zip", ".tar", ".gz", ".bz2", ".7z",
			".db", ".sqlite", ".sqlite3",
			".log", ".tmp":
			return true
		}
	}
	return false
}

func detectLang(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".py":
		return "python"
	case ".js", ".jsx", ".ts", ".tsx":
		return "javascript"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc":
		return "cpp"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".sh":
		return "shell"
	case ".md":
		return "markdown"
	case ".yml", ".yaml":
		return "yaml"
	case ".json":
		return "json"
	case ".toml":
		return "toml"
	case ".dockerfile", ".dockerignore":
		return "docker"
	default:
		base := strings.ToLower(filepath.Base(path))
		if base == "dockerfile" || base == "makefile" {
			return "docker"
		}
		if base == "makefile" {
			return "make"
		}
		return ""
	}
}

// ---- Linear Access Cache ---------------------------------------------------

// FileCacheEntry holds cached file content keyed by hash.
type FileCacheEntry struct {
	Hash    string
	Content []byte
}

var fileCache = make(map[string]FileCacheEntry)
var fileCacheMu sync.RWMutex

// ReadCached reads file from disk or returns from cache if hash matches.
func ReadCached(path string) ([]byte, error) {
	fileCacheMu.RLock()
	ce, ok := fileCache[path]
	fileCacheMu.RUnlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(data)
	hash := fmt.Sprintf("%x", h)

	if ok && ce.Hash == hash {
		return ce.Content, nil
	}

	fileCacheMu.Lock()
	fileCache[path] = FileCacheEntry{Hash: hash, Content: data}
	fileCacheMu.Unlock()
	return data, nil
}

// InvalidateCache removes a path from cache after writes.
func InvalidateCache(path string) {
	fileCacheMu.Lock()
	delete(fileCache, path)
	fileCacheMu.Unlock()
}
