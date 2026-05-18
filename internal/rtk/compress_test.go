package rtk

import (
	"strings"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		input string
		want  ContentType
	}{
		{"package main\nimport \"fmt\"", ContentCode},
		{"func main() {}", ContentCode},
		{"def hello():", ContentCode},
		{"{\"key\":\"value\"}", ContentJSON},
		{"[{\"id\":1}]", ContentJSON},
		{"diff --git a/file.txt b/file.txt", ContentDiff},
		{"2024-01-01 INFO starting", ContentLog},
		{"[INFO] something happened", ContentLog},
		{"12:34:56 DEBUG", ContentLog},
		{"just plain text", ContentText},
	}

	for _, tt := range tests {
		got := DetectContentType(tt.input)
		if got != tt.want {
			t.Errorf("DetectContentType(%q) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestCompress_Ansi(t *testing.T) {
	input := "\x1b[31mred\x1b[0m and \x1b[32mgreen\x1b[0m"
	cfg := Config{StripAnsi: true}
	got := Compress(input, ContentText, cfg)
	if strings.Contains(got, "\x1b[") {
		t.Errorf("ANSI not stripped: %q", got)
	}
	if got != "red and green" {
		t.Errorf("got %q, want 'red and green'", got)
	}
}

func TestCompress_CollapseBlanks(t *testing.T) {
	input := "line1\n\n\n\nline2\n\n\nline3"
	cfg := Config{CollapseBlanks: true}
	got := Compress(input, ContentText, cfg)
	expected := "line1\n\nline2\n\nline3"
	if got != expected {
		t.Errorf("got:\n%q\nwant:\n%q", got, expected)
	}
}

func TestCompress_MaxLines(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	cfg := Config{MaxLines: 3, HeadTailSplit: false}
	got := Compress(input, ContentText, cfg)
	if strings.Count(got, "\n") > 4 { // 3 lines + indicator
		t.Errorf("too many lines: %q", got)
	}
	if !strings.Contains(got, "omitted") {
		t.Errorf("missing omission indicator: %q", got)
	}
}

func TestCompress_HeadTail(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5\nline6\nline7"
	cfg := Config{HeadTailSplit: true, HeadLines: 2, TailLines: 2, MaxLines: 4}
	got := Compress(input, ContentLog, cfg)
	if !strings.Contains(got, "line1") {
		t.Error("missing head")
	}
	if !strings.Contains(got, "line7") {
		t.Error("missing tail")
	}
	if !strings.Contains(got, "omitted") {
		t.Error("missing omission indicator")
	}
	t.Logf("head+tail result:\n%s", got)
}

func TestCompress_CodeStripComments(t *testing.T) {
	input := `package main
import "fmt"
// main entry point
func main() {
	fmt.Println("hello")
}
/* block comment */
func helper() int {
	return 42
}`
	cfg := Config{StripComments: true, StripBodies: true}
	got := Compress(input, ContentCode, cfg)

	if strings.Contains(got, "// main") {
		t.Error("line comment not stripped")
	}
	if strings.Contains(got, "block comment") {
		t.Error("block comment not stripped")
	}
	if strings.Contains(got, `fmt.Println`) {
		t.Error("func body not stripped")
	}
	if !strings.Contains(got, "func main()") {
		t.Error("func signature missing")
	}
	if !strings.Contains(got, "func helper()") {
		t.Error("helper signature missing")
	}
	t.Logf("compressed code:\n%s", got)
}

func TestCompress_JSON(t *testing.T) {
	input := "{\n  \"key\": \"value\",\n  \"arr\": [1, 2, 3]\n}"
	cfg := Config{MaxLines: 2}
	got := Compress(input, ContentJSON, cfg)
	// Should be compacted to single line
	if strings.Count(got, "\n") > 1 {
		t.Errorf("JSON not compacted: %q", got)
	}
	if !strings.Contains(got, `"key":"value"`) {
		t.Errorf("JSON content lost: %q", got)
	}
}

func TestCompress_Log(t *testing.T) {
	var lines []string
	for i := 0; i < 20; i++ {
		lines = append(lines, "2024-01-01 INFO log line"+string(rune('0'+i%10)))
	}
	input := strings.Join(lines, "\n")
	cfg := Config{HeadTailSplit: true, HeadLines: 5, TailLines: 5, MaxLines: 8}
	got := Compress(input, ContentLog, cfg)
	if !strings.Contains(got, " omitted ") {
		t.Error("log should have omission indicator")
	}
	// Should keep first 5 and last 5
	if !strings.Contains(got, "log line0") {
		t.Error("missing first log lines")
	}
	if !strings.Contains(got, "log line9") {
		t.Error("missing last log lines")
	}
}

func TestCompress_MaxChars(t *testing.T) {
	input := strings.Repeat("a", 2000)
	cfg := Config{MaxChars: 100}
	got := Compress(input, ContentText, cfg)
	if len(got) > 200 {
		t.Errorf("too long: %d chars", len(got))
	}
	if !strings.Contains(got, "omitted") {
		t.Error("missing chars omission indicator")
	}
}

func TestEstimateTokens(t *testing.T) {
	input := strings.Repeat("x", 400)
	got := EstimateTokens(input)
	if got != 100 {
		t.Errorf("EstimateTokens(400 chars) = %d, want 100", got)
	}
}

func TestCompress_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	input := "line1\nline2\nline3"
	got := Compress(input, ContentText, cfg)
	if got != input {
		t.Errorf("default config should not modify small input")
	}
}

func TestCompress_AggressiveConfig(t *testing.T) {
	cfg := AggressiveConfig()
	input := strings.Repeat("line\n", 100)
	got := Compress(input, ContentCode, cfg)
	if strings.Count(got, "\n") > cfg.MaxLines+5 {
		t.Errorf("aggressive config should reduce to ~%d lines, got %d", cfg.MaxLines, strings.Count(got, "\n"))
	}
}

func TestCompress_LargeCodeFile(t *testing.T) {
	// Simulate a large Go file
	var lines []string
	for i := 0; i < 300; i++ {
		lines = append(lines, "func Something"+string(rune('0'+i%10))+"()")
		lines = append(lines, "\t// do something")
		lines = append(lines, "\tfmt.Println(\"hello\")")
		lines = append(lines, "}")
	}
	input := strings.Join(lines, "\n")

	cfg := Config{
		MaxLines:       50,
		StripAnsi:      true,
		CollapseBlanks: true,
		StripComments:  true,
		StripBodies:    true,
	}
	got := Compress(input, ContentCode, cfg)
	tokensBefore := EstimateTokens(input)
	tokensAfter := EstimateTokens(got)
	reduction := 100.0 - float64(tokensAfter)*100.0/float64(tokensBefore)

	t.Logf("Before: %d lines, %d tokens", strings.Count(input, "\n"), tokensBefore)
	t.Logf("After: %d lines, %d tokens, reduction: %.1f%%", strings.Count(got, "\n"), tokensAfter, reduction)

	if reduction < 70 {
		t.Errorf("token reduction only %.1f%%, want >70%%", reduction)
	}
	if !strings.Contains(got, "func Something") {
		t.Error("signatures lost")
	}
	if strings.Contains(got, "fmt.Println") {
		t.Error("bodies not stripped")
	}
}
