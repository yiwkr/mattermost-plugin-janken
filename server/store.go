package main

import (
	"errors"
	"fmt"

	"github.com/mattermost/mattermost-server/v5/plugin"
)

const (
	// expireInSeconds is expiry time. 604800 seconds equal to 7 days)
	expireInSeconds int64 = 604800

	// keyPrefix is store key prefix
	keyPrefix string = "janken_"
)

// Store is an interface to interact with the KV store.
type Store struct {
	API         plugin.API
	jankenStore jankenStoreInterface
}

// NewStore returns the new Store
func NewStore(api plugin.API) *Store {
	store := Store{
		API: api,
		jankenStore: jankenStore{
			API: api,
		},
	}
	return &store
}

// jankenStoreInterface allows to access janken games in the KV store.
type jankenStoreInterface interface {
	Get(string) (*game, error)
	Save(*game) error
	Delete(string) error
}

// jankenStore allows to access janken games in the KV store.
type jankenStore struct {
	API plugin.API
}

// Get returns the janken game for a given id. Returns an error if the janken game doesn't exist or a KV store error occured.
func (s jankenStore) Get(id string) (*game, error) {
	b, appErr := s.API.KVGet(keyPrefix + id)
	if appErr != nil {
		return nil, appErr
	}

	game, err := gameFromBytes(b)
	if err != nil {
		return nil, err
	}
	return game, nil
}

// Save creates or updates a janken game for a given id.
func (s jankenStore) Save(game *game) error {
	gameID := game.ID
	s.API.LogDebug("Save", "id", gameID, "game", fmt.Sprintf("%#v", game))
	b, err := game.ToBytes()
	if err != nil {
		return err
	}
	appErr := s.API.KVSetWithExpiry(keyPrefix+gameID, b, expireInSeconds);
	if appErr != nil {
		return errors.New(appErr.DetailedError)
	}
	return nil
}

// Delete deletes a janken game from the KV store.
func (s jankenStore) Delete(id string) error {
	s.API.LogDebug("Delete", "id", id)
	return s.API.KVDelete(keyPrefix + id)
}
