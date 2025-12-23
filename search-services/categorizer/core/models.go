package core

import (
	"maps"
	"slices"
)

type Comic struct {
	ID    int
	URL   string
	Words []string
	Title string
}

var CategoriesDesc map[string]string = map[string]string{
	"Science & Mathematics":               "Physics, chemistry, biology, astronomy, equations, and thought experiments",
	"Programming & IT":                    "Coding jokes, software development, internet culture, security, and algorithms.",
	"Romance & Relationships":             "Love, dating, marriage, and emotional dialogues, often featuring the 'stick figure' couple.",
	"Sarcasm & Social Commentary":         "Irony and criticism aimed at societal norms, politics, and everyday absurdities.",
	"Language & Linguistics":              "Puns, grammar, translation, etymology, and communication quirks.",
	"Everyday Life & Observational Humor": "Mundane situations from the office, home, or store, highlighting relatable frustrations.",
	"Philosophical & Existential":         "Jokes about the meaning of life, time, consciousness, ethics, and the universe.",
	"Pop Culture & Parodies":              "References and parodies of movies, TV shows, books, and video games.",
	"Geography & Maps":                    "Jokes about cartography, map projections, coordinates, and travel.",
	"Time & Space":                        "Concepts of time travel, relativity, history, and space-time paradoxes.",
	"Security & Privacy":                  "Topics like passwords, encryption, phishing, and online safety.",
	"Academic & Educational":              "Humor about university life, PhDs, publishing, conferences, and teaching.",
	"Meta-Humor & Visual Gags":            "Comics about the format of xkcd itself, drawing, or breaking the fourth wall.",
	"Dark Humor & Absurdity":              "Jokes with unexpected, morbid, or logically bizarre punchlines.",
	"Environment & Climate":               "Comics about climate change, energy, ecology, and environmental data.",
	"Other":                               "Category for comics which doesn't fit any other category.",
}

func GetCategories() []string {
	return slices.Collect(maps.Keys(CategoriesDesc))
}

type Message struct {
	Subject string
	Payload []byte
}
