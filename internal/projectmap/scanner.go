package projectmap

import (
	"regexp"
	"strings"
)

// parseGo extracts symbols and imports from Go source.
// Lightweight regex-based (no full AST — scales to 100K+ files).
func parseGo(data []byte) (symbols []Symbol, imports []string) {
	src := string(data)
	lines := strings.Split(src, "\n")

	// --- Imports ----------------------------------------------------------
	importBlock := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "import (") {
			importBlock = true
			continue
		}
		if importBlock && strings.HasPrefix(trim, ")") {
			importBlock = false
			continue
		}
		if importBlock {
			imp := extractImport(trim)
			if imp != "" {
				imports = append(imports, imp)
			}
		} else if strings.HasPrefix(trim, "import ") {
			imp := extractImport(trim[7:])
			if imp != "" {
				imports = append(imports, imp)
			}
		}
	}

	// --- Symbols ----------------------------------------------------------
	// Types: func, method, struct, interface, var, const
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		lineNo := i + 1

		// Function declarations
		if m := funcDeclRe.FindStringSubmatch(trim); m != nil {
			name := m[2]
			recv := m[1]
			sig := buildSignature(trim)
			typ := "func"
			if recv != "" {
				typ = "method"
			}
			symbols = append(symbols, Symbol{
				Name:      name,
				Type:      typ,
				Line:      lineNo,
				Signature: sig,
			})
			continue
		}

		// Type declarations
		if strings.HasPrefix(trim, "type ") {
			rest := strings.TrimSpace(trim[5:])
			if idx := strings.IndexAny(rest, " \t{"); idx > 0 {
				name := rest[:idx]
				typ := "type"
				if strings.Contains(trim, "interface{") || strings.Contains(trim, "interface {") {
					typ = "interface"
				} else if strings.Contains(trim, "struct{") || strings.Contains(trim, "struct {") {
					typ = "struct"
				}
				symbols = append(symbols, Symbol{Name: name, Type: typ, Line: lineNo})
			}
			continue
		}

		// Var / const blocks
		if strings.HasPrefix(trim, "var ") && !strings.HasPrefix(trim, "var (") {
			name := extractVarName(trim[4:])
			if name != "" {
				symbols = append(symbols, Symbol{Name: name, Type: "var", Line: lineNo})
			}
		} else if strings.HasPrefix(trim, "const ") && !strings.HasPrefix(trim, "const (") {
			name := extractVarName(trim[6:])
			if name != "" {
				symbols = append(symbols, Symbol{Name: name, Type: "const", Line: lineNo})
			}
		}
	}

	return symbols, imports
}

var funcDeclRe = regexp.MustCompile(`^(?:func\s+)?(?:\(([^)]+)\)\s+)?([A-Za-z_]\w*)\s*\(`)

func extractImport(s string) string {
	s = strings.TrimSpace(s)
	// Named import: "xyz "abc/def""
	if strings.Contains(s, `"`) {
		idx := strings.LastIndex(s, `"`)
		if idx > 0 {
			start := strings.LastIndex(s[:idx], `"`)
			if start >= 0 {
				return s[start+1 : idx]
			}
		}
	}
	// Dot or underscore import
	if idx := strings.IndexByte(s, '"'); idx >= 0 {
		end := strings.IndexByte(s[idx+1:], '"')
		if end >= 0 {
			return s[idx+1 : idx+1+end]
		}
	}
	return ""
}

func extractVarName(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexAny(s, " =:\t"); idx > 0 {
		return s[:idx]
	}
	return ""
}

func buildSignature(line string) string {
	// Truncate to just the signature part
	line = strings.TrimSpace(line)
	if idx := strings.IndexByte(line, '{'); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	return line
}

// IndexWorkspaceDir is a standalone helper that indexes a directory tree.
func IndexWorkspaceDir(root string) error {
	pm, err := OpenOrCreate(root)
	if err != nil {
		return err
	}
	defer pm.Close()
	return pm.IndexWorkspace()
}

// QuickStats returns basic stats without opening full DB.
func QuickStats(root string) (map[string]interface{}, error) {
	pm, err := OpenOrCreate(root)
	if err != nil {
		return nil, err
	}
	defer pm.Close()
	return pm.Stats()
}

// Search is a convenience wrapper.
func Search(root, query string, limit int) ([]SearchResult, error) {
	pm, err := OpenOrCreate(root)
	if err != nil {
		return nil, err
	}
	defer pm.Close()
	return pm.SearchText(query, limit)
}
