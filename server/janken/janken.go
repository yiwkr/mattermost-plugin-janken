// Package mattermost-plugin-janken
package janken

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"golang.org/x/text/language"
)

const (
	MAX_HANDS = 10
	DEFAULT_MAX_ROUNDS = 5
)

// 英語名の手と日本語名の手の対応
var Hands = map[string]string{
	"rock": "Rock",
	"scissors": "Scissors",
	"paper": "Paper",
}

// emojiとの対応
var HandIcons = map[string]string{
	"rock": ":fist_raised:",
	"scissors": ":v:",
	"paper": ":hand:",
}

var NewJankenGameFuncMapping = map[string](func() JankenGameInterface){
	"JankenGameImpl1": NewJankenGameImpl1,
}

type Participant struct {
	// MattermostのUserId
	UserId          string         `json:"user_id"`
	// N回戦目までに出す手
	Hands           []string       `json:"hands"`
	// 順位
	Rank            int            `json:"rank"`
}

func NewParticipant(userId string) *Participant {
	return &Participant{
		UserId: userId,
		Hands: make([]string, MAX_HANDS),  // 事前に全要素を初期化
		Rank: 0,
	}
}

func (p *Participant) SetHands(hands []string) {
	copy(p.Hands, hands)
}

// ClearHandsAfterは指定したi番目以降の手を空文字("")で初期化する
func (p *Participant) ClearHandsAfter(i int) {
	for j := i; j<len(p.Hands); j++ {
		p.Hands[j] = ""
	}
}

/*
GetHandはiで指定した手を返す．
未設定の場合はランダムな手を返す．このときの取得した値は保存される
*/
func (p *Participant) GetHand(i int) string {
	hand := p.Hands[i]

	if hand == "" {
		/*
		Goのmapは実行ごとに異なる順番で要素を取り出すため
		Handsの最初の1つを取り出せばランダムに手を取得できる
		*/
		for h := range Hands {
			hand = h
			break
		}
		p.Hands[i] = hand
	}

	return hand
}

type JankenGameInterface interface {
	GetResult(g *JankenGame) []*Participant
}

type JankenGameBase struct {
	base *JankenGame
}

func JankenGameFromBytes(b []byte) (*JankenGame, error) {
	var tmp interface{}
	json.Unmarshal(b, &tmp)
	gameType := tmp.(map[string]interface{})["game_type"]
	if gameType == nil {
		return nil, errors.New("failed to get game type")
	}

	f := NewJankenGameFuncMapping[gameType.(string)]
	if f == nil {
		return nil, errors.New("failed to get function: "+gameType.(string))
	}
	impl := f()
	g := NewJankenGame(impl)
	json.Unmarshal(b, g)
	return g, nil
}

type JankenGame struct {
	// ID
	Id              string              `json:"id"`
	// 作成日時
	CreatedAt       int64               `json:"created_at"`
	// 作成日時
	PostId          string              `json:"post_id"`
	// 作成者
	Creator         string              `json:"creator"`
	// 最大対戦回数
	MaxRounds       int                 `json:"max_rounds"`
	// 最大参加人数
	MaxParticipants int                 `json:"max_participants"`
	// 参加者
	Participants    []*Participant      `json:"participants"`
	Language        string              `json:"language"`
	GameType        string              `json:"game_type"`
	Impl            JankenGameInterface `json:"impl"`
}

func NewJankenGame(impl JankenGameInterface) *JankenGame {
	g := &JankenGame{
		Id: model.NewId(),
		CreatedAt: model.GetMillis(),
		Creator: "",
		MaxRounds: DEFAULT_MAX_ROUNDS,
		Participants: make([]*Participant, 0),
		Language: language.English.String(),
	}
	g.Impl = impl
	g.SetGameType(impl)
	return g
}

func (g *JankenGame) SetGameType(impl JankenGameInterface) {
	// ["*janken", "JankenGameImpl1"]
	split_type := strings.Split(reflect.TypeOf(impl).String(), ".")
	// "JankenGameImpl1"
	g.GameType = split_type[len(split_type) - 1]
}

func (g *JankenGame) GetResult() []*Participant {
	return g.Impl.GetResult(g)
}

func (g *JankenGame) ToBytes() ([]byte, error) {
	b, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (g *JankenGame) GetShortId() string {
	return g.Id[:7]
}

/*
GetHandsByUserIdは指定したuserIdのParticipantを返す．
一致するユーザーがいない場合はnilを返す
*/
func (g *JankenGame) GetParticipant(userId string) *Participant {
	for _, p := range g.Participants {
		if p.UserId == userId {
			return p
		}
	}
	return nil
}

// RemoveParticipantは指定したuserIdのParticipantを削除する
func (g *JankenGame) RemoveParticipant(userId string) {
	participants := make([]*Participant, 0)
	for _, p := range g.Participants {
		if p.UserId != userId {
			participants = append(participants, p)
		}
	}
	g.Participants = participants
}

/*
UpdateHandsはParticipantのHandsを更新する．
指定したuserIdのParticipantが存在いない場合は新しく追加する．
*/
func (g *JankenGame) UpdateHands(userId string, hands []string) {
	for _, p := range g.Participants {
		if p.UserId == userId {
			p.Hands = hands
			return
		}
	}
	// 一致するユーザーがいない場合は新しく追加する
	participant := NewParticipant(userId)
	participant.SetHands(hands)
	g.Participants = append(g.Participants, participant)
}

/*
ジャンケン1回の勝敗を判定する．
Args:
    participants: 参加者
    round: 何手目で勝負するか
Returns:
    []Participant: 勝者
    []Participant: 敗者
    []Participant: あいこ
*/
func janken(participants []*Participant, round int) ([]*Participant, []*Participant, []*Participant) {
	// 手の種類とParticipantのmapを作る
	set := make(map[string][]*Participant)
	for _, p := range participants {
		hand := p.GetHand(round)
		if set[hand] == nil {
			set[hand] = []*Participant{}
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
	for h, _ := range set {
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
	var winners, losers []*Participant
	winners = set[win]
	losers = set[lose]
	return winners, losers, nil
}
