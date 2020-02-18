// Package main はJankenプラグインの実装です
package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"golang.org/x/text/language"
)

const (
	maxHands         = 10
	defaultMaxRounds = 5
)

var handNames = []string{"rock", "scissors", "paper"}

// 英語名の手と日本語名の手の対応
var handNamesMap = map[string]string{
	"rock":     "Rock",
	"scissors": "Scissors",
	"paper":    "Paper",
}

// emojiとの対応
var handIcons = map[string]string{
	"rock":     ":fist_raised:",
	"scissors": ":v:",
	"paper":    ":hand:",
}

var newGameFuncMapping = map[string](func() gameInterface){
	"gameImpl1": newGameImpl1,
}

type participant struct {
	// MattermostのUserId
	UserID string `json:"user_id"`
	// N回戦目までに出す手
	Hands []string `json:"hands"`
	// 順位
	Rank int `json:"rank"`
}

func newParticipant(userID string) *participant {
	return &participant{
		UserID: userID,
		Hands:  make([]string, maxHands), // 事前に全要素を初期化
		Rank:   0,
	}
}

func (p *participant) setHands(hands []string) {
	copy(p.Hands, hands)
}

// clearHandsAfterは指定したi番目以降の手を空文字("")で初期化する
func (p *participant) clearHandsAfter(i int) {
	for j := i; j < len(p.Hands); j++ {
		p.Hands[j] = ""
	}
}

/*
GetHandはiで指定した手を返す．
未設定の場合はランダムな手を返す．このときの取得した値は保存される
*/
func (p *participant) getHand(i int) string {
	hand := p.Hands[i]
	if hand == "" {
		hand = handNames[rand.Intn(len(handNames))]
		p.Hands[i] = hand
	}
	return hand
}

type gameInterface interface {
	getResult(g *game) []*participant
}

type gameBase struct {
	base *game
}

func gameFromBytes(b []byte) (*game, error) {
	var tmp interface{}
	json.Unmarshal(b, &tmp)
	gameType := tmp.(map[string]interface{})["game_type"]
	if gameType == nil {
		return nil, errors.New("failed to get game type")
	}

	f := newGameFuncMapping[gameType.(string)]
	if f == nil {
		return nil, errors.New("failed to get function: " + gameType.(string))
	}
	impl := f()
	g := newGame(impl)
	json.Unmarshal(b, g)
	return g, nil
}

type game struct {
	// ID
	ID string `json:"id"`
	// 作成日時
	CreatedAt int64 `json:"created_at"`
	// 作成日時
	PostID string `json:"post_id"`
	// 作成者
	Creator string `json:"creator"`
	// 最大対戦回数
	MaxRounds int `json:"max_rounds"`
	// 最大参加人数
	MaxParticipants int `json:"max_participants"`
	// 参加者
	Participants []*participant `json:"participants"`
	Language     string         `json:"language"`
	GameType     string         `json:"game_type"`
	Impl         gameInterface  `json:"impl"`
}

func newGame(impl gameInterface) *game {
	g := &game{
		ID:           model.NewId(),
		CreatedAt:    model.GetMillis(),
		Creator:      "",
		MaxRounds:    defaultMaxRounds,
		Participants: make([]*participant, 0),
		Language:     language.English.String(),
	}
	g.Impl = impl
	g.setGameType(impl)
	return g
}

func (g *game) setGameType(impl gameInterface) {
	// ["*janken", "gameImpl1"]
	splitType := strings.Split(reflect.TypeOf(impl).String(), ".")
	// "gameImpl1"
	g.GameType = splitType[len(splitType)-1]
}

func (g *game) getResult() []*participant {
	return g.Impl.getResult(g)
}

// ToBytes returns byte slice of a game.
func (g *game) ToBytes() ([]byte, error) {
	b, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (g *game) getShortID() string {
	return g.ID[:7]
}

/*
GetParticipant は指定したuserIDのparticipantを返す．
一致するユーザーがいない場合はnilを返す
*/
func (g *game) GetParticipant(userID string) *participant {
	for _, p := range g.Participants {
		if p.UserID == userID {
			return p
		}
	}
	return nil
}

// RemoveParticipant は指定したuserIDのparticipantを削除する
func (g *game) RemoveParticipant(userID string) {
	participants := make([]*participant, 0)
	for _, p := range g.Participants {
		if p.UserID != userID {
			participants = append(participants, p)
		}
	}
	g.Participants = participants
}

/*
UpdateHands はparticipantのHandsを更新する．
指定したuserIDのparticipantが存在いない場合は新しく追加する．
*/
func (g *game) UpdateHands(userID string, hands []string) {
	for _, p := range g.Participants {
		if p.UserID == userID {
			p.Hands = hands
			return
		}
	}
	// 一致するユーザーがいない場合は新しく追加する
	p := newParticipant(userID)
	p.setHands(hands)
	g.Participants = append(g.Participants, p)
}

/*
ジャンケン1回の勝敗を判定する．
Args:
    participants: 参加者
    round: 何手目で勝負するか
Returns:
    []participant: 勝者
    []participant: 敗者
    []participant: あいこ
*/
func janken(participants []*participant, round int) ([]*participant, []*participant, []*participant) {
	// 手の種類とparticipantのmapを作る
	set := make(map[string][]*participant)
	for _, p := range participants {
		hand := p.getHand(round)
		if set[hand] == nil {
			set[hand] = []*participant{}
		}
		set[hand] = append(set[hand], p)
	}

	// 手の種類数が1か3のときはあいこ
	if len(set) == 1 || len(set) == 3 {
		// 参加者全員あいことして返す
		return nil, nil, participants
	}

	// それ以外(手の種類数==2)のとき
	// 手の種類を抽出してソートしておく
	hands := []string{}
	for h := range set {
		hands = append(hands, h)
	}
	sort.Slice(hands, func(i, j int) bool {
		return hands[i] < hands[j]
	})

	// 勝ちの手と負けの手を取得
	var win, lose string
	switch {
	case hands[0] == "rock" && hands[1] == "scissors":
		win, lose = "rock", "scissors"
	case hands[0] == "paper" && hands[1] == "scissors":
		win, lose = "scissors", "paper"
	case hands[0] == "paper" && hands[1] == "rock":
		win, lose = "paper", "rock"
	}

	// 勝者と敗者を返す
	var winners, losers []*participant
	winners = set[win]
	losers = set[lose]
	return winners, losers, nil
}
