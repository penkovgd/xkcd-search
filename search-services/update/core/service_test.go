package core

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDifference(t *testing.T) {
	cases := []struct {
		name string
		A    []int
		B    []int
		want []int
	}{
		{"empty", []int{}, []int{}, []int{}},
		{"simple", []int{1, 2, 3}, []int{1, 2}, []int{3}},
		{"B bigger than A", []int{1}, []int{1, 2}, []int{}},
		{"A is empty", []int{}, []int{1}, []int{}},
		{"B is empty", []int{1}, []int{}, []int{1}},
		{"A = B", []int{1, 2, 3}, []int{3, 2, 1}, []int{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.ElementsMatch(t, c.want, difference(c.A, c.B))
		})
	}
}

func newServiceWithMocks(t *testing.T, comics map[int]Comic, infos map[int]XKCDInfo, concurrency int64, dbDelay time.Duration) (*FakeDB, *PublisherSpy, *Service, error) {
	t.Helper()

	fakeDB := NewFakeDB(comics, dbDelay)
	publisherSpy := NewPublisherSpy()

	service, err := NewService(
		slog.Default(),
		fakeDB,
		NewFakeXKCD(infos),
		&FakeWords{},
		publisherSpy,
		concurrency,
	)

	t.Cleanup(func() {
		err := fakeDB.Drop(context.TODO())
		require.NoError(t, err)
		publisherSpy.ResetMessages()
	})

	return fakeDB, publisherSpy, service, err
}

func TestNewService_NegativeConcurrency(t *testing.T) {
	_, _, _, err := newServiceWithMocks(t, nil, nil, -1, 0)
	want := "wrong concurrency specified"
	require.Error(t, err, want)
}

func TestStats_WithEmptyDBAndXKCD(t *testing.T) {
	// Arrange
	_, _, service, err := newServiceWithMocks(t, nil, nil, 10, 0)
	require.NoError(t, err)
	want := ServiceStats{
		DBStats: DBStats{
			WordsTotal:    0,
			WordsUnique:   0,
			ComicsFetched: 0,
		},
		ComicsTotal: 0,
	}

	// Act
	stats, err := service.Stats(context.TODO())
	require.NoError(t, err)

	// Assert
	require.Equal(t, want, stats)
}

func TestStats_Simple(t *testing.T) {
	// Arrange
	comics := map[int]Comic{
		1: {
			ID:    1,
			URL:   "url",
			Words: []string{"1", "2", "3"},
		},
	}
	infos := map[int]XKCDInfo{
		1: {
			ID:          1,
			URL:         "1",
			Title:       "2",
			SafeTitle:   "3",
			Description: "",
			Alt:         "",
		},
	}
	_, _, service, err := newServiceWithMocks(t, comics, infos, 10, 0)
	require.NoError(t, err)
	want := ServiceStats{
		DBStats: DBStats{
			WordsTotal:    3,
			WordsUnique:   3,
			ComicsFetched: 1,
		},
		ComicsTotal: 1,
	}

	// Act
	stats, err := service.Stats(context.TODO())
	require.NoError(t, err)

	// Assert
	require.Equal(t, want, stats)
}

func TestUpdate_Success(t *testing.T) {
	// Arrange
	infos := map[int]XKCDInfo{
		1: {
			ID:          1,
			URL:         "url",
			Title:       "title",
			SafeTitle:   "",
			Description: "",
			Alt:         "",
		},
	}
	db, pub, service, err := newServiceWithMocks(t, nil, infos, 10, 0)
	require.NoError(t, err)
	want := map[int]Comic{
		1: {
			ID:    1,
			URL:   "url",
			Words: []string{"1", "title"},
		},
	}

	// Act
	err = service.Update(context.TODO())
	require.NoError(t, err)

	// Assert
	actual := db.Comics
	published := pub.GetPublishedMessages()
	assert.Equal(t, want, actual)
	require.Len(t, published, 1)
	assert.Contains(t, published[0].Subject, "update")
}

func TestStatus_WhenNotUpdating(t *testing.T) {
	_, _, service, err := newServiceWithMocks(t, nil, nil, 10, 0)
	require.NoError(t, err)
	want := StatusIdle

	status := service.Status(context.TODO())

	assert.Equal(t, want, status)
}

func TestStatus_WhenUpdating(t *testing.T) {
	// Arrange
	infos := map[int]XKCDInfo{
		1: {
			ID:          1,
			URL:         "url",
			Title:       "title",
			SafeTitle:   "",
			Description: "",
			Alt:         "",
		},
	}
	dbDelay := 10 * time.Millisecond
	_, _, service, err := newServiceWithMocks(t, nil, infos, 10, dbDelay)
	require.NoError(t, err)

	// Act
	go func() {
		_ = service.Update(context.TODO())
	}()
	assert.Eventually(t, func() bool {
		return service.Status(context.TODO()) == StatusRunning
	}, 100*time.Millisecond, 1*time.Millisecond,
		"status should become %s", StatusRunning)
}

func TestUpdate_Parrallel(t *testing.T) {
	// Arrange
	infos := map[int]XKCDInfo{
		1: {
			ID:    1,
			URL:   "url",
			Title: "title",
		},
	}
	dbDelay := 10 * time.Millisecond
	_, pub, service, err := newServiceWithMocks(t, nil, infos, 10, dbDelay)
	require.NoError(t, err)

	var wg sync.WaitGroup
	errors := make([]error, 2)

	// Act
	wg.Go(func() {
		errors[0] = service.Update(context.TODO())
	})
	time.Sleep(3 * time.Millisecond)
	wg.Go(func() {
		errors[1] = service.Update(context.TODO())
	})

	wg.Wait()
	published := pub.GetPublishedMessages()

	// Assert
	assert.NoError(t, errors[0], "first Update should succeed")
	assert.Error(t, errors[1], "second Update should fail")
	assert.ErrorIs(t, errors[1], ErrAlreadyExists)
	require.Len(t, published, 1)
	assert.Contains(t, published[0].Subject, "updated") // проверка данных
}

func TestDrop(t *testing.T) {
	// Arrange
	comics := map[int]Comic{
		1: {
			ID:    1,
			URL:   "url",
			Words: []string{"1", "2", "3"},
		},
	}
	db, pub, service, err := newServiceWithMocks(t, comics, nil, 10, 0)
	require.NoError(t, err)

	// Act
	err = service.Drop(context.TODO())
	require.NoError(t, err)
	published := pub.GetPublishedMessages()

	// Assert
	assert.Empty(t, db.Comics, "db must be empty after drop")
	require.Len(t, published, 1)
	assert.Contains(t, published[0].Subject, "drop")
}
