package sub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"yadro.com/course/categorizer/core"
)

const subSubject = "xkcd.db.added_comic"
const pubSubject = "xkcd.categorizer.categorized"

type pubPayload struct {
	ComicID  int
	Category string
}

type Subsciber struct {
	log          *slog.Logger
	nc           *nats.Conn
	subscription *nats.Subscription
	categorizer  core.Categorizer
	pub          core.Publisher
	wg           sync.WaitGroup
	workers      chan struct{}
	ctx          context.Context
	cancel       context.CancelFunc
}

func New(log *slog.Logger, url string, cat core.Categorizer, pub core.Publisher, maxWorkers int) (*Subsciber, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats on address: %s: %w", url, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Subsciber{
		log:         log,
		nc:          nc,
		categorizer: cat,
		pub:         pub,
		workers:     make(chan struct{}, maxWorkers),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (s *Subsciber) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	var err error
	s.subscription, err = s.nc.Subscribe(subSubject, func(msg *nats.Msg) {
		s.log.Debug("received message", "subj", subSubject)

		s.wg.Add(1)
		go s.processMessage(msg)
	})
	if err != nil {
		return fmt.Errorf("subscribe to %v: %w", subSubject, err)
	}
	s.log.Debug("started listening", "subj", subSubject)
	return nil
}

func (s *Subsciber) processMessage(msg *nats.Msg) {
	defer s.wg.Done()

	select {
	case s.workers <- struct{}{}:
		defer func() { <-s.workers }()
	case <-s.ctx.Done():
		s.log.Debug("context done, skipping message")
		return
	}

	var comic core.Comic
	if err := json.Unmarshal(msg.Data, &comic); err != nil {
		s.log.Error("unmarshal message payload", "err", err)
		return
	}
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	category, err := s.categorizer.Categorize(ctx, comic)
	if err != nil {
		s.log.Error("categorize comic", "error", err, "comic_id", comic.ID)
		return
	}

	payload := pubPayload{ComicID: comic.ID, Category: category}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error("marshal payload to publish", "error", err)
		return
	}

	err = s.pub.Publish(ctx, core.Message{Subject: pubSubject, Payload: payloadBytes})
	if err != nil {
		s.log.Error("publish message", "error", err)
		return
	}

	s.log.Debug("successfully processed comic", "comic_id", comic.ID, "category", category)
}

func (s *Subsciber) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	s.wg.Wait()

	if s.subscription != nil {
		if err := s.subscription.Unsubscribe(); err != nil {
			return fmt.Errorf("unsubscribe from %v: %w", subSubject, err)
		}
	}

	if s.nc != nil && !s.nc.IsClosed() {
		err := s.nc.Drain()
		if err != nil {
			return fmt.Errorf("nats drain: %w", err)
		}
		s.nc.Close()
	}

	return nil
}
