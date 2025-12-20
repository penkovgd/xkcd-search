package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/penkovgd/closer"
	"yadro.com/course/api/core"
)

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := PingResponse{
			Replies: make(map[string]string),
		}
		for name, pinger := range pingers {
			if err := pinger.Ping(r.Context()); err != nil {
				reply.Replies[name] = "unavailable"
				log.Error("one of services is not available", "service", name)
				continue
			}
			reply.Replies[name] = "ok"
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			log.Error("failed to encode reply", "error", err)
		}
	}

}

type Authenticator interface {
	Login(user, password string) (string, error)
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func NewLoginHandler(log *slog.Logger, auth Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			log.Error("decode login body", "error", err)
			http.Error(w, "invalid request body", http.StatusInternalServerError)
			return
		}
		defer closer.CloseOrLog(log, r.Body)

		if body.Name == "" || body.Password == "" {
			http.Error(w, "name & password required", http.StatusUnauthorized)
			return
		}

		token, err := auth.Login(body.Name, body.Password)
		if err != nil {
			if errors.Is(err, core.ErrInvalidCredentials) {
				log.Debug("invalid credentials", "name", body.Name, "passwors", body.Password, "error", err)
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			log.Debug("login", "error", err)
			http.Error(w, "failed to login", http.StatusInternalServerError)
			return
		}

		if _, err := fmt.Fprint(w, token); err != nil {
			log.Error("write response", "error", err)
		}
	}
}

func NewUpdateHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Update(r.Context()); err != nil {
			if errors.Is(err, core.ErrAlreadyExists) {
				log.Info("update is already running", "error", err)
				w.WriteHeader(http.StatusAccepted)
				return
			}
			log.Error("failed to perform update", "error", err)
			http.Error(w, "failed to perform update", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

type StatsResponse struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

func NewUpdateStatsHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := updater.Stats(r.Context())
		if err != nil {
			log.Error("failed to get updater stats", "error", err)
			http.Error(w, "failed to get updater stats", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(
			StatsResponse{
				WordsTotal:    stats.WordsTotal,
				WordsUnique:   stats.WordsUnique,
				ComicsFetched: stats.ComicsFetched,
				ComicsTotal:   stats.ComicsTotal,
			}); err != nil {
			log.Error("failed to encode reply", "error", err)
		}
	}
}

type StatusResponse struct {
	Status core.UpdateStatus `json:"status"`
}

func NewUpdateStatusHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := updater.Status(r.Context())
		if err != nil {
			log.Error("failed to get updater status", "error", err)
			http.Error(w, "failed to get updater status", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(StatusResponse{Status: status}); err != nil {
			log.Error("failed to encode reply", "error", err)
		}
	}
}

func NewDropHandler(log *slog.Logger, updater core.Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := updater.Drop(r.Context()); err != nil {
			log.Error("failed to perform drop", "error", err)
			http.Error(w, "failed to perform drop", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

type SearchResponse struct {
	Comics []core.Comic `json:"comics"`
	Total  int          `json:"total"`
}

func NewSearchHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if phrase == "" {
			log.Error("get query params", "error", "phrase cannot be empty")
			http.Error(w, "phrase cannot be empty", http.StatusBadRequest)
			return
		}
		if limit == 0 {
			limit = 10
		}

		comics, err := searcher.Search(r.Context(), phrase, limit)
		if err != nil {
			log.Error(
				"search rest handler: perform search",
				"phrase", phrase,
				"limit", limit,
				"error", err,
			)
			if errors.Is(err, core.ErrBadArguments) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to perform search", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(
			SearchResponse{
				Comics: comics,
				Total:  len(comics),
			},
		); err != nil {
			log.Error("failed to encode reply", "error", err)
		}
	}
}

func NewSearchIndexHandler(log *slog.Logger, searcher core.Searcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phrase := r.URL.Query().Get("phrase")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if phrase == "" {
			log.Error("get query params", "error", "phrase cannot be empty")
			http.Error(w, "phrase cannot be empty", http.StatusBadRequest)
			return
		}
		if limit == 0 {
			limit = 10
		}

		comics, err := searcher.SearchIndex(r.Context(), phrase, limit)
		if err != nil {
			log.Error(
				"isearch rest handler: perform isearch",
				"phrase", phrase,
				"limit", limit,
				"error", err,
			)
			if errors.Is(err, core.ErrBadArguments) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to perform isearch", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(
			SearchResponse{
				Comics: comics,
				Total:  len(comics),
			},
		); err != nil {
			log.Error("failed to encode reply", "error", err)
		}
	}
}
