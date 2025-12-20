package words

import (
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
)

func Normalize(text string) []string {
	normalized_set := make(map[string]bool)
	// var normalized []string
	if text == "" {
		return []string{}
	}
	cleaned := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, w := range cleaned {
		w = strings.ToLower(w)
		if english.IsStopWord(w) {
			continue
		}
		stemmed := english.Stem(w, false)
		normalized_set[stemmed] = true
		// normalized = append(normalized, stemmed)
	}
	normalized := slices.Collect(maps.Keys(normalized_set))
	return normalized
}
