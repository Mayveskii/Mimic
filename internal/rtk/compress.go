// Package rtk implements output compression inspired by Mayveskii/rtk
// (RTK: CLI proxy for token reduction, 49K stars, Rust)
//
// Core behaviors extracted:
//   - strip_ansi: remove color/escape codes
//   - collapse_blank_lines: 3+ newlines → 1
//   - max_lines: configurable line limit
//   - head_tail: keep first N + last M lines (useful for logs)
//   - smart_truncate: "... X lines omitted ..." indicator
//   - language_aware: for code, strip bodies keep signatures
//
// Every tool output passes through Compress() before returning to model.
// This prevents context window exhaustion on large outputs.
package rtk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ContentType classifies output for compression strategy
type ContentType string

const (
	ContentCode ContentType = "code" // Go, Rust, Python, etc.
	ContentLog  ContentType = "log"  // Build logs, command output
	ContentJSON ContentType = "json" // Structured data
	ContentText ContentType = "text" // Generic text
	ContentDiff ContentType = "diff" // Git diff, patches
)

// Config controls compression aggressiveness
type Config struct {
	// Generic limits
	MaxLines       int // 0 = unlimited
	MaxChars       int // 0 = unlimited
	StripAnsi      bool
	CollapseBlanks bool // 3+ consecutive blank lines → 1
	HeadTailSplit  bool // keep head + tail instead of truncating
	HeadLines      int  // lines to keep at start (if HeadTailSplit)
	TailLines      int  // lines to keep at end (if HeadTailSplit)

	// Code-specific
	StripComments  bool // remove // and /* */ comments
	KeepSignatures bool // keep func signatures, strip bodies
	StripBodies    bool // keep imports + type defs, strip func bodies
}

// DefaultConfig returns balanced compression (suggested for most tools)
func DefaultConfig() Config {
	return Config{
		MaxLines:       200,
		MaxChars:       10000,
		StripAnsi:      true,
		CollapseBlanks: true,
		HeadTailSplit:  false,
		HeadLines:      100,
		TailLines:      50,
		StripComments:  false,
		KeepSignatures: false,
		StripBodies:    false,
	}
}

// AggressiveConfig for large outputs where only structure matters
func AggressiveConfig() Config {
	return Config{
		MaxLines:       50,
		MaxChars:       5000,
		StripAnsi:      true,
		CollapseBlanks: true,
		HeadTailSplit:  true,
		HeadLines:      25,
		TailLines:      15,
		StripComments:  true,
		KeepSignatures: true,
		StripBodies:    true,
	}
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// DetectContentType guesses the output type from first 100 chars
func DetectContentType(sample string) ContentType {
	trimmed := strings.TrimSpace(sample)
	if len(trimmed) == 0 {
		return ContentText
	}

	// JSON detection
	if (trimmed[0] == '{' || trimmed[0] == '[') && isJSON(trimmed) {
		return ContentJSON
	}

	// Diff detection
	if strings.HasPrefix(trimmed, "diff ") || strings.HasPrefix(trimmed, "--- ") ||
		strings.HasPrefix(trimmed, "index ") || strings.HasPrefix(trimmed, "@@ ") {
		return ContentDiff
	}

	// Log detection (timestamps, levels)
	if logPattern.MatchString(trimmed) {
		return ContentLog
	}

	// Code detection (imports, func, package, etc.)
	firstLine := strings.Split(trimmed, "\n")[0]
	if codePattern.MatchString(firstLine) {
		return ContentCode
	}

	return ContentText
}

var logPattern = regexp.MustCompile(`^\d{4}[-/]\d{2}[-/]\d{2}|^\[|^\d{2}:\d{2}:\d{2}|^(INFO|DEBUG|WARN|ERROR|TRACE|FATAL)`)
var codePattern = regexp.MustCompile(`^(package |import |func |def |class |#include |using namespace|public |private )`)

func isJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// Compress reduces output size while preserving semantic meaning
func Compress(output string, contentType ContentType, cfg Config) string {
	if output == "" {
		return ""
	}

	// Step 1: Strip ANSI
	if cfg.StripAnsi {
		output = ansiEscape.ReplaceAllString(output, "")
	}

	// Step 2: Collapse blank lines
	if cfg.CollapseBlanks {
		output = collapseBlankLines(output)
	}

	// Step 3: Content-type specific filtering
	switch contentType {
	case ContentJSON:
		output = compressJSON(output, cfg)
	case ContentCode:
		output = compressCode(output, cfg)
	case ContentDiff:
		output = compressDiff(output, cfg)
	case ContentLog:
		output = compressLog(output, cfg)
	}

	// Step 4: Line-based truncation
	output = truncateLines(output, cfg)

	// Step 5: Character limit
	output = truncateChars(output, cfg.MaxChars)

	return output
}

func collapseBlankLines(s string) string {
	var buf bytes.Buffer
	blankCount := 0
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == "" {
			blankCount++
			if blankCount <= 2 {
				buf.WriteString("\n")
			}
		} else {
			blankCount = 0
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(line)
		}
	}
	// Collapse adjacent newlines: replace sequences of 3+ with exactly 2
	result := buf.String()
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimRight(result, "\n")
}

func compressJSON(s string, cfg Config) string {
	// Compact JSON if it's large
	if cfg.MaxLines > 0 && strings.Count(s, "\n") > cfg.MaxLines {
		var js interface{}
		if json.Unmarshal([]byte(s), &js) == nil {
			compact, _ := json.Marshal(js)
			return string(compact)
		}
	}
	return s
}

func compressCode(s string, cfg Config) string {
	if !cfg.StripComments && !cfg.KeepSignatures && !cfg.StripBodies {
		return s
	}

	lines := strings.Split(s, "\n")
	var result []string
	inBlockComment := false
	braceDepth := 0
	inFuncBody := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track block comments
		if strings.Contains(trimmed, "/*") {
			inBlockComment = true
		}
		if strings.Contains(trimmed, "*/") {
			inBlockComment = false
			if cfg.StripComments {
				continue
			}
		}

		// Strip comments
		if cfg.StripComments {
			if inBlockComment {
				continue
			}
			if strings.HasPrefix(trimmed, "//") {
				continue
			}
			if idx := strings.Index(line, "//"); idx >= 0 {
				line = line[:idx]
			}
		}

		// Detect function start for body stripping
		if cfg.StripBodies || cfg.KeepSignatures {
			if regexFuncStart.MatchString(trimmed) {
				inFuncBody = true
				braceDepth = 0
				result = append(result, line) // keep signature line
				continue
			}
			if inFuncBody {
				braceDepth += strings.Count(line, "{")
				braceDepth -= strings.Count(line, "}")
				if braceDepth <= 0 && strings.Contains(line, "}") {
					inFuncBody = false
					if cfg.KeepSignatures {
						result = append(result, "}") // keep closing brace
					}
				}
				if cfg.StripBodies {
					continue // skip body lines
				}
			}
		}

		result = append(result, line)
		_ = i // unused in loop header
	}

	return strings.Join(result, "\n")
}

func compressDiff(s string, cfg Config) string {
	// For diffs, we can compress context lines (non +/- lines)
	if cfg.MaxLines > 0 && strings.Count(s, "\n") > cfg.MaxLines {
		lines := strings.Split(s, "\n")
		var result []string
		contextCount := 0
		for _, line := range lines {
			if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "@@") {
				result = append(result, line)
				contextCount = 0
			} else {
				contextCount++
				if contextCount <= 3 { // keep up to 3 context lines
					result = append(result, line)
				}
			}
		}
		return strings.Join(result, "\n")
	}
	return s
}

func compressLog(s string, cfg Config) string {
	// Head + tail for logs (keep beginning and end, truncate middle)
	if cfg.HeadTailSplit && cfg.MaxLines > 0 {
		return headTailTruncate(s, cfg.HeadLines, cfg.TailLines)
	}
	return s
}

var regexFuncStart = regexp.MustCompile(`^(func\s|def\s|fn\s|public\s|private\s|async\s)`)

func truncateLines(s string, cfg Config) string {
	if cfg.MaxLines <= 0 {
		return s
	}

	lines := strings.Split(s, "\n")
	if len(lines) <= cfg.MaxLines {
		return s
	}

	if cfg.HeadTailSplit {
		return headTailTruncate(s, cfg.HeadLines, cfg.TailLines)
	}

	// Simple truncation
	omitted := len(lines) - cfg.MaxLines
	result := lines[:cfg.MaxLines]
	return strings.Join(result, "\n") + fmt.Sprintf("\n\n... %d lines omitted ...", omitted)
}

func headTailTruncate(s string, head, tail int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= head+tail {
		return s
	}

	result := make([]string, 0, head+tail+3)
	result = append(result, lines[:head]...)
	result = append(result, "")
	result = append(result, fmt.Sprintf("... %d lines omitted ...", len(lines)-head-tail))
	result = append(result, "")
	result = append(result, lines[len(lines)-tail:]...)
	return strings.Join(result, "\n")
}

func truncateChars(s string, maxChars int) string {
	if maxChars <= 0 || len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + fmt.Sprintf("\n\n... %d chars omitted ...", len(s)-maxChars)
}

// EstimateTokens approximates token count (~4 chars/token for English/code)
func EstimateTokens(s string) int {
	return len(s) / 4
}
