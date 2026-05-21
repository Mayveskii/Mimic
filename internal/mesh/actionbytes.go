package mesh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// PatchEntry is a single file operation decoded from ActionBytes.
type PatchEntry struct {
	Path    string
	Content []byte
	OpType  string // "replace", "insert", "shell"
}

// DecodeActionBytes attempts multiple heuristics to decode ActionBytes.
func DecodeActionBytes(data []byte) ([]PatchEntry, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty ActionBytes")
	}

	// Check if it's valid UTF-8 JSON / array
	if data[0] == '[' || data[0] == '{' {
		return decodeJSON(data)
	}

	// Check if it's printable ASCII (shell command)
	isPrintable := true
	for i := 0; i < len(data) && i < 64; i++ {
		if data[i] < 32 && data[i] != '\n' && data[i] != '\t' {
			isPrintable = false
			break
		}
	}
	if isPrintable && len(data) < 4096 {
		// Treat as shell command
		return []PatchEntry{{Path: "", Content: data, OpType: "shell"}}, nil
	}

	// Try embryo binary patch format (!-delimited entries)
	entries, err := decodeEmbryoBinary(data)
	if err == nil && len(entries) > 0 {
		return entries, nil
	}

	// Fallback: treat as raw binary, return single entry
	return nil, fmt.Errorf("unknown ActionBytes format: first 16 bytes = %x", data[:min(16, len(data))])
}

// decodeJSON parses JSON array of commands or patch objects.
func decodeJSON(data []byte) ([]PatchEntry, error) {
	var cmds []string
	if err := json.Unmarshal(data, &cmds); err == nil {
		// Array of shell commands
		var entries []PatchEntry
		for _, cmd := range cmds {
			entries = append(entries, PatchEntry{Path: "", Content: []byte(cmd), OpType: "shell"})
		}
		return entries, nil
	}

	var objects []map[string]interface{}
	if err := json.Unmarshal(data, &objects); err == nil {
		var entries []PatchEntry
		for _, obj := range objects {
			path, _ := obj["path"].(string)
			content, _ := obj["content"].(string)
			opType, _ := obj["op"].(string)
			if opType == "" {
				opType = "replace"
			}
			entries = append(entries, PatchEntry{Path: path, Content: []byte(content), OpType: opType})
		}
		return entries, nil
	}

	return nil, fmt.Errorf("failed to parse JSON ActionBytes")
}

// decodeEmbryoBinary parses the embryo binary patch format.
// Format observed from etcd data:
//
//	Entry := '!' (0x21) + opcode(1) + path_len(1) + path(N) + '\0' + content(M)
//	where content is the rest until the next '!' or EOF.
func decodeEmbryoBinary(data []byte) ([]PatchEntry, error) {
	// Find all '!' positions
	var positions []int
	for i := 0; i < len(data); i++ {
		if data[i] == 0x21 {
			positions = append(positions, i)
		}
	}

	if len(positions) == 0 {
		return nil, fmt.Errorf("no '!' markers found")
	}

	var entries []PatchEntry
	for i := 0; i < len(positions); i++ {
		start := positions[i] + 1 // skip '!'
		end := len(data)
		if i+1 < len(positions) {
			end = positions[i+1]
		}

		segment := data[start:end]
		if len(segment) < 3 {
			continue
		}

		entry := parseSegment(segment)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no valid entries parsed from %d markers", len(positions))
	}
	return entries, nil
}

// parseSegment parses a single entry segment (between ! markers).
func parseSegment(seg []byte) *PatchEntry {
	// Skip any leading control bytes (opcode/flags)
	off := 0
	for off < len(seg) && seg[off] < 0x20 {
		off++
	}
	if off >= len(seg) {
		return nil
	}

	// Find null terminator — marks end of path
	nullIdx := bytes.IndexByte(seg[off:], 0x00)
	if nullIdx < 0 {
		// No null terminator — path goes to end
		return &PatchEntry{
			Path:    string(seg[off:]),
			Content: nil,
			OpType:  "replace",
		}
	}
	nullIdx += off

	path := string(seg[off:nullIdx])
	if path == "" || strings.Contains(path, "\x01") || len(path) > 2048 {
		return nil
	}

	// Content is everything after the null
	// Skip any leading length bytes (content_len)
	contentStart := nullIdx + 1
	for contentStart < len(seg) && seg[contentStart] < 0x20 {
		contentStart++
	}

	return &PatchEntry{
		Path:    path,
		Content: seg[contentStart:],
		OpType:  "replace",
	}
}

// FormatActionBytes returns a human-readable representation of decoded patches.
func FormatActionBytes(data []byte) (string, error) {
	entries, err := DecodeActionBytes(data)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for i, e := range entries {
		switch e.OpType {
		case "shell":
			b.WriteString(fmt.Sprintf("[%d] SHELL: %s\n", i+1, string(e.Content)))
		default:
			contentPreview := string(e.Content)
			if len(contentPreview) > 200 {
				contentPreview = contentPreview[:197] + "..."
			}
			b.WriteString(fmt.Sprintf("[%d] %s: %s\n    Content (%d bytes): %s\n",
				i+1, e.OpType, e.Path, len(e.Content), contentPreview))
		}
	}
	return b.String(), nil
}

// ToToolCalls converts decoded patches into MCP tool call representations.
func (e PatchEntry) ToToolCalls() []map[string]interface{} {
	var calls []map[string]interface{}

	switch e.OpType {
	case "shell":
		calls = append(calls, map[string]interface{}{
			"tool": "SYS_EXEC",
			"args": map[string]string{"cmd": string(e.Content)},
		})

	case "replace":
		if e.Path == "" {
			break
		}
		// For replace, we'd need oldString/newString
		// Since we only have new content, this requires reading the file first
		calls = append(calls, map[string]interface{}{
			"tool": "FILE_EDIT",
			"args": map[string]interface{}{
				"path":      e.Path,
				"oldString": "[AGENT: read file and determine what to replace]",
				"newString": string(e.Content),
			},
		})
	}

	return calls
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
