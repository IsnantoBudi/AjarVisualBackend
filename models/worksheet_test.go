package models

import (
	"testing"
)

func TestSoalList_ValueAndScan(t *testing.T) {
	// Create sample SoalList
	original := SoalList{
		{
			Pertanyaan:   "1 + 1 = ?",
			JawabanBenar: "2",
			Opsi:         []string{"1", "2", "3", "4"},
			ImagePrompt:  "An apple",
			ImageURL:     "http://example.com/apple.jpg",
		},
	}

	// Test Value() (serialization)
	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	bytes, ok := val.([]byte)
	if !ok {
		t.Fatalf("Value() did not return []byte")
	}

	// Test Scan() (deserialization)
	var scanned SoalList
	err = scanned.Scan(bytes)
	if err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(scanned) != 1 {
		t.Fatalf("Expected 1 Soal, got %d", len(scanned))
	}

	if scanned[0].Pertanyaan != original[0].Pertanyaan {
		t.Errorf("Expected pertanyaan '%s', got '%s'", original[0].Pertanyaan, scanned[0].Pertanyaan)
	}
}

func TestSoalList_ScanInvalidType(t *testing.T) {
	var scanned SoalList
	err := scanned.Scan("invalid string type")
	if err == nil {
		t.Error("Expected error when scanning non-[]byte type, got nil")
	}
}
