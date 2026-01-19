package dictionary

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Dictionary provides fast word validation for Japanese wordchain game
// Uses sorted uint64 hashes with binary search for memory efficiency
type Dictionary struct {
	hashes []uint64 // Sorted hash array for binary search O(log n)
	loaded bool
	mu     sync.RWMutex
}

var (
	instance *Dictionary
	once     sync.Once
)

// GetInstance returns singleton dictionary instance
func GetInstance() *Dictionary {
	once.Do(func() {
		instance = &Dictionary{}
	})
	return instance
}

// LoadFromFile loads dictionary from text file (one word per line)
func (d *Dictionary) LoadFromFile(filePath string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open dictionary file: %w", err)
	}
	defer file.Close()

	hashes := make([]uint64, 0, 200000)
	hashSet := make(map[uint64]bool, 200000) // Deduplicate during load

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" {
			continue
		}

		// Normalize: katakana → hiragana
		normalized := normalizeWord(word)
		hash := hashWord(normalized)

		// Skip duplicates
		if !hashSet[hash] {
			hashes = append(hashes, hash)
			hashSet[hash] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading dictionary: %w", err)
	}

	// Sort for binary search
	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i] < hashes[j]
	})

	d.hashes = hashes
	d.loaded = true

	return nil
}

// IsValidWord checks if a word exists in dictionary (O(log n) binary search)
func (d *Dictionary) IsValidWord(word string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.loaded {
		return false
	}

	normalized := normalizeWord(word)
	hash := hashWord(normalized)

	// Binary search in sorted hash array
	idx := sort.Search(len(d.hashes), func(i int) bool {
		return d.hashes[i] >= hash
	})

	return idx < len(d.hashes) && d.hashes[idx] == hash
}

// IsLoaded returns whether dictionary has been loaded
func (d *Dictionary) IsLoaded() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.loaded
}

// GetWordCount returns total number of words in dictionary
func (d *Dictionary) GetWordCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.hashes)
}

// GetRandomWord returns a random word from the dictionary
func (d *Dictionary) GetRandomWord() string {
	commonWords := []string{
		"あき", "いぬ", "うみ", "えび", "おか",
		"かぜ", "きつね", "くも", "けんか", "こころ",
		"さくら", "しま", "すいか", "せかい", "そら",
		"たいよう", "ちから", "つき", "てんき", "とけい",
		"なまえ", "にわ", "ぬま", "ねこ", "のりもの",
	}
	
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return commonWords[r.Intn(len(commonWords))]
}

// normalizeWord converts katakana to hiragana for consistent lookup
func normalizeWord(word string) string {
	runes := []rune(word)
	for i, r := range runes {
		// Katakana range: 0x30A0-0x30FF → Hiragana: 0x3040-0x309F
		if r >= 0x30A0 && r <= 0x30FF {
			runes[i] = r - 0x60 // Convert to hiragana
		}
	}
	return string(runes)
}

// hashWord creates FNV-1a 64-bit hash for fast lookup
func hashWord(word string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(word))
	return h.Sum64()
}
