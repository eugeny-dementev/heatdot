package github

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseCountSVG(t *testing.T) {
	html := readFixture(t, "svg.html")
	date := time.Date(2024, 9, 5, 12, 0, 0, 0, time.Local)

	count, err := ParseCountForDate(html, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestParseCountAria(t *testing.T) {
	html := readFixture(t, "aria.html")
	date := time.Date(2024, 9, 5, 9, 0, 0, 0, time.Local)

	count, err := ParseCountForDate(html, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4, got %d", count)
	}
}

func TestParseCountAriaNoContrib(t *testing.T) {
	html := readFixture(t, "aria.html")
	date := time.Date(2024, 9, 6, 9, 0, 0, 0, time.Local)

	count, err := ParseCountForDate(html, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return data
}
