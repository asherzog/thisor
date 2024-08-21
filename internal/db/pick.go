package db

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"github.com/asherzog/thisor/internal/espn"
	"google.golang.org/api/iterator"
)

type Pick struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	GameID    string    `json:"game_id"`
	Selection espn.Team `json:"selection,required"`
	Week      int       `json:"week"`
	WinScore  int       `json:"win_score"`
	LoseScore int       `json:"lose_score"`
	IsLocked  bool      `json:"is_locked"`
}

type PickList struct {
	Users map[string][]Pick
}

func (p *PickList) GetUsers() map[string][]Pick {
	if p != nil {
		return p.Users
	}
	return nil
}

func (d *DB) pickCollection() *firestore.CollectionRef {
	return d.Client.Collection("picks")
}

func (d *DB) CreatePick(ctx context.Context, pick Pick) (*Pick, error) {
	// Get Game details
	game, err := d.GetGame(ctx, pick.GameID)
	if err != nil {
		return nil, err
	}
	selection, err := validateSelection(game, pick)
	if err != nil {
		return nil, err
	}
	pick.Selection = selection
	// Check if pick for this game already exists
	iter := d.pickCollection().Where("GameID", "==", pick.GameID).Where("UserID", "==", pick.UserID).Documents(ctx)
	isUpdate := false
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			if isUpdate {
				return &pick, nil
			}
			break
		}
		if err != nil {
			return nil, err
		}
		var oldPick Pick
		err = doc.DataTo(&oldPick)
		if err != nil {
			d.lg.Error("unable to parse pick", "error", err.Error())
			return nil, err
		}
		pick.ID = oldPick.ID
		if _, err = d.pickCollection().Doc(oldPick.ID).Set(ctx, pick); err != nil {
			return nil, err
		}
		isUpdate = true
	}
	// create pick
	docRef := d.pickCollection().NewDoc()
	pick.ID = docRef.ID
	_, err = docRef.Set(ctx, pick)
	if err != nil {
		return nil, err
	}
	return &pick, nil
}

func (d *DB) GetPick(ctx context.Context, id string) (*Pick, error) {
	doc, err := d.pickCollection().Doc(id).Get(ctx)
	if err != nil {
		d.lg.Error("unable to fetch pick", "error", err.Error())
		return nil, err
	}
	var pick Pick
	err = doc.DataTo(&pick)
	if err != nil {
		d.lg.Error("unable to parse pick", "error", err.Error())
		return nil, err
	}
	return &pick, nil
}

func (d *DB) GetAllPicks(ctx context.Context) (*PickList, error) {
	picks := PickList{Users: map[string][]Pick{}}
	iter := d.pickCollection().Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			d.lg.Error("unable to fetch pick", "error", err.Error())
			return nil, err
		}
		var pick Pick
		if err := doc.DataTo(&pick); err != nil {
			d.lg.Error("unable to parse pick", "error", err.Error())
			return nil, err
		}
		picks.Users[pick.UserID] = append(picks.Users[pick.UserID], pick)
	}
	return &picks, nil
}

func (d *DB) GetPicksForUser(ctx context.Context, id string) (*PickList, error) {
	pickList := PickList{Users: map[string][]Pick{}}
	iter := d.pickCollection().Where("UserID", "==", id).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var pick Pick
		err = doc.DataTo(&pick)
		if err != nil {
			d.lg.Error("unable to parse pick", "error", err.Error())
			return nil, err
		}
		pickList.Users[id] = append(pickList.Users[id], pick)
	}
	return &pickList, nil
}

func validateSelection(g espn.Game, p Pick) (espn.Team, error) {
	selection := p.Selection.ID
	if selection == g.Home.ID {
		return g.Home, nil
	}
	if selection == g.Away.ID {
		return g.Away, nil
	}
	return espn.Team{}, errors.New("invalid pick selection")
}
