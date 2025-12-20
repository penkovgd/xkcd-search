package words

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", []string{}},
		{"simple", "simple", []string{"simpl"}},
		{"sprintf", fmt.Sprintf("%d %s %s", 1, "two", "three"), []string{"1", "two", "three"}},
		{"sprintf emply", fmt.Sprintf("%d %s %s", 1, "", ""), []string{"1"}},
		{"stopwords", "I to and", []string{}},
		{"duplicates", "this is duplicate duplicate", []string{"duplic"}},
		{"punctuation", "'testing, with: punctuation!\".", []string{"test", "punctuat"}},
		{"numbers", "123 words cats0", []string{"123", "word", "cats0"}},
		{"hard", "Русский, trying:To   'coNnecTion' 000-this!beautiful", []string{"русский", "tri", "connect", "000", "beauti"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.ElementsMatch(t, c.want, Normalize(c.input))
		})
	}
}
