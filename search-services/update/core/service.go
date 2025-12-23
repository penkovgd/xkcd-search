package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	publisher   Publisher
	img         ImageStorage
	concurrency int64
	running     atomic.Bool
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, pub Publisher, img ImageStorage, concurrency int64,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		publisher:   pub,
		img:         img,
		concurrency: concurrency,
	}, nil
}

func difference(a, b []int) []int {
	bMap := make(map[int]struct{}, len(b))
	for _, x := range b {
		bMap[x] = struct{}{}
	}
	var diff []int
	for _, x := range a {
		if _, found := bMap[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func (s *Service) Update(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		return ErrAlreadyExists
	}
	defer s.running.Store(false)

	ids, err := s.db.IDs(ctx)
	if err != nil {
		s.log.Error("failed to get ids from db", "error", err)
		return err
	}
	lastId, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to get last id from xkcd", "error", err)
		return err
	}
	firstToLast := make([]int, lastId)
	for i := range lastId {
		firstToLast[i] = i + 1
	}
	diff := difference(firstToLast, ids)
	diff = difference(diff, []int{404}) // removing 404 id as it doesnt exist

	sem := semaphore.NewWeighted(s.concurrency)
	var wg sync.WaitGroup
	added := 0
	for _, id := range diff {
		wg.Go(func() {
			if err = sem.Acquire(ctx, 1); err != nil {
				s.log.Warn("failed to acquire semaphore, skipping", "id", id, "error", err)
				return
			}
			defer sem.Release(1)

			info, err := s.xkcd.Get(ctx, id)
			if err != nil {
				s.log.Warn("failed to fetch comic, skipping", "id", id, "error", err)
				return
			}

			words, err := s.words.Norm(ctx, fmt.Sprintf("%d %s %s %s %s",
				info.ID, info.Title, info.Description, info.Alt, info.SafeTitle))
			if err != nil {
				s.log.Warn("failed to get words, skipping", "id", id, "error", err)
				return
			}

			imageBytes, ext, err := s.xkcd.GetImage(ctx, info.URL)
			if err != nil {
				s.log.Warn("failed to fetch image, skipping", "id", id, "error", err)
				return
			}
			imageUrlLocal, err := s.img.Save(ctx, strconv.Itoa(info.ID)+ext, imageBytes)
			if err != nil {
				s.log.Warn("failed to save image, skipping", "id", id, "error", err)
				return
			}

			comic := Comic{ID: info.ID, URL: imageUrlLocal, Words: words, Title: info.SafeTitle, Date: info.Date}
			err = s.db.Add(ctx, comic)
			if err != nil {
				s.log.Warn("failed to save comic to db, skipping", "id", id, "error", err)
				return
			}

			payload, err := json.Marshal(comic)
			if err != nil {
				s.log.Warn("failed to marshal comic to json, skipping", "id", id, "error", err)
				return
			}
			err = s.publisher.Publish(ctx, Message{Subject: "xkcd.db.added_comic", Payload: payload})
			if err != nil {
				s.log.Warn("failed to publish xkcd.db.added_comic message, skipping", "id", id, "error", err)
				return
			}
			s.log.Debug("emitted message", "message", "xkcd.db.added_comic")

			added++
		})
	}
	wg.Wait()
	err = s.publisher.Publish(ctx, Message{Subject: "xkcd.db.updated"})
	if err != nil {
		s.log.Warn("publish message", "error", err)
	}
	s.log.Debug("emitted message", "message", "xkcd.db.updated", "added", added)
	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		s.log.Error("failed to get stats from db", "error", err)
		return ServiceStats{}, err
	}
	lastId, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to get last id from xkcd", "error", err)
		return ServiceStats{}, err
	}

	var comicsTotal int
	if lastId >= 404 {
		comicsTotal = lastId - 1 // comic 404 doesnt exists
	} else {
		comicsTotal = lastId
	}

	return ServiceStats{
		DBStats: DBStats{
			WordsTotal:    dbStats.WordsTotal,
			WordsUnique:   dbStats.WordsUnique,
			ComicsFetched: dbStats.ComicsFetched,
		},
		ComicsTotal: comicsTotal,
	}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	if s.running.Load() {
		return StatusRunning
	}
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	err := s.db.Drop(ctx)
	if err != nil {
		return fmt.Errorf("drop db: %w", err)
	}

	err = s.publisher.Publish(ctx, Message{Subject: "xkcd.db.dropped"})
	if err != nil {
		s.log.Warn("publish message:", "error", err)
	}

	return nil
}
