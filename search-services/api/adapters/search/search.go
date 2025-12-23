package search

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"yadro.com/course/api/core"
	searchpb "yadro.com/course/proto/search"
)

type Client struct {
	log    *slog.Logger
	client searchpb.SearchClient
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
		client: searchpb.NewSearchClient(conn),
		log:    log,
		conn:   conn,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, nil)
	return err
}

func (c *Client) Search(ctx context.Context, phrase string, limit int) ([]core.Comic, error) {
	out, err := c.client.Search(ctx, &searchpb.SearchRequest{
		Phrase: phrase,
		Limit:  int64(limit),
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return nil, fmt.Errorf("search gRPC client: %w", core.ErrBadArguments)
		}
		return nil, fmt.Errorf("search gRPC client: %w", err)
	}
	comicspb := out.GetComics()

	comics := make([]core.Comic, len(comicspb))
	for i, c := range comicspb {
		comics[i] = core.Comic{
			ID:       int(c.GetId()),
			URL:      c.GetUrl(),
			Title:    c.GetTitle(),
			Date:     c.GetDate().AsTime(),
			Category: c.GetCategory(),
		}
	}
	return comics, nil
}

func (c *Client) SearchIndex(ctx context.Context, phrase string, limit int) ([]core.Comic, error) {
	out, err := c.client.SearchIndex(ctx, &searchpb.SearchRequest{
		Phrase: phrase,
		Limit:  int64(limit),
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return nil, fmt.Errorf("search gRPC client: %w", core.ErrBadArguments)
		}
		return nil, fmt.Errorf("search gRPC client: %w", err)
	}
	comicspb := out.GetComics()

	comics := make([]core.Comic, len(comicspb))
	for i, c := range comicspb {
		comics[i] = core.Comic{
			ID:       int(c.GetId()),
			URL:      c.GetUrl(),
			Title:    c.GetTitle(),
			Date:     c.GetDate().AsTime(),
			Category: c.GetCategory(),
		}
	}
	return comics, nil
}
