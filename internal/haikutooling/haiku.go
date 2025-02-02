package haikutooling

import (
	"slices"
	"sync"

	haikunator "github.com/atrox/haikunatorgo"
)

const alphaNumericSymbols = "0123456789abcdefghijklmnopqrstuvwxyz"

var (
	// Needed because haikunator uses rand.NewSource which is not thread-safe.
	// We only need to lock it where the upstream package calls rand.*, which in our case is Generate().
	mu        = &sync.Mutex{}
	generator = haikunator.New()

	muV2        = &sync.Mutex{}
	generatorV2 = haikunator.New()
)

func init() {
	generator.Adjectives = filter(generator.Adjectives, "broken", "limit", "throbbing")
	generator.Nouns = filter(generator.Nouns, "disk", "wood")
	generator.TokenLength = 8

	generatorV2.Adjectives = filter(generator.Adjectives, "broken", "limit", "throbbing")
	generatorV2.Nouns = filter(generator.Nouns, "disk", "wood")
	generatorV2.TokenLength = 6
	generatorV2.TokenChars = alphaNumericSymbols
}

func filter(vals []string, toRemove ...string) []string {
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		if !slices.Contains(toRemove, v) {
			out = append(out, v)
		}
	}
	return out
}

func Generate() string {
	mu.Lock()
	defer mu.Unlock()

	return generator.Haikunate()
}

func GenerateV2() string {
	muV2.Lock()
	defer muV2.Unlock()

	return generatorV2.Haikunate()
}
