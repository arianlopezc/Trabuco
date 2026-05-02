package sync

import (
	"strings"
	"testing"
)

// TestExtractManagedBlock_Present verifies we can pull the body out of a
// well-formed file with both markers in order.
func TestExtractManagedBlock_Present(t *testing.T) {
	content := "user lines\n" +
		managedBlockBegin + "\n" +
		"managed line one\n" +
		"managed line two\n" +
		managedBlockEnd + "\n" +
		"more user lines\n"
	got := extractManagedBlock(content)
	want := "managed line one\nmanaged line two\n"
	if got != want {
		t.Errorf("extracted body mismatch\n got: %q\nwant: %q", got, want)
	}
}

// TestExtractManagedBlock_MissingBegin and _MissingEnd ensure we don't
// salvage a partially-marked file — the caller treats that as "no
// managed block present".
func TestExtractManagedBlock_MissingBegin(t *testing.T) {
	content := "user line\n" +
		"managed line\n" +
		managedBlockEnd + "\n"
	if got := extractManagedBlock(content); got != "" {
		t.Errorf("expected empty body when begin marker missing, got %q", got)
	}
}

func TestExtractManagedBlock_MissingEnd(t *testing.T) {
	content := managedBlockBegin + "\n" +
		"managed line\n"
	if got := extractManagedBlock(content); got != "" {
		t.Errorf("expected empty body when end marker missing, got %q", got)
	}
}

func TestExtractManagedBlock_Empty(t *testing.T) {
	if got := extractManagedBlock(""); got != "" {
		t.Errorf("expected empty body from empty input, got %q", got)
	}
}

// TestApplyManagedBlock_AppendsToFileWithoutMarkers covers the common
// upgrade path: an old project's .gitignore has no Trabuco markers, sync
// must append a fresh block at the end without touching existing rules.
func TestApplyManagedBlock_AppendsToFileWithoutMarkers(t *testing.T) {
	existing := "target/\n*.iml\n"
	body := ".ai/security-audit/findings.md\n"

	got := applyManagedBlock(existing, body)

	if !strings.HasPrefix(got, existing) {
		t.Errorf("user content must be preserved verbatim at start of file\ngot: %q", got)
	}
	if !strings.Contains(got, managedBlockBegin) || !strings.Contains(got, managedBlockEnd) {
		t.Error("appended output must contain both markers")
	}
	if !strings.Contains(got, body) {
		t.Errorf("appended output must contain expected body\ngot: %q", got)
	}
}

// TestApplyManagedBlock_ReplacesExistingBlock covers the second-sync
// path: markers are already in place from a prior sync run; the body
// between them must be regenerated to whatever the current CLI produces,
// and content outside the markers must be left exactly as-is.
func TestApplyManagedBlock_ReplacesExistingBlock(t *testing.T) {
	prefix := "user line A\nuser line B\n"
	suffix := "user line C\nuser line D\n"
	stale := "stale managed line\n"
	existing := prefix +
		managedBlockBegin + "\n" +
		stale +
		managedBlockEnd + "\n" +
		suffix

	body := "fresh managed line\n"
	got := applyManagedBlock(existing, body)

	if strings.Contains(got, "stale managed line") {
		t.Errorf("stale managed content must be removed\ngot: %q", got)
	}
	if !strings.Contains(got, "fresh managed line") {
		t.Errorf("fresh managed content must be present\ngot: %q", got)
	}
	if !strings.HasPrefix(got, prefix) {
		t.Errorf("preceding user content must be preserved verbatim\ngot: %q", got)
	}
	if !strings.HasSuffix(got, suffix) {
		t.Errorf("trailing user content must be preserved verbatim\ngot: %q", got)
	}
}

// TestApplyManagedBlock_Idempotent: applying the same block twice is a
// no-op. This is the round-trip guarantee — fresh-init projects must
// not show drift on a follow-up sync.
func TestApplyManagedBlock_Idempotent(t *testing.T) {
	existing := "user\n"
	body := "managed\n"
	first := applyManagedBlock(existing, body)
	second := applyManagedBlock(first, body)
	if first != second {
		t.Errorf("apply must be idempotent\nfirst:  %q\nsecond: %q", first, second)
	}
}

// TestApplyManagedBlock_HandlesMissingTrailingNewline normalizes a body
// that doesn't end in newline so the closing marker always starts on
// its own line.
func TestApplyManagedBlock_HandlesMissingTrailingNewline(t *testing.T) {
	got := applyManagedBlock("", "no newline at end")
	wantTail := "no newline at end\n" + managedBlockEnd + "\n"
	if !strings.HasSuffix(got, wantTail) {
		t.Errorf("body without trailing newline must be normalized\ngot: %q", got)
	}
}

// TestApplyManagedBlock_DamagedExistingBlock: only one marker present in
// existing content. Defensive behavior: ignore the dangling marker and
// append a fresh block at the end.
func TestApplyManagedBlock_DamagedExistingBlock(t *testing.T) {
	existing := "user A\n" + managedBlockBegin + "\norphan\n"
	body := "managed\n"
	got := applyManagedBlock(existing, body)

	if !strings.Contains(got, "orphan") {
		t.Error("damaged block content should be left in place — user can clean up by hand")
	}
	// And a fresh, complete block is appended.
	count := strings.Count(got, managedBlockBegin)
	if count != 2 {
		t.Errorf("expected 2 begin markers (orphan + appended block), got %d in %q", count, got)
	}
	if !strings.Contains(got, managedBlockEnd) {
		t.Error("appended block must close with end marker")
	}
}

// TestApplyManagedBlock_NoConsecutiveBlanks ensures we don't accumulate
// blank lines between user content and the appended block on repeated
// runs over a file that already has trailing blank lines.
func TestApplyManagedBlock_NoConsecutiveBlanks(t *testing.T) {
	existing := "user\n\n"
	body := "managed\n"
	got := applyManagedBlock(existing, body)
	if strings.Contains(got, "\n\n\n") {
		t.Errorf("must not produce 3+ consecutive newlines\ngot: %q", got)
	}
}

// TestApplyManagedBlock_PreservesCRLF: a Windows-edited .gitignore with
// CRLF line endings must round-trip through splice without converting
// to LF (which would produce false drift on every subsequent sync).
func TestApplyManagedBlock_PreservesCRLF(t *testing.T) {
	existing := "user line\r\n"
	body := "managed line\n"

	got := applyManagedBlock(existing, body)

	if strings.Contains(got, "\r\n\r\n\r\n") {
		t.Errorf("must not produce 3+ consecutive CRLF\ngot: %q", got)
	}
	if strings.Count(got, "\r\n") < 3 {
		t.Errorf("expected output to preserve CRLF endings\ngot: %q", got)
	}
	// Critically: idempotence under CRLF — second apply produces no diff.
	second := applyManagedBlock(got, body)
	if got != second {
		t.Errorf("CRLF apply must be idempotent\nfirst:  %q\nsecond: %q", got, second)
	}
}

// TestExtractManagedBlock_NormalizesCRLF guards the contract that
// extract returns LF-normalized bodies. The expected file (always LF on
// disk) and a CRLF-edited project file must yield identical bodies.
func TestExtractManagedBlock_NormalizesCRLF(t *testing.T) {
	crlf := managedBlockBegin + "\r\n" + "line one\r\n" + "line two\r\n" + managedBlockEnd + "\r\n"
	got := extractManagedBlock(crlf)
	want := "line one\nline two\n"
	if got != want {
		t.Errorf("extract must LF-normalize CRLF body\n got: %q\nwant: %q", got, want)
	}
}
