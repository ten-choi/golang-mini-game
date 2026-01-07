package dictionary

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strings"
	"sync"
)

// Dictionary represents an in-memory word dictionary using hash-based search
type Dictionary struct {
	hashes []uint64
	mu     sync.RWMutex
	loaded bool
}

var (
	instance *Dictionary
	once     sync.Once
)

// GetInstance returns the singleton dictionary instance
func GetInstance() *Dictionary {
	once.Do(func() {
		instance = &Dictionary{
			hashes: make([]uint64, 0),
			loaded: false,
		}
	})
	return instance
}

// LoadFromFile loads words from a file and builds the hash index
func (d *Dictionary) LoadFromFile(filepath string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open dictionary file: %w", err)
	}
	defer file.Close()

	wordSet := make(map[uint64]struct{})
	scanner := bufio.NewScanner(file)
	lineCount := 0

	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Normalize and hash the word
		normalized := normalizeJapanese(line)
		if normalized != "" {
			hash := hashWord(normalized)
			wordSet[hash] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading dictionary file: %w", err)
	}

	// Convert map to sorted slice for binary search
	d.hashes = make([]uint64, 0, len(wordSet))
	for hash := range wordSet {
		d.hashes = append(d.hashes, hash)
	}
	sort.Slice(d.hashes, func(i, j int) bool {
		return d.hashes[i] < d.hashes[j]
	})

	d.loaded = true
	fmt.Printf("Dictionary loaded: %d unique words from %d lines\n", len(d.hashes), lineCount)
	return nil
}

// IsValidWord checks if a word exists in the dictionary
func (d *Dictionary) IsValidWord(word string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.loaded {
		return false
	}

	normalized := normalizeJapanese(word)
	if normalized == "" {
		return false
	}

	hash := hashWord(normalized)

	// Binary search
	idx := sort.Search(len(d.hashes), func(i int) bool {
		return d.hashes[i] >= hash
	})

	return idx < len(d.hashes) && d.hashes[idx] == hash
}

// IsLoaded returns whether the dictionary has been successfully loaded
func (d *Dictionary) IsLoaded() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.loaded
}

// GetWordCount returns the number of words in the dictionary
func (d *Dictionary) GetWordCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.hashes)
}

// normalizeJapanese converts Katakana to Hiragana and removes special characters
func normalizeJapanese(word string) string {
	var normalized strings.Builder

	for _, r := range word {
		// Convert Katakana to Hiragana
		if r >= 0x30A1 && r <= 0x30F6 { // Katakana range
			r = r - 0x0060 // Convert to Hiragana
		}

		// Keep only Hiragana and some special characters
		if (r >= 0x3041 && r <= 0x3096) || // Hiragana
			r == 'ー' || // Long vowel mark
			r == 'っ' || r == 'ッ' { // Small tsu
			normalized.WriteRune(r)
		}
	}

	result := normalized.String()

	// Remove only trailing 'ん' as it typically can't start words
	// Keep 'ー' in the middle and at the end
	result = strings.TrimRight(result, "ん")

	return result
}

// hashWord creates a 64-bit hash from a normalized word
func hashWord(word string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(word))
	return h.Sum64()
}

// GetFirstChar returns the first character of a word (for wordchain matching)
func GetFirstChar(word string) string {
	normalized := normalizeJapanese(word)
	if len(normalized) == 0 {
		return ""
	}

	runes := []rune(normalized)
	if len(runes) == 0 {
		return ""
	}

	return string(runes[0])
}

// GetLastChar returns the last character of a word (for wordchain matching)
func GetLastChar(word string) string {
	normalized := normalizeJapanese(word)
	if len(normalized) == 0 {
		return ""
	}

	runes := []rune(normalized)
	if len(runes) == 0 {
		return ""
	}

	// Find the last valid character (skip only trailing 'ん')
	for i := len(runes) - 1; i >= 0; i-- {
		char := runes[i]
		if char != 'ん' {
			return string(char)
		}
	}

	// If all characters are 'ん', return the last one
	return string(runes[len(runes)-1])
}

// IsHiragana checks if a rune is Hiragana
func IsHiragana(r rune) bool {
	return r >= 0x3041 && r <= 0x3096
}

// IsKatakana checks if a rune is Katakana
func IsKatakana(r rune) bool {
	return r >= 0x30A1 && r <= 0x30F6
}

// IsJapanese checks if a string contains Japanese characters
func IsJapanese(word string) bool {
	for _, r := range word {
		if IsHiragana(r) || IsKatakana(r) {
			return true
		}
	}
	return false
}

// ValidateWordChain checks if word2 can follow word1 in wordchain
func ValidateWordChain(word1, word2 string) bool {
	if word1 == "" || word2 == "" {
		return false
	}

	lastChar := GetLastChar(word1)
	firstChar := GetFirstChar(word2)

	if lastChar == "" || firstChar == "" {
		return false
	}

	return lastChar == firstChar
}
