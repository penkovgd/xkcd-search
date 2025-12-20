package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/penkovgd/closer"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {

	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Add(ctx context.Context, comic core.Comic) error {
	res, err := db.conn.ExecContext(ctx,
		`INSERT INTO comics(id, url, words) 
		VALUES ($1, $2, $3) 
		ON CONFLICT DO NOTHING`,
		comic.ID, comic.URL, comic.Words,
	)
	if err != nil {
		db.log.Error("failed to add comic", "id", comic.ID, "error", err)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		db.log.Error("failed to get number of rows affected", "error", err)
		return err
	}

	if rowsAffected == 0 {
		db.log.Debug("skip adding comic: already exists", "id", comic.ID)
	} else {
		db.log.Debug("successfully added comic to storage", "id", comic.ID)
	}

	return nil
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	rows, err := db.conn.QueryxContext(ctx, `SELECT words FROM comics`)
	if err != nil {
		db.log.Error("failed to query comics stats", "error", err)
		return core.DBStats{}, err
	}
	defer closer.CloseOrLog(db.log, rows)

	wordsTotal := 0
	wordsUnique := make(map[string]bool)
	rowsTotal := 0

	for rows.Next() {
		var words []string
		if err := rows.Scan(pq.Array(&words)); err != nil {
			db.log.Error("failed to scan words", "error", err)
			return core.DBStats{}, err
		}
		wordsTotal += len(words)
		for _, w := range words {
			wordsUnique[w] = true
		}
		rowsTotal++
	}
	if err := rows.Err(); err != nil {
		db.log.Error("rows iteration error", "error", err)
		return core.DBStats{}, err
	}

	db.log.Debug("successfuly calculated stats",
		"comics_count", rowsTotal,
		"words_total", wordsTotal,
		"unique_words", len(wordsUnique))

	return core.DBStats{
		WordsTotal:    wordsTotal,
		WordsUnique:   len(wordsUnique),
		ComicsFetched: rowsTotal,
	}, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	err := db.conn.SelectContext(ctx, &ids, `SELECT id FROM comics`)
	if err != nil {
		db.log.Error("failed to select comics ids", "error", err)
		return nil, err
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, `TRUNCATE TABLE comics`)
	if err != nil {
		db.log.Error("failed to truncate table", "error", err)
	}
	db.log.Debug("successfully deleted all comics")
	return nil
}
