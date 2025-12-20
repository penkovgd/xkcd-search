package tickInit

import (
	"context"
	"log/slog"
	"time"

	"yadro.com/course/search/core"
)

// driver adapter, calls indexer
type Initiator struct {
	log     *slog.Logger
	ttl     time.Duration
	indexer core.Indexer
	ticker  *time.Ticker
}

func New(log *slog.Logger, ttl time.Duration, indexer core.Indexer) *Initiator {
	return &Initiator{
		log:     log,
		ttl:     ttl,
		indexer: indexer,
	}
}

func (i *Initiator) Start(ctx context.Context) {
	i.log.Debug("initiator starting", "TTL", i.ttl)

	if err := i.indexer.Create(ctx); err != nil {
		i.log.Error("init create index:", "error", err)
	}

	i.ticker = time.NewTicker(i.ttl)
	go i.run(ctx)
}

func (i *Initiator) run(ctx context.Context) {
	defer i.ticker.Stop()

	for {
		select {
		case <-i.ticker.C:
			if err := i.indexer.Create(ctx); err != nil {
				i.log.Error("recreate index:", "error", err)
			}
			i.log.Debug("recreated index")
		case <-ctx.Done():
			i.log.Debug("initiator stopped")
			return
		}
	}
}
