package core

import "time"

type UpdateStatus string

const (
	StatusUpdateUnknown UpdateStatus = "unknown"
	StatusUpdateIdle    UpdateStatus = "idle"
	StatusUpdateRunning UpdateStatus = "running"
)

type UpdateStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
	ComicsTotal   int
}

type Comic struct {
	ID       int
	URL      string
	Title    string
	Date     time.Time
	Category string
}

type CategoryStats struct {
	Category string
	Count    int
}
