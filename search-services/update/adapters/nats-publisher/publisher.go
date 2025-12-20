package publisher

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"yadro.com/course/update/core"
)

type Publisher struct {
	log *slog.Logger
	nc  *nats.Conn
}

func New(log *slog.Logger, url string) (*Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}
	return &Publisher{log: log, nc: nc}, nil
}

func (p *Publisher) Publish(ctx context.Context, msg core.Message) error {
	err := p.nc.Publish(msg.Subject, msg.Payload)
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}
	if err = p.nc.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	p.log.Debug("published message", "msg", msg)
	return nil
}

func (p *Publisher) Close() error {
	p.nc.Close()
	return nil
}
