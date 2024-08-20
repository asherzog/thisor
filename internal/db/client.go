package db

import (
	"context"
	"log/slog"

	"cloud.google.com/go/firestore"
	"github.com/asherzog/thisor/internal/espn"
	"github.com/joeshaw/envdecode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DB struct {
	ProjectID string `env:"FIRESTORE_PROJECT_ID"`
	Client    *firestore.Client
	Schedule  *espn.Schedule
	lg        *slog.Logger
}

type Game struct {
	Test string
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

func (d DB) gameCollection() *firestore.CollectionRef {
	return d.Client.Collection("games")
}

func (d DB) GetGame(ctx context.Context, gameId string) (Game, error) {
	doc, err := d.gameCollection().Doc(gameId).Get(ctx)
	if err != nil && status.Code(err) != codes.NotFound {
		return Game{}, err
	}
	if err != nil && status.Code(err) == codes.NotFound {
		return Game{
			Test: "",
		}, nil
	}

	var game Game
	err = doc.DataTo(&game)
	if err != nil {
		return Game{}, err
	}

	return game, nil
}
