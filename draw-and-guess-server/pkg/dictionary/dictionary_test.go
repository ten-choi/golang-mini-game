package dictionary

import (
	"os"
	"testing"
)

func TestNormalizeJapanese(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"リンゴ", "りんご"},
		{"ゴリラ", "ごりら"},
		{"らーめん", "らーめ"},  // Remove trailing ん
		{"ラーメン", "らーめ"},  // Remove trailing ん
		{"コーヒー", "こーひー"}, // Keep trailing ー
		{"こーひー", "こーひー"}, // Keep trailing ー
		{"テレビ", "てれび"},
		{"さくらんぼ", "さくらんぼ"},
		{"りんごん", "りんご"},    // Remove trailing ん
		{"らーめんー", "らーめんー"}, // Keep trailing ー
	}

	for _, tt := range tests {
		result := normalizeJapanese(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeJapanese(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetFirstChar(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"りんご", "り"},
		{"リンゴ", "り"},
		{"ラーメン", "ら"},
		{"らーめん", "ら"},
		{"", ""},
	}

	for _, tt := range tests {
		result := GetFirstChar(tt.input)
		if result != tt.expected {
			t.Errorf("GetFirstChar(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetLastChar(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"りんご", "ご"},
		{"リンゴ", "ご"},
		{"ラーメン", "め"},  // Normalized to らーめ, last char before ん
		{"らーめんー", "ー"}, // Keep trailing ー as last char
		{"さくらん", "ら"},  // Skip trailing ん
		{"コーヒー", "ー"},  // Last char is ー
		{"", ""},
	}

	for _, tt := range tests {
		result := GetLastChar(tt.input)
		if result != tt.expected {
			t.Errorf("GetLastChar(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestValidateWordChain(t *testing.T) {
	tests := []struct {
		word1    string
		word2    string
		expected bool
	}{
		{"りんご", "ゴリラ", true}, // ご -> ご (normalized)
		{"ゴリラ", "ラッパ", true}, // ら -> ら
		{"ラッパ", "パンダ", true}, // ぱ -> ぱ
		{"りんご", "かき", false}, // ご -> か (mismatch)
		{"", "りんご", false},   // empty word1
		{"りんご", "", false},   // empty word2
	}

	for _, tt := range tests {
		result := ValidateWordChain(tt.word1, tt.word2)
		if result != tt.expected {
			t.Errorf("ValidateWordChain(%q, %q) = %v, want %v",
				tt.word1, tt.word2, result, tt.expected)
		}
	}
}

func TestDictionaryLoadAndSearch(t *testing.T) {
	// Create a temporary test dictionary file
	content := `# Test dictionary
りんご
ゴリラ
ラッパ
パンダ
`
	tmpfile, err := os.CreateTemp("", "test_dict_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Create a new dictionary instance for testing
	dict := &Dictionary{
		hashes: make([]uint64, 0),
		loaded: false,
	}

	// Load the dictionary
	if err := dict.LoadFromFile(tmpfile.Name()); err != nil {
		t.Fatalf("Failed to load dictionary: %v", err)
	}

	// Test word validation
	tests := []struct {
		word     string
		expected bool
	}{
		{"りんご", true},
		{"リンゴ", true}, // Katakana should match Hiragana
		{"ごりら", true},
		{"ゴリラ", true},
		{"らっぱ", true},
		{"ぱんだ", true},
		{"バナナ", false},    // Not in dictionary
		{"banana", false}, // Not Japanese
		{"", false},       // Empty
	}

	for _, tt := range tests {
		result := dict.IsValidWord(tt.word)
		if result != tt.expected {
			t.Errorf("IsValidWord(%q) = %v, want %v", tt.word, result, tt.expected)
		}
	}

	// Test loaded status
	if !dict.IsLoaded() {
		t.Error("Dictionary should be marked as loaded")
	}

	// Test word count
	wordCount := dict.GetWordCount()
	if wordCount != 4 {
		t.Errorf("GetWordCount() = %d, want 4", wordCount)
	}
}
