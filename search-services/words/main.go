package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	wordspb "yadro.com/course/proto/words"
	words "yadro.com/course/words/words"
)

const maxPhraseLen = 16 * 1024 // 16 KiB

type server struct {
	wordspb.UnimplementedWordsServer
}

func (s *server) Ping(_ context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	slog.Info("ping called")
	return nil, nil
}

func (s *server) Norm(_ context.Context, in *wordspb.WordsRequest) (*wordspb.WordsReply, error) {
	if len(in.GetPhrase()) > maxPhraseLen {
		errMsg := "request is too big"
		slog.Error(errMsg, "code", codes.ResourceExhausted)
		return nil, status.Error(codes.ResourceExhausted, errMsg)
	}
	normalized := words.Normalize(in.GetPhrase())

	return &wordspb.WordsReply{Words: normalized}, nil
}

type Config struct {
	Port string `yaml:"port" env:"PORT" env-default:"8080"`
}

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config.yaml", "path to config file")
	flag.Parse()

	var cfg Config
	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		slog.Error("cannot read config", "path", cfgPath, "err", err)
	}

	address := ":" + cfg.Port
	slog.Info(fmt.Sprintf("Listening at %v", cfg.Port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	wordspb.RegisterWordsServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
