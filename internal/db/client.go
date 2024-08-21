package db

import (
	"context"
	"log/slog"

	"cloud.google.com/go/firestore"
	"github.com/asherzog/thisor/internal/espn"
	"github.com/joeshaw/envdecode"
)

type DB struct {
	ProjectID string `env:"FIRESTORE_PROJECT_ID"`
	Client    *firestore.Client
	Schedule  *espn.Schedule
	lg        *slog.Logger
}

func NewClient(ctx context.Context, lg *slog.Logger) (*DB, error) {
	var db = DB{lg: lg}

	if err := envdecode.StrictDecode(&db); err != nil {
		return nil, err
	}

	client, err := firestore.NewClient(ctx, db.ProjectID)
	if err != nil {
		return nil, err
	}

	db.Client = client
	return &db, nil
}
