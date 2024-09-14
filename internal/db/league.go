package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/asherzog/thisor/internal/espn"
	"google.golang.org/api/iterator"
)

type League struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Users []User                 `json:"users"`
	Weeks map[string][]espn.Game `json:"weeks"`
	Admin string                 `json:"admin"`
}

type LeagueList struct {
	Leagues []League `json:"leagues"`
}

func (d *DB) leagueCollection() *firestore.CollectionRef {
	return d.Client.Collection("league")
}

func (d *DB) GetAllLeagues(ctx context.Context) (*LeagueList, error) {
	leagues := LeagueList{Leagues: []League{}}
	iter := d.leagueCollection().Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			d.lg.Error("unable to fetch league", "error", err.Error())
			return nil, err
		}
		var league League
		if err := doc.DataTo(&league); err != nil {
			d.lg.Error("unable to parse league", "error", err.Error())
			return nil, err
		}
		leagues.Leagues = append(leagues.Leagues, league)
	}
	return &leagues, nil
}

func (d *DB) CreateLeague(ctx context.Context, l League) (*League, error) {
	// TODO: pull admin from authenticated user
	admin, err := d.GetUser(ctx, l.Admin, "")
	if err != nil {
		return nil, errors.New("invalid admin user")
	}
	docRef := d.leagueCollection().NewDoc()
	l.ID = docRef.ID
	l.Admin = admin.ID
	if l.Users == nil {
		l.Users = []User{*admin}
	}

	if l.Weeks == nil {
		l.Weeks = make(map[string][]espn.Game)
	}

	s, err := d.GetSchedule(ctx, "2024")
	if err != nil {
		return nil, err
	}

	for _, g := range s.Games {
		// reg season only
		if g.Type == 2 {
			w := strconv.Itoa(g.Week)
			l.Weeks[w] = append(l.Weeks[w], g)
		}
	}

	_, err = docRef.Set(ctx, l)
	if err != nil {
		return nil, err
	}

	admin.Leagues = append(admin.Leagues, l)
	_, err = d.userCollection().Doc(admin.ID).Update(ctx, []firestore.Update{
		{
			Path:  "Leagues",
			Value: admin.Leagues,
		},
	})
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (d *DB) GetLeague(ctx context.Context, id string) (*League, error) {
	doc, err := d.leagueCollection().Doc(id).Get(ctx)
	if err != nil {
		d.lg.Error("unable to fetch league", "error", err.Error())
		return nil, err
	}
	var league League
	err = doc.DataTo(&league)
	if err != nil {
		d.lg.Error("unable to parse league", "error", err.Error())
		return nil, err
	}
	return &league, nil
}

func (d *DB) AddUserToLeague(ctx context.Context, leagueId, userId string) (*[]User, error) {
	user, err := d.GetUser(ctx, userId, "")
	if err != nil {
		return nil, errors.New("invalid user provided")
	}
	doc, err := d.leagueCollection().Doc(leagueId).Get(ctx)
	if err != nil {
		d.lg.Error("unable to fetch league", "error", err.Error())
		return nil, err
	}
	var league League
	err = doc.DataTo(&league)
	if err != nil {
		d.lg.Error("unable to parse league", "error", err.Error())
		return nil, err
	}

	for _, u := range league.Users {
		if u.ID == user.ID {
			return &league.Users, nil
		}
	}
	league.Users = append(league.Users, *user)
	_, err = d.leagueCollection().Doc(leagueId).Update(ctx, []firestore.Update{
		{
			Path:  "Users",
			Value: league.Users,
		},
	})
	if err != nil {
		return nil, err
	}

	user.Leagues = append(user.Leagues, league)
	_, err = d.userCollection().Doc(user.ID).Update(ctx, []firestore.Update{
		{
			Path:  "Leagues",
			Value: user.Leagues,
		},
	})
	if err != nil {
		return nil, err
	}
	return &league.Users, nil
}

func (d *DB) DeleteUserFromLeague(ctx context.Context, leagueId, userId string) (*[]User, error) {
	user, err := d.GetUser(ctx, userId, "")
	if err != nil {
		return nil, errors.New("invalid user provided")
	}
	doc, err := d.leagueCollection().Doc(leagueId).Get(ctx)
	if err != nil {
		d.lg.Error("unable to fetch league", "error", err.Error())
		return nil, err
	}
	var league League
	err = doc.DataTo(&league)
	if err != nil {
		d.lg.Error("unable to parse league", "error", err.Error())
		return nil, err
	}
	userList := []User{}
	for _, u := range league.Users {
		if u.ID != user.ID {
			userList = append(userList, u)
		}
	}
	_, err = d.leagueCollection().Doc(leagueId).Update(ctx, []firestore.Update{
		{
			Path:  "Users",
			Value: userList,
		},
	})
	if err != nil {
		return nil, err
	}
	return &userList, nil
}

func (d *DB) UpsertWeekResults(ctx context.Context, leagueId string, week []espn.Game) ([]espn.Game, error) {
	doc, err := d.leagueCollection().Doc(leagueId).Get(ctx)
	if err != nil {
		d.lg.Error("unable to fetch league", "error", err.Error())
		return nil, err
	}
	var league League
	err = doc.DataTo(&league)
	if err != nil {
		d.lg.Error("unable to parse league", "error", err.Error())
		return nil, err
	}
	weekId := fmt.Sprintf("%d", week[0].Week)
	league.Weeks[weekId] = week
	_, err = d.leagueCollection().Doc(leagueId).Update(ctx, []firestore.Update{
		{
			Path:  "Weeks",
			Value: league.Weeks,
		},
	})
	if err != nil {
		return nil, err
	}
	return league.Weeks[weekId], nil
}
