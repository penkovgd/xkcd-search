package services

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"

	"yadro.com/course/search/core"
)

type SearchService struct {
	log     *slog.Logger
	storage core.Storage
	norm    core.Normalizer
}

func NewSearchService(log *slog.Logger, storage core.Storage, norm core.Normalizer) (*SearchService, error) {
	return &SearchService{
		log:     log,
		storage: storage,
		norm:    norm,
	}, nil
}

func (s *SearchService) Search(ctx context.Context, phrase string, limit int) ([]core.Comic, error) {
	if limit < 1 {
		limit = 10
	}
	if phrase == "" {
		return nil, fmt.Errorf("phrase cannot be empty: %w", core.ErrBadArguments)
	}

	normalized, err := s.norm.Norm(ctx, phrase)
	if err != nil {
		return nil, fmt.Errorf("normalize search phrase '%s': %w", phrase, err)
	}

	comics, err := s.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all comics: %w", err)
	}

	type scoredComic struct {
		comic core.Comic
		score float64
	}

	var scoredComics []scoredComic

	for _, comic := range comics {
		score := 0.0
		for _, kw := range normalized {
			score += float64(countOccurrences(kw, comic.Words))
		}

		if score > 0 {
			score /= math.Log(float64(len(comic.Words) + 1))
			scoredComics = append(scoredComics, scoredComic{comic, score})
		}
	}

	if len(scoredComics) == 0 {
		return []core.Comic{}, nil
	}

	sort.Slice(scoredComics, func(i, j int) bool {
		return scoredComics[i].score > scoredComics[j].score
	})

	if limit > len(scoredComics) {
		limit = len(scoredComics)
	}

	result := make([]core.Comic, limit)
	for i := 0; i < limit; i++ {
		result[i] = scoredComics[i].comic
	}

	return result, nil
}

func countOccurrences(keyword string, document []string) int {
	count := 0
	for _, w := range document {
		if keyword == w {
			count++
		}
	}
	return count
}
