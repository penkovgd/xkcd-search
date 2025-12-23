package core

import "time"

type Comic struct {
	Id       int
	Url      string
	Words    []string
	Title    string
	Date     time.Time
	Category string
}
