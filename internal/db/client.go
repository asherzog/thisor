package db

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/joeshaw/envdecode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DB struct {
	ProjectID string `env:"FIRESTORE_PROJECT_ID"`
	Client    *firestore.Client
}

type Game struct {
	Test string
}

func NewClient(ctx context.Context) (*DB, error) {
	var db = DB{}

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
