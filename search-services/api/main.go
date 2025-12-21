package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/penkovgd/closer"
	"yadro.com/course/api/adapters/aaa"
	"yadro.com/course/api/adapters/rest"
	"yadro.com/course/api/adapters/rest/middleware"
	"yadro.com/course/api/adapters/search"
	"yadro.com/course/api/adapters/update"
	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

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
		return fmt.Errorf("init words adapter: %v", err)
	}
	defer closer.CloseOrLog(log, wordsClient)

	updateClient, err := update.NewClient(cfg.UpdateAddress, log)
	if err != nil {
		return fmt.Errorf("init update adapter: %v", err)
	}
	defer closer.CloseOrLog(log, updateClient)

	searchClient, err := search.NewClient(cfg.SearchAddress, log)
	if err != nil {
		return fmt.Errorf("init search adapter: %v", err)
	}
	defer closer.CloseOrLog(log, searchClient)

	aaa, err := aaa.New(cfg.TokenTTL, log)
	if err != nil {
		return fmt.Errorf("init aaa adapter: %v", err)
	}
	defer closer.CloseOrLog(log, searchClient)

	mux := http.NewServeMux()
	mux.Handle("GET /api/ping", rest.NewPingHandler(log, map[string]core.Pinger{
		"words":  wordsClient,
		"update": updateClient,
		"search": searchClient,
	}))
	mux.Handle("POST /api/db/update", middleware.Auth(rest.NewUpdateHandler(log, updateClient), aaa))
	mux.Handle("GET /api/db/stats", rest.NewUpdateStatsHandler(log, updateClient))
	mux.Handle("GET /api/db/status", rest.NewUpdateStatusHandler(log, updateClient))
	mux.Handle("DELETE /api/db", middleware.Auth(rest.NewDropHandler(log, updateClient), aaa))
	mux.Handle("GET /api/search", middleware.Concurrency(rest.NewSearchHandler(log, searchClient), cfg.SearchConcurrency))
	mux.Handle("GET /api/isearch", middleware.Rate(rest.NewSearchIndexHandler(log, searchClient), cfg.SearchRate))
	mux.Handle("POST /api/login", rest.NewLoginHandler(log, aaa))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     middleware.CORS(mux.ServeHTTP, cfg),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server closed unexpectedly: %v", err)
		}
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
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level, AddSource: true})
	return slog.New(handler)
}
