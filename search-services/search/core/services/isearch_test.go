package services

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/search/core"
	mock_core "yadro.com/course/search/core/mocks"
)

var (
	c1 = core.Comic{
		Id:    1,
		Url:   "https://xkcd.com/1/",
		Words: []string{"hello", "world", "comic"},
	}
	c2 = core.Comic{
		Id:    2,
		Url:   "https://xkcd.com/2/",
		Words: []string{"test", "comic"},
	}
	c3 = core.Comic{
		Id:    3,
		Url:   "https://xkcd.com/3/",
		Words: []string{"go", "test", "hello"},
	}

	simpleIndex = map[string][]core.Comic{
		"hello": {c1, c3},
		"world": {c1},
		"comic": {c1, c2},
		"test":  {c2, c3},
		"go":    {c3},
	}
)

func newSimpleMocksISearch(t *testing.T, ctrl *gomock.Controller, index map[string][]core.Comic) (*mock_core.MockIndexer, *mock_core.MockNormalizer) {
	t.Helper()

	mockIndexerSimple := mock_core.NewMockIndexer(ctrl)
	mockIndexerSimple.
		EXPECT().
		Get().
		Return(index).
		AnyTimes()
	mockNormSimple := mock_core.NewMockNormalizer(ctrl)
	mockNormSimple.
		EXPECT().
		Norm(context.TODO(), gomock.Any()).
		DoAndReturn(func(_ context.Context, phrase string) ([]string, error) {
			return strings.Fields(phrase), nil
		}).
		AnyTimes()

	return mockIndexerSimple, mockNormSimple
}

func TestISearch_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	mockIndexer, mockNorm := newSimpleMocksISearch(t, ctrl, simpleIndex)

	tests := []struct {
		name   string
		phrase string
		limit  int
		want   any
	}{
		{
			name:   "simple",
			phrase: "go",
			limit:  1,
			want:   []core.Comic{c3},
		},
		{
			name:   "simple 2",
			phrase: "comic",
			limit:  1,
			want:   []core.Comic{c2},
		},
		{
			name:   "russian",
			phrase: "русские хакеры hello world",
			limit:  1,
			want:   []core.Comic{c1},
		},
		{
			name:   "no comics found",
			phrase: "no relevant comics for this: dlsfkjlsvsdflksdf",
			limit:  10,
			want:   []core.Comic{},
		},
		{
			name:   "relevant comics less than limit",
			phrase: "comic",
			limit:  10,
			want:   []core.Comic{c1, c2},
		},
		{
			name:   "relevant comics more than limit - return most scored",
			phrase: "hello world comic",
			limit:  1,
			want:   []core.Comic{c1},
		},
	}

	service, err := NewISearchService(slog.Default(), mockNorm, mockIndexer)
	require.NoError(t, err)

	// Act
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.SearchIndex(context.TODO(), tt.phrase, tt.limit)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestISearch_BadLimit(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	mockIndexer, mockNorm := newSimpleMocksISearch(t, ctrl, simpleIndex)

	phrase := "hello test go"
	want := []core.Comic{c1, c2, c3}

	tests := []struct {
		name  string
		limit int
		want  []core.Comic
	}{
		{
			name:  "zero",
			limit: 0,
			want:  want,
		},
		{
			name:  "negative",
			limit: -1,
			want:  want,
		},
		{
			name:  "more than search has",
			limit: 10_000,
			want:  want,
		},
	}

	service, err := NewISearchService(slog.Default(), mockNorm, mockIndexer)
	require.NoError(t, err)

	// Act
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.SearchIndex(context.TODO(), phrase, tt.limit)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestISearch_EmptyPhrase(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	mockIndexer, mockNorm := newSimpleMocksISearch(t, ctrl, simpleIndex)

	phrase := ""
	limit := 10

	service, err := NewISearchService(slog.Default(), mockNorm, mockIndexer)
	require.NoError(t, err)

	// Act
	got, err := service.SearchIndex(context.TODO(), phrase, limit)
	// Assert
	assert.ErrorIs(t, err, core.ErrBadArguments)
	assert.Empty(t, got)
}

func TestISearch_NoComicsInIndex(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	mockIndexer, mockNorm := newSimpleMocksISearch(t, ctrl, map[string][]core.Comic{})
	service, err := NewISearchService(slog.Default(), mockNorm, mockIndexer)
	require.NoError(t, err)

	phrase := "doesnt matter"
	limit := 10

	// Act
	got, err := service.SearchIndex(context.TODO(), phrase, limit)
	// Assert
	assert.NoError(t, err)
	assert.Empty(t, got)
}
