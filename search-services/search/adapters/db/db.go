package db

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/search/core"
)

type DB struct {
	log *slog.Logger
	db  *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}
	return &DB{log: log, db: db}, nil
}

type ComicDB struct {
	Id    int            `db:"id"`
	Url   string         `db:"url"`
	Words pq.StringArray `db:"words"`
}

func (db *DB) GetAll(ctx context.Context) ([]core.Comic, error) {
	var comicsDB []ComicDB
	if err := db.db.SelectContext(ctx, &comicsDB,
		`SELECT * FROM public.comics
		ORDER BY id ASC`,
	); err != nil {
		return nil, fmt.Errorf("get all comics: %w", err)
	}

	// map ComicDB to core Comic
	var comics []core.Comic
	for _, c := range comicsDB {
		comics = append(comics, core.Comic{
			Id:    c.Id,
			Url:   c.Url,
			Words: c.Words,
		})
	}

	return comics, nil
}
