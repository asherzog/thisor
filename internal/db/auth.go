package db

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

type SessionManager struct {
	Sessions map[string]string
}

func (d *DB) NewSessionManager() *SessionManager {
	return &SessionManager{Sessions: map[string]string{}}
}

func (s *SessionManager) NewSession() (string, error) {
	state, err := generateRandomState()
	if err != nil {
		return "", err
	}
	s.Sessions[state] = state
	return state, nil
}

func (s *SessionManager) GetSession(state string) (string, error) {
	session := s.Sessions[state]
	if session == "" {
		return "", errors.New("invalid session")
	}
	return state, nil
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.StdEncoding.EncodeToString(b)
	return state, nil
}
