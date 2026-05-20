package mesh

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TextSlot is a markdown-native mesh slot. Self-contained, human-readable,
// LLM-parseable. Replaces binary gob InvariantNode.
type TextSlot struct {
	ID          string
	Domain      string
	Invariant   string
	Context     string
	Actions     []string      // One per line, e.g. "SYS_FILE_READ path=/tmp"
	Links       []SlotLink    // Cross-domain edges
	Embed       [EmbedDim]int8
	Metadata    map[string]string
}

// SlotLink represents a cross-domain relationship.
type SlotLink struct {
	TargetID string
	Relation string  // used_in | similar_to | composes_with
	Weight   float64 // 0.0 .. 1.0
}

// ToMarkdown serializes a TextSlot to markdown.
func (s *TextSlot) ToMarkdown() []byte {
	var b strings.Builder
	b.WriteString("# Slot: " + s.ID + "\n")
	b.WriteString("## Domain\n" + s.Domain + "\n\n")
	b.WriteString("## Invariant\n" + s.Invariant + "\n\n")
	if s.Context != "" {
		b.WriteString("## Context\n" + s.Context + "\n\n")
	}
	if len(s.Actions) > 0 {
		b.WriteString("## Actions\n")
		for _, a := range s.Actions {
			b.WriteString("- " + a + "\n")
		}
		b.WriteString("\n")
	}
	if len(s.Links) > 0 {
		b.WriteString("## Cross-Domain Links\n")
		for _, l := range s.Links {
			b.WriteString(fmt.Sprintf("- [%s] %s %.2f\n", l.TargetID, l.Relation, l.Weight))
		}
		b.WriteString("\n")
	}
	// Embedding stored as base64 of raw int8 bytes
	b.WriteString("## Embedding\n")
	embedBytes := make([]byte, EmbedDim)
	for i, v := range s.Embed {
		embedBytes[i] = byte(v)
	}
	b.WriteString(base64.StdEncoding.EncodeToString(embedBytes) + "\n\n")
	if len(s.Metadata) > 0 {
		b.WriteString("## Metadata\n")
		keys := make([]string, 0, len(s.Metadata))
		for k := range s.Metadata {
			keys = append(keys, k)
		}
		for _, k := range keys {
			b.WriteString(fmt.Sprintf("- %s: %s\n", k, s.Metadata[k]))
		}
		b.WriteString("\n")
	}
	return []byte(b.String())
}

// ParseTextSlot parses a markdown text slot from raw bytes.
func ParseTextSlot(data []byte) (*TextSlot, error) {
	slot := &TextSlot{
		Metadata: make(map[string]string),
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var section string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimRight(line, " ")
		if strings.HasPrefix(line, "# Slot: ") {
			slot.ID = strings.TrimPrefix(line, "# Slot: ")
			continue
		}
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimPrefix(line, "## ")
			continue
		}
		switch section {
		case "Domain":
			if line != "" {
				slot.Domain = line
			}
		case "Invariant":
			if line != "" {
				if slot.Invariant != "" {
					slot.Invariant += "\n"
				}
				slot.Invariant += line
			}
		case "Context":
			if line != "" {
				if slot.Context != "" {
					slot.Context += "\n"
				}
				slot.Context += line
			}
		case "Actions":
			if strings.HasPrefix(line, "- ") {
				slot.Actions = append(slot.Actions, strings.TrimPrefix(line, "- "))
			}
		case "Cross-Domain Links":
			if strings.HasPrefix(line, "- [") {
				link, err := parseLinkLine(line)
				if err == nil {
					slot.Links = append(slot.Links, link)
				}
			}
		case "Embedding":
			if line != "" {
				embedBytes, err := base64.StdEncoding.DecodeString(line)
				if err == nil && len(embedBytes) == EmbedDim {
					for i := 0; i < EmbedDim; i++ {
						slot.Embed[i] = int8(embedBytes[i])
					}
				}
			}
		case "Metadata":
			if strings.HasPrefix(line, "- ") {
				parts := strings.SplitN(strings.TrimPrefix(line, "- "), ": ", 2)
				if len(parts) == 2 {
					slot.Metadata[parts[0]] = parts[1]
				}
			}
		}
	}
	return slot, scanner.Err()
}

func parseLinkLine(line string) (SlotLink, error) {
	// Format: "- [target-id] relation weight"
	line = strings.TrimPrefix(line, "- [")
	idx := strings.Index(line, "]")
	if idx < 0 {
		return SlotLink{}, fmt.Errorf("invalid link format")
	}
	target := line[:idx]
	rest := strings.TrimSpace(line[idx+1:])
	fields := strings.Fields(rest)
	if len(fields) < 1 {
		return SlotLink{}, fmt.Errorf("invalid link format")
	}
	relation := fields[0]
	weight := 1.0
	if len(fields) >= 2 {
		if w, err := strconv.ParseFloat(fields[1], 64); err == nil {
			weight = w
		}
	}
	return SlotLink{TargetID: target, Relation: relation, Weight: weight}, nil
}

// SaveToFile writes a TextSlot to a .md file.
func (s *TextSlot) SaveToFile(dir string) error {
	path := filepath.Join(dir, s.Domain, s.ID+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, s.ToMarkdown(), 0644)
}

// LoadTextSlotFromFile reads a single .md slot from disk.
func LoadTextSlotFromFile(path string) (*TextSlot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseTextSlot(data)
}

// LoadAllTextSlots scans a directory tree for .md files and loads them.
func LoadAllTextSlots(dir string) ([]*TextSlot, error) {
	var slots []*TextSlot
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		slot, err := LoadTextSlotFromFile(path)
		if err != nil {
			return err
		}
		slots = append(slots, slot)
		return nil
	})
	return slots, err
}
