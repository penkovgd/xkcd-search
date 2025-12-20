package services

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"

	"yadro.com/course/search/core"
)

type ISearchService struct {
	log  *slog.Logger
	norm core.Normalizer
	idx  core.Indexer
}

func NewISearchService(log *slog.Logger, norm core.Normalizer, idx core.Indexer) (*ISearchService, error) {
	return &ISearchService{
		log:  log,
		norm: norm,
		idx:  idx,
	}, nil
}

func (s *ISearchService) SearchIndex(ctx context.Context, phrase string, limit int) ([]core.Comic, error) {
	if limit < 1 {
		limit = 10
	}
	if phrase == "" {
		return nil, fmt.Errorf("phrase cannot be empty: %w", core.ErrBadArguments)
	}

	keywords, err := s.norm.Norm(ctx, phrase)
	if err != nil {
		return nil, fmt.Errorf("normalize search phrase '%s': %w", phrase, err)
	}

	index := s.idx.Get()
	scores := make(map[int]float64)       // id -> score
	comicsMap := make(map[int]core.Comic) // id -> comic

	for _, kw := range keywords {
		comics, ok := index[kw]
		if !ok {
			continue
		}
		for _, c := range comics {
			scores[c.Id] += float64(countOccurrences(kw, c.Words))
			// scores[c.Id]++
			comicsMap[c.Id] = c
		}
	}

	// normalize scores
	for id := range scores {
		wordCount := len(comicsMap[id].Words)
		scores[id] /= math.Log(float64(wordCount) + 1)
	}

	type idScore struct {
		id    int
		score float64
	}
	var idScores []idScore
	for id, score := range scores {
		idScores = append(idScores, idScore{id, score})
	}
	sort.Slice(idScores, func(i, j int) bool {
		return idScores[i].score > idScores[j].score
	})

	if limit > len(idScores) {
		limit = len(idScores)
	}

	topComics := make([]core.Comic, 0, limit)
	for i := 0; i < limit; i++ {
		topComics = append(topComics, comicsMap[idScores[i].id])
	}

	return topComics, nil
}
