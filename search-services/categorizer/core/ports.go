package core

import "context"

type Categorizer interface {
	Categorize(context.Context, Comic) (string, error)
}

type Publisher interface {
	Publish(context.Context, Message) error
}

