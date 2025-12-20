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

func TestCountOccurrences(t *testing.T) {
	tests := []struct {
		name     string
		keyword  string
		document []string
		want     int
	}{
		{
			name:     "empty document",
			keyword:  "test",
			document: []string{},
			want:     0,
		},
		{
			name:     "single occurrence",
			keyword:  "test",
			document: []string{"test"},
			want:     1,
		},
		{
			name:     "multiple occurrences",
			keyword:  "test",
			document: []string{"test", "hello", "test", "world", "test"},
			want:     3,
		},
		{
			name:     "no occurrences",
			keyword:  "test",
			document: []string{"hello", "world"},
			want:     0,
		},
		{
			name:     "case sensitive",
			keyword:  "Test",
			document: []string{"test", "Test", "TEST"},
			want:     1,
		},
		{
			name:     "empty keyword",
			keyword:  "",
			document: []string{"", "hello", "", "world"},
			want:     2,
		},
		{
			name:     "russian",
			keyword:  "привет",
			document: []string{"привет", "мир", "привет"},
			want:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countOccurrences(tt.keyword, tt.document)
			assert.Equal(t, tt.want, got)
		})
	}
}

var (
	simpleStorageReturn = []core.Comic{c1, c2, c3}
)

func newSimpleMocksSearch(t *testing.T, ctrl *gomock.Controller, comics []core.Comic) (*mock_core.MockStorage, *mock_core.MockNormalizer) {
	t.Helper()

	mockStorageSimple := mock_core.NewMockStorage(ctrl)
	mockStorageSimple.
		EXPECT().
		GetAll(context.TODO()).
		Return(comics, nil).
		AnyTimes()
	mockNormSimple := mock_core.NewMockNormalizer(ctrl)
	mockNormSimple.
		EXPECT().
		Norm(context.TODO(), gomock.Any()).
		DoAndReturn(func(_ context.Context, phrase string) ([]string, error) {
			return strings.Fields(phrase), nil
		}).
		AnyTimes()

	return mockStorageSimple, mockNormSimple
}

func TestSearch_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	mockStorage, mockNorm := newSimpleMocksSearch(t, ctrl, simpleStorageReturn)

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

	service, err := NewSearchService(slog.Default(), mockStorage, mockNorm)
	require.NoError(t, err)

	// Act
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.Search(context.TODO(), tt.phrase, tt.limit)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestSearch_BadLimit(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	mockStorage, mockNorm := newSimpleMocksSearch(t, ctrl, simpleStorageReturn)

	phrase := "hello test go"

	tests := []struct {
		name  string
		limit int
		want  []core.Comic
	}{
		{
			name:  "zero",
			limit: 0,
			want:  simpleStorageReturn,
		},
		{
			name:  "negative",
			limit: -1,
			want:  simpleStorageReturn,
		},
		{
			name:  "more than search has",
			limit: 10_000,
			want:  simpleStorageReturn,
		},
	}

	service, err := NewSearchService(slog.Default(), mockStorage, mockNorm)
	require.NoError(t, err)

	// Act
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.Search(context.TODO(), phrase, tt.limit)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestSearch_EmptyPhrase(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	mockStorage, mockNorm := newSimpleMocksSearch(t, ctrl, simpleStorageReturn)

	phrase := ""
	limit := 10

	service, err := NewSearchService(slog.Default(), mockStorage, mockNorm)
	require.NoError(t, err)

	// Act
	got, err := service.Search(context.TODO(), phrase, limit)
	// Assert
	assert.ErrorIs(t, err, core.ErrBadArguments)
	assert.Empty(t, got)
}

func TestSearch_NoComicsInStorage(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	mockStorage, mockNorm := newSimpleMocksSearch(t, ctrl, []core.Comic{})
	service, err := NewSearchService(slog.Default(), mockStorage, mockNorm)
	require.NoError(t, err)

	phrase := "doesnt matter"
	limit := 10

	// Act
	got, err := service.Search(context.TODO(), phrase, limit)
	// Assert
	assert.NoError(t, err)
	assert.Empty(t, got)
}
