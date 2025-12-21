package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/core"
)

func NewServer(log *slog.Logger, search core.Searcher, isearch core.ISearcher) *Server {
	return &Server{
		log:     log,
		search:  search,
		isearch: isearch,
	}
}

type Server struct {
	searchpb.UnimplementedSearchServer
	search  core.Searcher
	isearch core.ISearcher
	log     *slog.Logger
}

func (s *Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *Server) Search(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	comics, err := s.search.Search(ctx, in.GetPhrase(), int(in.GetLimit()))
	if err != nil {
		s.log.Error("gRPC search:",
			"phrase", in.GetPhrase(),
			"limit", in.GetLimit(),
			"error", err,
		)
		if errors.Is(err, core.ErrBadArguments) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to perform search")
	}

	pbComics := make([]*searchpb.Comic, len(comics))
	for i, c := range comics {
		pbComics[i] = &searchpb.Comic{Id: int64(c.Id), Url: c.Url, Title: c.Title, Date: timestamppb.New(c.Date)}
	}

	return &searchpb.SearchReply{Comics: pbComics}, nil
}

func (s *Server) SearchIndex(ctx context.Context, in *searchpb.SearchRequest) (*searchpb.SearchReply, error) {
	comics, err := s.isearch.SearchIndex(ctx, in.GetPhrase(), int(in.GetLimit()))
	if err != nil {
		s.log.Error("gRPC isearch:",
			"phrase", in.GetPhrase(),
			"limit", in.GetLimit(),
			"error", err,
		)
		if errors.Is(err, core.ErrBadArguments) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to perform isearch")
	}

	pbComics := make([]*searchpb.Comic, len(comics))
	for i, c := range comics {
		pbComics[i] = &searchpb.Comic{Id: int64(c.Id), Url: c.Url, Title: c.Title, Date: timestamppb.New(c.Date)}
	}

	return &searchpb.SearchReply{Comics: pbComics}, nil
}
