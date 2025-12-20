package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/penkovgd/closer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	searchpb "yadro.com/course/proto/search"
	"yadro.com/course/search/adapters/db"
	searchgrpc "yadro.com/course/search/adapters/grpc"
	eventInit "yadro.com/course/search/adapters/initiators/event_initiator"
	tickInit "yadro.com/course/search/adapters/initiators/tick_initiator"
	"yadro.com/course/search/adapters/words"
	"yadro.com/course/search/config"
	"yadro.com/course/search/core/services"
)

func main() {
	cfgPath := flag.String("config", "./config.yaml", "path to the config file")
	flag.Parse()
	cfg := config.MustLoad(*cfgPath)

	log := mustMakeLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	wordsClient, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		return fmt.Errorf("create words client: %w", err)
	}
	defer closer.CloseOrPanic(log, wordsClient)

	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		return fmt.Errorf("create storage: %w", err)
	}

	search, err := services.NewSearchService(log, storage, wordsClient)
	if err != nil {
		return fmt.Errorf("create search: %w", err)
	}

	indexer := services.NewIndexer(storage)
	tickInitiator := tickInit.New(log, cfg.IndexTTL, indexer)
	eventInitiator, err := eventInit.New(log, cfg.BrokerAddress, indexer)
	if err != nil {
		return fmt.Errorf("create event initiator: %w", err)
	}
	defer closer.CloseOrLog(log, eventInitiator)

	isearch, err := services.NewISearchService(log, wordsClient, indexer)
	if err != nil {
		return fmt.Errorf("create isearch: %w", err)
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s := grpc.NewServer()
	searchpb.RegisterSearchServer(s, searchgrpc.NewServer(log, search, isearch))
	reflection.Register(s)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tickInitiator.Start(ctx)
	err = eventInitiator.Start(ctx)
	if err != nil {
		return fmt.Errorf("start event initiator: %w", err)
	}

	go func() {
		<-ctx.Done()
		log.Debug("trying to shutdown gracefully...")
		timer := time.AfterFunc(5*time.Second, func() {
			log.Warn("server couldn't stop gracefully in time. doing force stop")
			s.Stop()
		})
		defer timer.Stop()
		s.GracefulStop()
		log.Debug("server stopped gracefully")
	}()

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	slog.Default()
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
