package janken

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJankenGameImpl1(t *testing.T) {
	t.Run("GetResult", func(t *testing.T){
		for name, test := range map[string]struct {
			MaxRounds     int
			Participants  []*Participant
			ExpectedRanks map[string]int
		}{
			"get result successfully": {
				MaxRounds: 2,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock",     "scissors"}},
					{UserId: "p2", Hands: []string{"rock",     "paper"}},
					{UserId: "p3", Hands: []string{"scissors", "paper"}},
					{UserId: "p4", Hands: []string{"scissors", "rock"}},
				},
				ExpectedRanks: map[string]int{"p1": 1, "p2": 2, "p3": 3, "p4": 4},
			},
			"the all of rounds are drawn": {
				MaxRounds: 2,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "rock"}},
					{UserId: "p2", Hands: []string{"rock", "scissors"}},
					{UserId: "p3", Hands: []string{"rock", "paper"}},
				},
				ExpectedRanks: map[string]int{"p1": 1, "p2": 1, "p3": 1},
			},
			"1 winner and 2 drawers": {
				MaxRounds: 2,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock",     "rock"}},
					{UserId: "p2", Hands: []string{"scissors", "rock"}},
					{UserId: "p3", Hands: []string{"scissors", "rock"}},
				},
				ExpectedRanks: map[string]int{"p1": 1, "p2": 2, "p3": 2},
			},
		}{
			t.Run(name, func(t *testing.T){
				assert := assert.New(t)
				g := &JankenGameImpl1{}
				b := NewJankenGame(g)
				b.MaxRounds = test.MaxRounds
				b.Participants = test.Participants

				result := g.GetResult(b)

				for _, r := range result {
					assert.Equal(test.ExpectedRanks[r.UserId], r.Rank)
				}
			})
		}
	})

	t.Run("nextRound", func(t *testing.T){
		for name, test := range map[string]struct {
			MaxRounds            int
			Participants         []*Participant
			Round                int
			Rank                 int
			ExpectedParticipants []*Participant
		}{
			"stop condition 1": {
				MaxRounds: 2,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "rock"}},
				},
				Round: 1,
				Rank: 1,
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", ""}, Rank: 1},
				},
			},
			"stop condition 2": {
				MaxRounds: 2,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "rock"}},
					{UserId: "p2", Hands: []string{"rock", "rock"}},
				},
				Round: 2,
				Rank: 2,
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "rock"}, Rank: 2},
					{UserId: "p2", Hands: []string{"rock", "rock"}, Rank: 2},
				},
			},
			"draw condition": {
				MaxRounds: 1,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock"}},
					{UserId: "p2", Hands: []string{"rock"}},
				},
				Round: 0,
				Rank: 1,
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Hands: []string{"rock"}, Rank: 1},
					{UserId: "p2", Hands: []string{"rock"}, Rank: 1},
				},
			},
			"winner and loser condition": {
				MaxRounds: 1,
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock"}},
					{UserId: "p2", Hands: []string{"scissors"}},
				},
				Round: 0,
				Rank: 1,
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Hands: []string{"rock"}, Rank: 1},
					{UserId: "p2", Hands: []string{"scissors"}, Rank: 2},
				},
			},
		}{
			t.Run(name, func(t *testing.T){
				assert := assert.New(t)
				g := &JankenGameImpl1{}

				result := g.nextRound(test.Participants, test.MaxRounds, test.Round, test.Rank, nil)

				for i, _ := range result {
					assert.Equal(test.ExpectedParticipants[i], test.Participants[i])
				}
			})
		}
	})
}
