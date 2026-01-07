package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type JMdict struct {
	Words []Word `json:"words"`
}

type Word struct {
	Kana  []KanaEntry `json:"kana"`
	Sense []Sense     `json:"sense"`
}

type KanaEntry struct {
	Common bool   `json:"common"`
	Text   string `json:"text"`
}

type Sense struct {
	PartOfSpeech []string `json:"partOfSpeech"`
}

func main() {
	// Open JMdict file
	fmt.Println("Opening jmdict-eng-3.6.1.json...")
	file, err := os.Open("pkg/dictionary/jmdict-eng-3.6.1.json")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Decode JSON
	fmt.Println("Parsing JSON (this may take a while)...")
	decoder := json.NewDecoder(file)

	// Read opening bracket
	_, err = decoder.Token()
	if err != nil {
		fmt.Printf("Error reading JSON: %v\n", err)
		return
	}

	// Skip to "words" array
	for decoder.More() {
		t, err := decoder.Token()
		if err != nil {
			fmt.Printf("Error reading token: %v\n", err)
			return
		}

		if t == "words" {
			// Read opening bracket of words array
			_, err = decoder.Token()
			if err != nil {
				fmt.Printf("Error reading words array: %v\n", err)
				return
			}
			break
		}

		// Skip value
		var skip interface{}
		decoder.Decode(&skip)
	}

	wordSet := make(map[string]bool)
	count := 0

	fmt.Println("Extracting words...")

	// Read words one by one
	for decoder.More() {
		var word Word
		if err := decoder.Decode(&word); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error decoding word: %v\n", err)
			continue
		}

		count++
		if count%10000 == 0 {
			fmt.Printf("Processed %d words, extracted %d unique nouns\n", count, len(wordSet))
		}

		// Check if word is a noun (명사만 추출)
		if !isNoun(word) {
			continue
		}

		for _, kana := range word.Kana {
			// Clean up the word
			text := strings.TrimSpace(kana.Text)

			// Skip if empty
			if text == "" {
				continue
			}

			// Check if it's pure Japanese (hiragana, katakana, or ー)
			if !isJapaneseWord(text) {
				continue
			}

			// Filter by length (minimum 2 characters, no maximum)
			runeCount := len([]rune(text))
			if runeCount < 2 {
				continue
			}

			// Skip words ending with ん (can't continue in word chain)
			if strings.HasSuffix(text, "ん") {
				continue
			}

			// Add to set
			wordSet[text] = true
		}
	}

	fmt.Printf("\nTotal words processed: %d\n", count)
	fmt.Printf("Unique valid words extracted: %d\n", len(wordSet))

	// Convert to sorted slice
	words := make([]string, 0, len(wordSet))
	for word := range wordSet {
		words = append(words, word)
	}
	sort.Strings(words)

	// Write to file
	fmt.Println("\nWriting to japanese_words_jmdict.txt...")
	outFile, err := os.Create("data/japanese_words_jmdict.txt")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	// Write header
	outFile.WriteString("# Japanese Word Dictionary for Word Chain Game\n")
	outFile.WriteString("# Extracted from JMdict-eng-3.6.1.json\n")
	outFile.WriteString(fmt.Sprintf("# Total words: %d\n", len(words)))
	outFile.WriteString("# Criteria:\n")
	outFile.WriteString("#   - Pure hiragana/katakana words\n")
	outFile.WriteString("#   - Length: 2+ characters (no maximum)\n")
	outFile.WriteString("#   - Does not end with ん\n")
	outFile.WriteString("#   - Nouns only (명사만)\n")
	outFile.WriteString("#   - Sorted alphabetically\n\n")

	for _, word := range words {
		outFile.WriteString(word + "\n")
	}

	fmt.Printf("\n✅ Successfully created japanese_words_jmdict.txt with %d nouns!\n", len(words))

	// Show some statistics
	showStatistics(words)
}

// isNoun checks if a word is a noun
func isNoun(word Word) bool {
	if len(word.Sense) == 0 {
		return false
	}

	// Check if any sense has noun part of speech
	for _, sense := range word.Sense {
		for _, pos := range sense.PartOfSpeech {
			// Common noun types in JMdict
			if pos == "n" || // noun (common)
				pos == "n-adv" || // adverbial noun
				pos == "n-pref" || // noun, used as a prefix
				pos == "n-suf" || // noun, used as a suffix
				pos == "n-t" || // noun (temporal)
				pos == "n-pr" { // proper noun
				return true
			}
		}
	}

	return false
}

func isJapaneseWord(text string) bool {
	// Check if word contains only Japanese characters
	for _, r := range text {
		// Hiragana: 3040-309F
		// Katakana: 30A0-30FF
		// Long vowel mark: 30FC (ー)
		if !((r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) || // Katakana
			r == 0x30FC) { // ー
			return false
		}
	}
	return true
}

func showStatistics(words []string) {
	fmt.Println("\n📊 Statistics:")

	// Count by length
	lengthCount := make(map[int]int)
	hiraganaCount := 0
	katakanaCount := 0
	mixedCount := 0

	for _, word := range words {
		runes := []rune(word)
		lengthCount[len(runes)]++

		// Classify by script
		hasHiragana := false
		hasKatakana := false

		for _, r := range runes {
			if r >= 0x3040 && r <= 0x309F {
				hasHiragana = true
			} else if r >= 0x30A0 && r <= 0x30FF {
				hasKatakana = true
			}
		}

		if hasHiragana && !hasKatakana {
			hiraganaCount++
		} else if hasKatakana && !hasHiragana {
			katakanaCount++
		} else {
			mixedCount++
		}
	}

	fmt.Println("\nBy Length:")
	var lengths []int
	for length := range lengthCount {
		lengths = append(lengths, length)
	}
	sort.Ints(lengths)
	for _, length := range lengths {
		fmt.Printf("  %d characters: %d words\n", length, lengthCount[length])
	}

	fmt.Println("\nBy Script:")
	fmt.Printf("  Hiragana only: %d\n", hiraganaCount)
	fmt.Printf("  Katakana only: %d\n", katakanaCount)
	fmt.Printf("  Mixed: %d\n", mixedCount)

	// Show first 20 words
	fmt.Println("\nFirst 20 words:")
	for i, word := range words {
		if i >= 20 {
			break
		}
		fmt.Printf("  %s\n", word)
	}

	fmt.Println("\nLast 10 words:")
	start := len(words) - 10
	if start < 0 {
		start = 0
	}
	for i := start; i < len(words); i++ {
		fmt.Printf("  %s\n", words[i])
	}
}
