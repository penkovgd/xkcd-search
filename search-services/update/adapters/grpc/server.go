package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/core"
)

func NewServer(service core.Updater, db core.DB) *Server {
	return &Server{service: service, db: db}
}

type Server struct {
	updatepb.UnimplementedUpdateServer
	service core.Updater
	db      core.DB
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *Server) Status(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatusReply, error) {
	var status updatepb.Status
	switch s.service.Status(ctx) {
	case core.StatusIdle:
		status = updatepb.Status_STATUS_IDLE
	case core.StatusRunning:
		status = updatepb.Status_STATUS_RUNNING
	default:
		status = updatepb.Status_STATUS_UNSPECIFIED
	}
	return &updatepb.StatusReply{Status: status}, nil
}

func (s *Server) Update(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Update(ctx); err != nil {
		if errors.Is(err, core.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "update is already running")
		}
		return nil, status.Error(codes.Internal, "failed to update")
	}
	return nil, nil
}

func (s *Server) Stats(ctx context.Context, _ *emptypb.Empty) (*updatepb.StatsReply, error) {
	stats, err := s.service.Stats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stats")
	}
	return &updatepb.StatsReply{
		WordsTotal:    int64(stats.WordsTotal),
		WordsUnique:   int64(stats.WordsUnique),
		ComicsTotal:   int64(stats.ComicsTotal),
		ComicsFetched: int64(stats.ComicsFetched),
	}, nil
}

func (s *Server) Drop(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.service.Drop(ctx); err != nil {
		return nil, status.Error(codes.Internal, "failed to drop")
	}
	return nil, nil
}

func (s *Server) GetComics(ctx context.Context, input *updatepb.GetComicsRequest) (*updatepb.GetComicsReply, error) {
	category := input.GetCategory()
	comics, err := s.db.GetComics(ctx, category)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get comics")
	}

	pbComics := make([]*updatepb.Comic, len(comics))
	for i, c := range comics {
		pbComics[i] = &updatepb.Comic{
			Id:       int64(c.ID),
			Url:      c.URL,
			Title:    c.Title,
			Date:     timestamppb.New(c.Date),
			Category: c.Category,
		}
	}

	return &updatepb.GetComicsReply{Comics: pbComics}, nil
}

func (s *Server) GetCategories(ctx context.Context, _ *emptypb.Empty) (*updatepb.GetCategoriesReply, error) {
	stats, err := s.db.GetCategories(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get categories")
	}

	resp := &updatepb.GetCategoriesReply{
		CategoryStats: make([]*updatepb.CategoryStats, 0, len(stats)),
	}

	for _, stat := range stats {
		resp.CategoryStats = append(resp.CategoryStats, &updatepb.CategoryStats{
			Category: stat.Category,
			Count:    int64(stat.Count),
		})
	}

	return resp, nil
}
