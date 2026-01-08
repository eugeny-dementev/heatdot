package icon

import (
	"bytes"
	"runtime"
	"testing"
)

func TestIconEncodingHeader(t *testing.T) {
	data, err := ForCount(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runtime.GOOS == "windows" {
		if len(data) < 4 {
			t.Fatalf("icon data too short: %d", len(data))
		}
		expected := []byte{0x00, 0x00, 0x01, 0x00}
		if !bytes.Equal(data[:4], expected) {
			t.Fatalf("expected ICO header %v, got %v", expected, data[:4])
		}
		return
	}

	if len(data) < 8 {
		t.Fatalf("icon data too short: %d", len(data))
	}
	expected := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if !bytes.Equal(data[:8], expected) {
		t.Fatalf("expected PNG header %v, got %v", expected, data[:8])
	}
}
