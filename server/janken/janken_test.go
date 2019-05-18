package janken

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestNewJankenGame(t *testing.T) {
	assert := assert.New(t)

	game := NewJankenGame()

	assert.Equal("", game.Creator)
	assert.Equal(5, game.MaxRounds)
	assert.Equal(make([]*Participant, 0), game.Participants)
}

func TestGetResult(t *testing.T) {
	for name, test := range map[string]struct {
		Game          JankenGame
		ExpectedRanks map[string]int
	}{
		"Normal": {
			Game: JankenGame{
				MaxRounds : 2,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock",     "scissors"}},
					{UserId: "p2", Hands: []string{"rock",     "paper"}},
					{UserId: "p3", Hands: []string{"scissors", "paper"}},
					{UserId: "p4", Hands: []string{"scissors", "rock"}},
				},
			},
			ExpectedRanks: map[string]int{"p1": 1, "p2": 2, "p3": 3, "p4": 4},
		},
		"Draw the all round": {
			Game: JankenGame{
				MaxRounds : 2,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "rock"}},
					{UserId: "p2", Hands: []string{"rock", "scissors"}},
					{UserId: "p3", Hands: []string{"rock", "paper"}},
				},
			},
			ExpectedRanks: map[string]int{"p1": 1, "p2": 1, "p3": 1},
		},
		"1 winner and 2 drawer": {
			Game: JankenGame{
				MaxRounds : 2,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock",     "rock"}},
					{UserId: "p2", Hands: []string{"scissors", "rock"}},
					{UserId: "p3", Hands: []string{"scissors", "rock"}},
				},
			},
			ExpectedRanks: map[string]int{"p1": 1, "p2": 2, "p3": 2},
		},
	}{
		t.Run(name, func(t *testing.T){
			assert := assert.New(t)
			result := test.Game.GetResult()
			for _, r := range result {
				assert.Equal(test.ExpectedRanks[r.UserId], r.Rank)
			}
		})
	}
}
