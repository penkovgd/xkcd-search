package eventInit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"yadro.com/course/search/core"
)

type Initiator struct {
	log       *slog.Logger
	indexer   core.Indexer
	nc        *nats.Conn
	subUpdate *nats.Subscription
	subDrop   *nats.Subscription
}

func New(log *slog.Logger, url string, indexer core.Indexer) (*Initiator, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats on address: %s: %w", url, err)
	}

	return &Initiator{
		log:     log,
		indexer: indexer,
		nc:      nc,
	}, nil
}

func (i *Initiator) Start(ctx context.Context) error {
	var err error
	i.subUpdate, err = i.nc.SubscribeSync("xkcd.db.updated")
	if err != nil {
		return fmt.Errorf("subscribe to update: %w", err)
	}
	i.subDrop, err = i.nc.SubscribeSync("xkcd.db.dropped")
	if err != nil {
		return fmt.Errorf("subscribe to drop: %w", err)
	}
	go i.listenUpdate(ctx)
	go i.listenDrop(ctx)

	return nil
}

func (i *Initiator) listenUpdate(ctx context.Context) {
	i.log.Debug("started listening for db update event")
	for {
		_, err := i.subUpdate.NextMsgWithContext(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				i.log.Debug("subscriber: listen context finished, exiting listen loop")
				return
			}
			i.log.Error("cannot get next message", "error", err)
			break
		}

		i.log.Debug("received db update event, reindexing...")
		err = i.indexer.Create(ctx)
		if err != nil {
			i.log.Warn("failed to reindex", "error", err)
			continue
		}
		i.log.Info("index successfully rebuilt")
	}
}

func (i *Initiator) listenDrop(ctx context.Context) {
	i.log.Debug("started listening for db drop event")
	for {
		_, err := i.subDrop.NextMsgWithContext(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				i.log.Debug("subscriber: listen context finished, exiting listen loop")
				return
			}
			i.log.Error("cannot get next message", "error", err)
			break
		}

		i.log.Debug("received db drop event, dropping index...")
		i.indexer.Drop()
		i.log.Info("index successfully dropped")
	}
}

func (i *Initiator) Close() error {
	if err := i.subUpdate.Unsubscribe(); err != nil {
		return fmt.Errorf("unsubsribe from update: %w", err)
	}

	if err := i.subDrop.Unsubscribe(); err != nil {
		return fmt.Errorf("unsubsribe from drop: %w", err)
	}

	if i.nc != nil && !i.nc.IsClosed() {
		err := i.nc.Drain()
		if err != nil {
			return fmt.Errorf("nats drain: %w", err)
		}
		i.nc.Close()
	}

	return nil
}
