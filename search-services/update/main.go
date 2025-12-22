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
	updatepb "yadro.com/course/proto/update"
	"yadro.com/course/update/adapters/db"
	updategrpc "yadro.com/course/update/adapters/grpc"
	"yadro.com/course/update/adapters/imgstorage"
	publisher "yadro.com/course/update/adapters/nats-publisher"
	"yadro.com/course/update/adapters/scheduler"
	"yadro.com/course/update/adapters/words"
	"yadro.com/course/update/adapters/xkcd"
	"yadro.com/course/update/config"
	"yadro.com/course/update/core"
)

func main() {

	// config
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	// logger
	log := mustMakeLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// database adapter
	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}
	if err := storage.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate db: %v", err)
	}

	// xkcd adapter
	xkcd, err := xkcd.NewClient(cfg.XKCD.URL, cfg.XKCD.Timeout, log)
	if err != nil {
		return fmt.Errorf("failed create XKCD client: %v", err)
	}

	// words adapter
	words, err := words.NewClient(cfg.WordsAddress, log)
	if err != nil {
		return fmt.Errorf("failed create Words client: %v", err)
	}
	defer closer.CloseOrLog(log, words)

	// nats broker
	publisher, err := publisher.New(log, cfg.BrokerAddress)
	if err != nil {
		return fmt.Errorf("failed to create nats publisher: %w", err)
	}
	defer closer.CloseOrLog(log, publisher)

	// minio image storage
	imgStorage, err := imgstorage.New(cfg.Minio.ConnectAddress, cfg.Minio.RootUser,
		cfg.Minio.RootPassword, cfg.Minio.BucketName, cfg.Minio.UseSSl, cfg.Minio.PublicAddress)
	if err != nil {
		return fmt.Errorf("failed to create minio storage: %w", err)
	}
	// service
	updater, err := core.NewService(log, storage, xkcd, words, publisher, imgStorage, cfg.XKCD.Concurrency)
	if err != nil {
		return fmt.Errorf("failed create Update service: %w", err)
	}

	// grpc server
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	updatepb.RegisterUpdateServer(s, updategrpc.NewServer(updater))
	reflection.Register(s)

	// context for Ctrl-C and docker
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	scheduler, err := scheduler.New(log, updater, cfg.XKCD.Schedule)
	if err != nil {
		return fmt.Errorf("failed to create Scheduler: %w", err)
	}
	err = scheduler.Start(ctx)
	if err != nil {
		return fmt.Errorf("start scheduler: %w", err)
	}
	defer closer.CloseOrLog(log, scheduler)

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		timer := time.AfterFunc(5*time.Second, func() {
			log.Warn("server couldn't stop gracefully in time. doing force stop")
			s.Stop()
		})
		defer timer.Stop()
		s.GracefulStop()
		log.Debug("server stopped gracefully")
	}()

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
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
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
