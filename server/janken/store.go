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
	API    plugin.API
	jankenStore  JankenStore
}

func NewStore(api plugin.API) (*Store, error) {
	store := Store{
		jankenStore: JankenStore{
			API: api,
		},
	}
	return &store, nil
}

type JankenStore struct {
	API plugin.API
}

/*
Getは指定したidに対応するJankenGameを取得する．
指定したidが存在しない場合、新しい構造体を返す．
*/
func (s *JankenStore) Get(id string) *JankenGame {
	b, _ := s.API.KVGet(KeyPrefix+id)
	game := JankenGameFromBytes(b)
	s.API.LogDebug("Get", "id", id, "game", fmt.Sprintf("%#v", game))
	return game
}

func (s *JankenStore) Save(game *JankenGame) {
	s.API.LogDebug("Save", "id", game.Id, "game", fmt.Sprintf("%#v", game))
	b := game.ToBytes()
	_ = s.API.KVSetWithExpiry(KeyPrefix+game.Id, b, expireInSeconds)
}

func (s *JankenStore) Delete(id string) {
	s.API.LogDebug("Delete", "id", id)
	_ = s.API.KVDelete(KeyPrefix+id)
}
