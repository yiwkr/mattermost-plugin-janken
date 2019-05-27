package janken

import (
	"fmt"

	"github.com/mattermost/mattermost-server/plugin"
)

const (
	// データ保存期間(1週間=60秒*60分*24時間*7日間=604800秒)
	expireInSeconds int64  = 604800

	KeyPrefix       string = "janken_"
)

type Store struct {
	API          plugin.API
	jankenStore  JankenStoreInterface
}

func NewStore(api plugin.API) *Store {
	store := Store{
		API: api,
		jankenStore: JankenStore{
			API: api,
		},
	}
	return &store
}

type JankenStoreInterface interface {
	Get(string) (*JankenGame, error)
	Save(*JankenGame) error
	Delete(string) error
}

type JankenStore struct {
	API plugin.API
}
/*
Getは指定したidに対応するJankenGameを取得する．
指定したidが存在しない場合、新しい構造体を返す．
*/
func (s JankenStore) Get(id string) (*JankenGame, error) {
	b, appErr := s.API.KVGet(KeyPrefix+id)
	if appErr != nil {
		return nil, appErr
	}

	game, err := JankenGameFromBytes(b)
	if err != nil {
		return nil, err
	}
	return game, nil
}

func (s JankenStore) Save(game *JankenGame) error {
	gameId := game.Id
	s.API.LogDebug("Save", "id", gameId, "game", fmt.Sprintf("%#v", game))
	b, err := game.ToBytes()
	if err != nil {
		return err
	}
	if err := s.API.KVSetWithExpiry(KeyPrefix+gameId, b, expireInSeconds); err != nil {
		return err
	}
	return nil
}

func (s JankenStore) Delete(id string) error {
	s.API.LogDebug("Delete", "id", id)
	return s.API.KVDelete(KeyPrefix+id)
}
