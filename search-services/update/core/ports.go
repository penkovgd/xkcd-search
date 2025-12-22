package core

import (
	"context"
)

type Updater interface {
	Update(context.Context) error
	Stats(context.Context) (ServiceStats, error)
	Status(context.Context) ServiceStatus
	Drop(context.Context) error
}

type DB interface {
	Add(context.Context, Comic) error
	Stats(context.Context) (DBStats, error)
	Drop(context.Context) error
	IDs(context.Context) ([]int, error)
}

type XKCD interface {
	Get(context.Context, int) (XKCDInfo, error)
	LastID(context.Context) (int, error)
	GetImage(ctx context.Context, url string) ([]byte, string, error)
}

type Words interface {
	Norm(ctx context.Context, phrase string) ([]string, error)
}

type Publisher interface {
	Publish(context.Context, Message) error
}

type ImageStorage interface {
	Save(ctx context.Context, image string, data []byte) (string, error)
}
