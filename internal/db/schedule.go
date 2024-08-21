package db

import (
	"context"
	"errors"

	"cloud.google.com/go/firestore"
	"github.com/asherzog/thisor/internal/espn"
)

func (d *DB) scheduleCollection() *firestore.CollectionRef {
	return d.Client.Collection("schedule")
}

func (d *DB) AddSchedule(ctx context.Context, schedule *espn.Schedule) (*espn.Schedule, error) {
	_, err := d.scheduleCollection().Doc("2024").Set(ctx, schedule)
	if err != nil {
		return nil, err
	}

	// refresh in memory schedule
	d.Schedule = schedule
	return schedule, nil
}

func (d *DB) GetSchedule(ctx context.Context, id string) (*espn.Schedule, error) {
	if d.Schedule != nil {
		d.lg.Info("using in memory schedule")
		return d.Schedule, nil
	}

	doc, err := d.scheduleCollection().Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}

	var schedule espn.Schedule
	err = doc.DataTo(&schedule)
	if err != nil {
		return nil, err
	}

	// set in memory schedule
	d.Schedule = &schedule
	return &schedule, nil
}

func (d *DB) GetWeek(ctx context.Context, id int) (*espn.Schedule, error) {
	var week espn.Schedule
	if d.Schedule == nil {
		d.lg.Warn("no in memory schedule")
		d.GetSchedule(ctx, "2024")
	}

	for _, game := range d.Schedule.Games {
		if game.Week == id && game.Type == 2 {
			week.Games = append(week.Games, game)
		}
	}

	if len(week.Games) == 0 {
		return nil, errors.New("invalid week")
	}
	return &week, nil
}

func (d DB) GetGame(ctx context.Context, id string) (espn.Game, error) {
	var game espn.Game
	if d.Schedule == nil {
		d.lg.Warn("no in memory schedule")
		d.GetSchedule(ctx, "2024")
	}

	for _, game := range d.Schedule.Games {
		if game.ID == id {
			return game, nil
		}
	}

	return game, errors.New("game not found")
}
