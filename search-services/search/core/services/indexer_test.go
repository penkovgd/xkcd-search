package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/search/core"
	mock_core "yadro.com/course/search/core/mocks"
)

var (
	expectedIndex = map[string][]core.Comic{
		"hello": {c1, c3},
		"world": {c1},
		"comic": {c1, c2},
		"test":  {c2, c3},
		"go":    {c3},
	}
)

func newMocksIndexer(t *testing.T, ctrl *gomock.Controller, comics []core.Comic, returnErr error) *mock_core.MockStorage {
	t.Helper()

	mockStorage := mock_core.NewMockStorage(ctrl)

	if returnErr != nil {
		mockStorage.EXPECT().
			GetAll(gomock.Any()).
			Return(nil, returnErr).
			AnyTimes()
	} else {
		mockStorage.EXPECT().
			GetAll(gomock.Any()).
			Return(comics, nil).
			AnyTimes()
	}

	return mockStorage
}

func TestIndexer_Create_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)

	mockStorage := newMocksIndexer(t, ctrl, simpleStorageReturn, nil)
	indexer := NewIndexer(mockStorage)

	// Act
	err := indexer.Create(context.TODO())
	require.NoError(t, err)

	// Assert
	idx := indexer.Get()

	for word, expectedComics := range expectedIndex {
		actualComics, exists := idx[word]
		assert.True(t, exists)
		assert.ElementsMatch(t, expectedComics, actualComics)
	}
	assert.Equal(t, len(expectedIndex), len(idx))
}

func TestIndexer_Create_EmptyStorage(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)

	emptyComics := []core.Comic{}
	mockStorage := newMocksIndexer(t, ctrl, emptyComics, nil)
	indexer := NewIndexer(mockStorage)

	// Act
	err := indexer.Create(context.TODO())
	require.NoError(t, err)

	// Assert
	index := indexer.Get()
	assert.Empty(t, index)
}

func TestIndexer_Create_StorageError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)

	expectedErr := errors.New("storage error")
	mockStorage := newMocksIndexer(t, ctrl, nil, expectedErr)
	indexer := NewIndexer(mockStorage)

	// Act
	err := indexer.Create(context.TODO())

	// Assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "create index")
	assert.ErrorContains(t, err, "storage error")
}

func TestIndexer_Create_ReplacesOldIndex(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)

	firstComics := []core.Comic{
		{
			Id:    1,
			Url:   "https://xkcd.com/1/",
			Words: []string{"old", "data"},
		},
	}

	mockStorage1 := newMocksIndexer(t, ctrl, firstComics, nil)
	indexer := NewIndexer(mockStorage1)

	err := indexer.Create(context.TODO())
	require.NoError(t, err)

	// Check first index
	firstIndex := indexer.Get()
	assert.Contains(t, firstIndex, "old")
	assert.Contains(t, firstIndex, "data")
	assert.Len(t, firstIndex, 2)

	// Create new index
	ctrl2 := gomock.NewController(t)

	secondComics := []core.Comic{
		{
			Id:    2,
			Url:   "https://xkcd.com/2/",
			Words: []string{"new", "data"},
		},
	}

	mockStorage2 := newMocksIndexer(t, ctrl2, secondComics, nil)
	indexer.storage = mockStorage2

	err = indexer.Create(context.Background())
	require.NoError(t, err)

	secondIndex := indexer.Get()

	assert.NotContains(t, secondIndex, "old")

	assert.Contains(t, secondIndex, "new")
	assert.Contains(t, secondIndex, "data")

	dataComics := secondIndex["data"]
	assert.Len(t, dataComics, 1)
	assert.Equal(t, 2, dataComics[0].Id)
}

func TestIndexer_Get_ReturnsCopy(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)

	mockStorage := newMocksIndexer(t, ctrl, simpleStorageReturn, nil)
	indexer := NewIndexer(mockStorage)

	err := indexer.Create(context.TODO())
	require.NoError(t, err)

	// Act
	index1 := indexer.Get()
	index1["extra"] = []core.Comic{{Id: 999}}
	index2 := indexer.Get()

	// Assert
	assert.NotContains(t, index2, "extra")

	for word := range expectedIndex {
		assert.Contains(t, index2, word)
	}
}

func TestIndexer_Drop_ClearsIndex(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)

	mockStorage := newMocksIndexer(t, ctrl, simpleStorageReturn, nil)
	indexer := NewIndexer(mockStorage)

	err := indexer.Create(context.TODO())
	require.NoError(t, err)

	indexBefore := indexer.Get()
	assert.NotEmpty(t, indexBefore)

	// Act
	indexer.Drop()

	// Assert
	indexAfter := indexer.Get()
	assert.Empty(t, indexAfter)
}

func TestIndexer_ConcurrentAccess(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)

	mockStorage := newMocksIndexer(t, ctrl, simpleStorageReturn, nil)
	indexer := NewIndexer(mockStorage)

	err := indexer.Create(context.Background())
	require.NoError(t, err)

	// Act & Assert
	done := make(chan bool)

	// Read index goroutine
	go func() {
		for range 100 {
			_ = indexer.Get()
		}
		done <- true
	}()

	// Write index goroutine
	go func() {
		for range 10 {
			indexer.Drop()
			err := indexer.Create(context.TODO())
			require.NoError(t, err)
		}
		done <- true
	}()

	<-done
	<-done

	index := indexer.Get()
	assert.NotNil(t, index)
	for word := range expectedIndex {
		assert.Contains(t, index, word)
	}
}
