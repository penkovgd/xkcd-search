package core

import "time"

type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusIdle    ServiceStatus = "idle"
)

type DBStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
}

type ServiceStats struct {
	DBStats
	ComicsTotal int
}

type Comic struct {
	ID    int
	URL   string
	Words []string
	Date  time.Time
	Title string
}

type XKCDInfo struct {
	ID          int
	URL         string
	Title       string
	SafeTitle   string
	Description string
	Alt         string
	Date        time.Time
}

type Message struct {
	Subject string
	Payload []byte
}
