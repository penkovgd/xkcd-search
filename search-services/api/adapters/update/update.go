package update

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"yadro.com/course/api/core"
	updatepb "yadro.com/course/proto/update"
)

type Client struct {
	log    *slog.Logger
	client updatepb.UpdateClient
	conn   *grpc.ClientConn
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  1 * time.Second,
				Multiplier: 1.6,
				MaxDelay:   10 * time.Second,
			},
			MinConnectTimeout: 10 * time.Second,
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: updatepb.NewUpdateClient(conn),
		log:    log,
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, nil)
	return err
}

func (c Client) Status(ctx context.Context) (core.UpdateStatus, error) {
	out, err := c.client.Status(ctx, nil)
	if err != nil {
		c.log.Error("failed to get updater status", "error", err)
		return core.StatusUpdateUnknown, err
	}

	var status core.UpdateStatus
	switch out.Status {
	case updatepb.Status_STATUS_IDLE:
		status = core.StatusUpdateIdle
	case updatepb.Status_STATUS_RUNNING:
		status = core.StatusUpdateRunning
	default:
		status = core.StatusUpdateUnknown
	}

	return status, nil
}

func (c Client) Stats(ctx context.Context) (core.UpdateStats, error) {
	out, err := c.client.Stats(ctx, nil)
	if err != nil {
		c.log.Error("failed to get updater stats", "error", err)
		return core.UpdateStats{}, err
	}
	return core.UpdateStats{
		WordsTotal:    int(out.WordsTotal),
		WordsUnique:   int(out.WordsUnique),
		ComicsFetched: int(out.ComicsFetched),
		ComicsTotal:   int(out.ComicsTotal),
	}, nil
}

func (c Client) Update(ctx context.Context) error {
	_, err := c.client.Update(ctx, nil)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			c.log.Info("update is already running", "error", err)
			return core.ErrAlreadyExists
		}
		c.log.Error("failed to update comics", "error", err)
		return err
	}
	return nil
}

func (c Client) Drop(ctx context.Context) error {
	_, err := c.client.Drop(ctx, nil)
	if err != nil {
		c.log.Error("failed to drop comics", "error", err)
		return err
	}
	return err
}

func (c Client) GetComics(ctx context.Context, category string) ([]core.Comic, error) {
	out, err := c.client.GetComics(
		ctx,
		&updatepb.GetComicsRequest{
			Category: category, // "" is acceptable
		},
	)
	if err != nil {
		c.log.Error(
			"failed to get comics",
			"category", category,
			"error", err,
		)

		if status.Code(err) == codes.InvalidArgument {
			return nil, core.ErrBadArguments
		}

		return nil, err
	}

	comics := make([]core.Comic, 0, len(out.Comics))
	for _, cpb := range out.Comics {
		comics = append(comics, core.Comic{
			ID:       int(cpb.Id),
			URL:      cpb.Url,
			Title:    cpb.Title,
			Date:     cpb.Date.AsTime(),
			Category: cpb.Category,
		})
	}

	return comics, nil
}

func (c Client) GetCategories(ctx context.Context) ([]core.CategoryStats, error) {
	out, err := c.client.GetCategories(ctx, nil)
	if err != nil {
		c.log.Error("failed to get categories", "error", err)
		return nil, err
	}

	stats := make([]core.CategoryStats, 0, len(out.CategoryStats))
	for _, item := range out.CategoryStats {
		stats = append(stats, core.CategoryStats{
			Category: item.Category,
			Count:    int(item.Count),
		})
	}

	return stats, nil
}
