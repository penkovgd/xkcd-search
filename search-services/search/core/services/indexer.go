package services

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"yadro.com/course/search/core"
)

type Indexer struct {
	mu      sync.RWMutex
	idx     map[string][]core.Comic
	storage core.Storage
}

func NewIndexer(s core.Storage) *Indexer {
	return &Indexer{
		idx:     make(map[string][]core.Comic),
		storage: s,
	}
}

func (i *Indexer) Get() map[string][]core.Comic {
	i.mu.RLock()
	defer i.mu.RUnlock()

	copy := make(map[string][]core.Comic)
	maps.Copy(copy, i.idx)
	return copy
}

func (i *Indexer) Create(ctx context.Context) error {
	comics, err := i.storage.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	newIdx := make(map[string][]core.Comic)
	addedWords := make(map[string]map[int]bool)

	for _, c := range comics {
		for _, word := range c.Words {
			if addedWords[word] == nil {
				addedWords[word] = make(map[int]bool)
			}
			if !addedWords[word][c.Id] {
				newIdx[word] = append(newIdx[word], c)
				addedWords[word][c.Id] = true
			}
		}
	}

	i.mu.Lock()
	defer i.mu.Unlock()
	i.idx = newIdx

	return nil
}

func (i *Indexer) Drop() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.idx = make(map[string][]core.Comic)
}
