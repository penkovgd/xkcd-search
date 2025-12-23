package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/penkovgd/closer"
	"yadro.com/course/categorizer/adapters/nats/pub"
	"yadro.com/course/categorizer/adapters/nats/sub"
	"yadro.com/course/categorizer/adapters/ollama"
	"yadro.com/course/categorizer/config"
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

	categorizer := ollama.New(log, cfg.Ollama.URL, cfg.Ollama.Model)

	// nats publisher
	pub, err := pub.New(log, cfg.BrokerAddress)
	if err != nil {
		return fmt.Errorf("create nats publisher: %w", err)
	}
	// nats subscriber
	sub, err := sub.New(log, cfg.BrokerAddress, categorizer, pub, cfg.Concurrency)
	if err != nil {
		return fmt.Errorf("create nats subsciber: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := sub.Start(ctx); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	<-ctx.Done()
	log.Debug("trying to shutdown gracefully...")
	closer.CloseOrLog(log, sub)
	closer.CloseOrLog(log, pub)
	log.Debug("server stopped gracefully")

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
