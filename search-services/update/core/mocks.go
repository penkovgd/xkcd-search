package core

import (
	"context"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"
)

type FakeDB struct {
	Comics map[int]Comic
	Delay  time.Duration
	mu     sync.RWMutex
}

func NewFakeDB(comics map[int]Comic, delay time.Duration) *FakeDB {
	if comics == nil {
		comics = make(map[int]Comic)
	}
	return &FakeDB{
		Comics: comics,
		Delay:  delay,
	}
}

func (db *FakeDB) Add(_ context.Context, c Comic) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	time.Sleep(db.Delay)
	db.Comics[c.ID] = c
	return nil
}

func (db *FakeDB) Stats(_ context.Context) (DBStats, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	wordsTotal := 0
	wordsUnique := make(map[string]bool)
	for _, comic := range db.Comics {
		wordsTotal += len(comic.Words)
		for _, w := range comic.Words {
			wordsUnique[w] = true
		}
	}
	return DBStats{
		WordsTotal:    wordsTotal,
		WordsUnique:   len(wordsUnique),
		ComicsFetched: len(db.Comics),
	}, nil
}

func (db *FakeDB) Drop(_ context.Context) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.Comics = make(map[int]Comic)
	return nil
}

func (db *FakeDB) IDs(_ context.Context) ([]int, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return slices.Collect(maps.Keys(db.Comics)), nil
}

type FakeXKCD struct {
	comics map[int]XKCDInfo
}

func (xkcd *FakeXKCD) GetImage(_ context.Context, url string) ([]byte, string, error) {
	return []byte(url), "png", nil
}

func NewFakeXKCD(infos map[int]XKCDInfo) *FakeXKCD {
	if infos == nil {
		infos = make(map[int]XKCDInfo)
	}
	return &FakeXKCD{
		comics: infos,
	}
}

func (xkcd *FakeXKCD) Get(_ context.Context, id int) (XKCDInfo, error) {
	c, ok := xkcd.comics[id]
	if !ok {
		return XKCDInfo{}, ErrNotFound
	}
	return c, nil
}

func (xkcd *FakeXKCD) LastID(_ context.Context) (int, error) {
	maxId := 0
	for id := range xkcd.comics {
		maxId = max(id, maxId)
	}
	return maxId, nil
}

type FakeWords struct{}

func (w *FakeWords) Norm(_ context.Context, phase string) ([]string, error) {
	return strings.Fields(phase), nil
}

type PublisherSpy struct {
	published []Message
}

func NewPublisherSpy() *PublisherSpy {
	return &PublisherSpy{
		published: make([]Message, 0),
	}
}

func (p *PublisherSpy) Publish(_ context.Context, msg Message) error {
	p.published = append(p.published, msg)
	return nil
}

func (p *PublisherSpy) GetPublishedMessages() []Message {
	return p.published
}

func (p *PublisherSpy) ResetMessages() {
	p.published = []Message{}
}

type ImageStorageStub struct{}

func (img ImageStorageStub) Save(_ context.Context, _ string, _ []byte) (string, error) {
	return "url", nil
}
