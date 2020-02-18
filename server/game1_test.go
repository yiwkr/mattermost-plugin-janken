package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameImpl1(t *testing.T) {
	t.Run("newGameImpl1", func(t *testing.T) {
		expectedGame := &gameImpl1{}
		g := newGameImpl1()
		assert.Equal(t, expectedGame, g)
	})

	t.Run("getResult", func(t *testing.T) {
		for name, test := range map[string]struct {
			MaxRounds     int
			participants  []*participant
			ExpectedRanks map[string]int
		}{
			"get result successfully": {
				MaxRounds: 2,
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
					{UserID: "p2", Hands: []string{"rock", "paper"}},
					{UserID: "p3", Hands: []string{"scissors", "paper"}},
					{UserID: "p4", Hands: []string{"scissors", "rock"}},
				},
				ExpectedRanks: map[string]int{"p1": 1, "p2": 2, "p3": 3, "p4": 4},
			},
			"the all of rounds are drawn": {
				MaxRounds: 2,
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "rock"}},
					{UserID: "p2", Hands: []string{"rock", "scissors"}},
					{UserID: "p3", Hands: []string{"rock", "paper"}},
				},
				ExpectedRanks: map[string]int{"p1": 1, "p2": 1, "p3": 1},
			},
			"1 winner and 2 drawers": {
				MaxRounds: 2,
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "rock"}},
					{UserID: "p2", Hands: []string{"scissors", "rock"}},
					{UserID: "p3", Hands: []string{"scissors", "rock"}},
				},
				ExpectedRanks: map[string]int{"p1": 1, "p2": 2, "p3": 2},
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)
				g := &gameImpl1{}
				b := newGame(g)
				b.MaxRounds = test.MaxRounds
				b.Participants = test.participants

				result := g.getResult(b)

				for _, r := range result {
					assert.Equal(test.ExpectedRanks[r.UserID], r.Rank)
				}
			})
		}
	})

	t.Run("nextRound", func(t *testing.T) {
		for name, test := range map[string]struct {
			MaxRounds            int
			participants         []*participant
			Round                int
			Rank                 int
			ExpectedParticipants []*participant
		}{
			"stop condition 1": {
				MaxRounds: 2,
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "rock"}},
				},
				Round: 1,
				Rank:  1,
				ExpectedParticipants: []*participant{
					{UserID: "p1", Hands: []string{"rock", ""}, Rank: 1},
				},
			},
			"stop condition 2": {
				MaxRounds: 2,
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "rock"}},
					{UserID: "p2", Hands: []string{"rock", "rock"}},
				},
				Round: 2,
				Rank:  2,
				ExpectedParticipants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "rock"}, Rank: 2},
					{UserID: "p2", Hands: []string{"rock", "rock"}, Rank: 2},
				},
			},
			"draw condition": {
				MaxRounds: 1,
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock"}},
					{UserID: "p2", Hands: []string{"rock"}},
				},
				Round: 0,
				Rank:  1,
				ExpectedParticipants: []*participant{
					{UserID: "p1", Hands: []string{"rock"}, Rank: 1},
					{UserID: "p2", Hands: []string{"rock"}, Rank: 1},
				},
			},
			"winner and loser condition": {
				MaxRounds: 1,
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock"}},
					{UserID: "p2", Hands: []string{"scissors"}},
				},
				Round: 0,
				Rank:  1,
				ExpectedParticipants: []*participant{
					{UserID: "p1", Hands: []string{"rock"}, Rank: 1},
					{UserID: "p2", Hands: []string{"scissors"}, Rank: 2},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)
				g := &gameImpl1{}

				result := g.nextRound(test.participants, test.MaxRounds, test.Round, test.Rank, nil)

				for i := range result {
					assert.Equal(test.ExpectedParticipants[i], test.participants[i])
				}
			})
		}
	})
}
