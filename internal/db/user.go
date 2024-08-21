package db

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type User struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Picks   []Pick   `json:"picks"`
	Leagues []League `json:"leagues"`
}

type UserList struct {
	Users []User `json:"users"`
}

func (d *DB) userCollection() *firestore.CollectionRef {
	return d.Client.Collection("user")
}

func (d *DB) GetAllUsers(ctx context.Context) (*UserList, error) {
	users := UserList{Users: []User{}}
	iter := d.userCollection().Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			d.lg.Error("unable to fetch user", "error", err.Error())
			return nil, err
		}
		var user User
		if err := doc.DataTo(&user); err != nil {
			d.lg.Error("unable to parse user", "error", err.Error())
			return nil, err
		}
		users.Users = append(users.Users, user)
	}
	return &users, nil
}

func (d *DB) CreateUser(ctx context.Context, user User) (*User, error) {
	_, err := d.userCollection().Doc(user.ID).Set(ctx, user)
	if err != nil {
		d.lg.Error("unable to create user", "error", err.Error())
		return nil, err
	}
	return &user, nil
}

func (d *DB) GetUser(ctx context.Context, id string) (*User, error) {
	doc, err := d.userCollection().Doc(id).Get(ctx)
	if err != nil {
		d.lg.Error("unable to fetch user", "error", err.Error())
		return nil, err
	}
	var user User
	err = doc.DataTo(&user)
	if err != nil {
		d.lg.Error("unable to parse user", "error", err.Error())
		return nil, err
	}
	return &user, nil
}
