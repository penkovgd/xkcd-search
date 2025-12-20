package core

import "context"

type Searcher interface {
	Search(ctx context.Context, phrase string, limit int) ([]Comic, error)
}
type ISearcher interface {
	SearchIndex(ctx context.Context, phrase string, limit int) ([]Comic, error)
}

type Normalizer interface {
	Norm(context.Context, string) ([]string, error)
}

type Storage interface {
	GetAll(context.Context) ([]Comic, error)
}

type Indexer interface {
	Create(context.Context) error
	Get() map[string][]Comic
	Drop()
}
