package main

import (
	"encoding/json"
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

type TestGameImpl struct {
	gameBase
}

func newTestGameImpl() gameInterface {
	return &TestGameImpl{}
}

func (g *TestGameImpl) getResult(game *game) []*participant {
	return []*participant{
		&participant{UserID: "p1", Rank: 1},
		&participant{UserID: "p2", Rank: 2},
	}
}

func TestParticipant(t *testing.T) {
	t.Run("NewParticipant", func(t *testing.T) {
		for name, test := range map[string]struct {
			UserID              string
			ExpectedParticipant *participant
		}{
			"create NewParticipant successfully": {
				UserID: "p1",
				ExpectedParticipant: &participant{
					UserID: "p1",
					Hands:  make([]string, 10),
					Rank:   0,
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				p := newParticipant(test.UserID)

				assert.Equal(test.ExpectedParticipant, p)
			})
		}
	})

	t.Run("GetHand", func(t *testing.T) {
		for name, test := range map[string]struct {
			Index         int
			participant   *participant
			ExpectedHands []string
		}{
			"get a hand successfully": {
				Index: 0,
				participant: &participant{
					UserID: "p1",
					Hands:  []string{"rock"},
				},
				ExpectedHands: []string{"rock"},
			},
			"get a random hand successfully": {
				Index: 0,
				participant: &participant{
					UserID: "p1",
					Hands:  []string{""},
				},
				ExpectedHands: []string{"rock", "scissors", "paper"},
			},
		} {
			t.Run(name, func(t *testing.T) {
				p := test.participant

				h := p.getHand(test.Index)

				assert := assert.New(t)
				assert.Contains(test.ExpectedHands, h)
			})
		}
	})
}

func TestGame(t *testing.T) {
	t.Run("newGame", func(t *testing.T) {
		t.Run("generate a new game successfully", func(t *testing.T) {
			assert := assert.New(t)

			g := newGame(&TestGameImpl{})

			assert.Equal("", g.Creator)
			assert.Equal(5, g.MaxRounds)
			assert.Equal(make([]*participant, 0), g.Participants)
			assert.Equal("en", g.Language)
			assert.Equal("TestGameImpl", g.GameType)
		})
	})

	t.Run("gameFromBytes", func(t *testing.T) {
		for name, test := range map[string]struct {
			Bytes              []byte
			ExpectedGame *game
			ShouldError        bool
		}{
			"successfully": {
				Bytes:              []byte(`{"game_type":"TestGameImpl"}`),
				ExpectedGame: &game{Impl: newTestGameImpl()},
				ShouldError:        false,
			},
			"game_type missing": {
				Bytes:              []byte("{}"),
				ExpectedGame: nil,
				ShouldError:        true,
			},
			"Invalid game_type": {
				Bytes:              []byte(`{"game_type":"InvalidGameType"}`),
				ExpectedGame: nil,
				ShouldError:        true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				newGameFuncMapping = map[string](func() gameInterface){
					"TestGameImpl": newTestGameImpl,
				}

				g, err := gameFromBytes(test.Bytes)

				assert := assert.New(t)
				if test.ExpectedGame != nil && g != nil {
					assert.Equal(test.ExpectedGame.Impl, g.Impl)
				} else {
					assert.Equal(test.ExpectedGame, g)
				}

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})

	t.Run("getResult", func(t *testing.T) {
		for name, test := range map[string]struct {
			ExpectedParticipants []*participant
		}{
			"pariticipant not found": {
				ExpectedParticipants: []*participant{
					{UserID: "p1", Rank: 1},
					{UserID: "p2", Rank: 2},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)
				g := newGame(&TestGameImpl{})

				r := g.getResult()

				assert.Equal(test.ExpectedParticipants, r)
			})
		}
	})

	t.Run("ToBytes", func(t *testing.T) {
		for name, test := range map[string]struct {
			PatchMarshal  func() *monkey.PatchGuard
			ExpectedBytes []byte
			ShouldError   bool
		}{
			"convert to bytes successfully": {
				PatchMarshal: func() *monkey.PatchGuard {
					return monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
						return []byte(`{}`), nil
					})
				},
				ExpectedBytes: []byte(`{}`),
				ShouldError:   false,
			},
			"failed because json.Marshal returns error": {
				PatchMarshal: func() *monkey.PatchGuard {
					return monkey.Patch(json.Marshal, func(interface{}) ([]byte, error) {
						return nil, errors.New("Marshal error")
					})
				},
				ExpectedBytes: nil,
				ShouldError:   true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				patch := test.PatchMarshal()
				defer patch.Unpatch()

				g := game{}

				b, err := g.ToBytes()

				assert := assert.New(t)
				assert.Equal(test.ExpectedBytes, b)

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})

	t.Run("getShortID", func(t *testing.T) {
		assert := assert.New(t)

		g := game{}
		g.ID = "abcdefghijklmnopqrstuvwxyz"
		assert.Equal("abcdefg", g.getShortID())
	})

	t.Run("GetParticipant", func(t *testing.T) {
		for name, test := range map[string]struct {
			UserID              string
			participants        []*participant
			ExpectedParticipant *participant
		}{
			"pariticipant not found": {
				UserID: "unknown",
				participants: []*participant{
					{UserID: "p1"},
					{UserID: "p2"},
					{UserID: "p3"},
				},
				ExpectedParticipant: nil,
			},
			"get pariticipant successfully": {
				UserID: "p2",
				participants: []*participant{
					{UserID: "p1"},
					{UserID: "p2"},
					{UserID: "p3"},
				},
				ExpectedParticipant: &participant{UserID: "p2"},
			},
		} {
			t.Run(name, func(t *testing.T) {
				g := game{}
				g.Participants = test.participants

				p := g.GetParticipant(test.UserID)

				assert := assert.New(t)
				assert.Equal(test.ExpectedParticipant, p)
			})
		}
	})

	t.Run("RemoveParticipant", func(t *testing.T) {
		for name, test := range map[string]struct {
			UserID               string
			participants         []*participant
			ExpectedParticipants []*participant
		}{
			"participant not found": {
				UserID: "unknown",
				participants: []*participant{
					{UserID: "p1"},
					{UserID: "p2"},
					{UserID: "p3"},
				},
				ExpectedParticipants: []*participant{
					{UserID: "p1"},
					{UserID: "p2"},
					{UserID: "p3"},
				},
			},
			"remove participant successufully": {
				UserID: "p2",
				participants: []*participant{
					{UserID: "p1"},
					{UserID: "p2"},
					{UserID: "p3"},
				},
				ExpectedParticipants: []*participant{
					{UserID: "p1"},
					{UserID: "p3"},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				g := game{}
				g.Participants = test.participants

				g.RemoveParticipant(test.UserID)

				assert := assert.New(t)
				assert.Equal(test.ExpectedParticipants, g.Participants)
			})
		}
	})

	t.Run("UpdateHands", func(t *testing.T) {
		for name, test := range map[string]struct {
			UserID               string
			Hands                []string
			participants         []*participant
			ExpectedParticipants []*participant
		}{
			"update hands successfully": {
				UserID: "p2",
				Hands:  []string{"paper", "scissors"},
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
					{UserID: "p2", Hands: []string{"rock", "scissors"}},
					{UserID: "p3", Hands: []string{"rock", "scissors"}},
				},
				ExpectedParticipants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
					{UserID: "p2", Hands: []string{"paper", "scissors"}},
					{UserID: "p3", Hands: []string{"rock", "scissors"}},
				},
			},
			"update by short length hands successfully": {
				UserID: "p2",
				Hands:  []string{"paper"},
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
					{UserID: "p2", Hands: []string{"rock", "scissors"}},
					{UserID: "p3", Hands: []string{"rock", "scissors"}},
				},
				ExpectedParticipants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
					{UserID: "p2", Hands: []string{"paper"}},
					{UserID: "p3", Hands: []string{"rock", "scissors"}},
				},
			},
			"update by long length hands successfully": {
				UserID: "p2",
				Hands:  []string{"paper", "scissors", "rock"},
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
					{UserID: "p2", Hands: []string{"rock", "scissors"}},
					{UserID: "p3", Hands: []string{"rock", "scissors"}},
				},
				ExpectedParticipants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
					{UserID: "p2", Hands: []string{"paper", "scissors", "rock"}},
					{UserID: "p3", Hands: []string{"rock", "scissors"}},
				},
			},
			"add new participant": {
				UserID: "p2",
				Hands:  []string{"paper", "scissors"},
				participants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
				},
				ExpectedParticipants: []*participant{
					{UserID: "p1", Hands: []string{"rock", "scissors"}},
					{UserID: "p2", Hands: []string{"paper", "scissors", "", "", "", "", "", "", "", ""}},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				g := game{}
				g.Participants = test.participants

				g.UpdateHands(test.UserID, test.Hands)

				assert := assert.New(t)
				assert.Equal(test.ExpectedParticipants, g.Participants)
			})
		}
	})
}

func TestJanken(t *testing.T) {
	for name, test := range map[string]struct {
		Round           int
		participants    []*participant
		ExpectedWinners []*participant
		ExpectedLosers  []*participant
		ExpectedDrawers []*participant
	}{
		"2 drawers": {
			Round: 0,
			participants: []*participant{
				{UserID: "p1", Hands: []string{"rock"}},
				{UserID: "p2", Hands: []string{"rock"}},
			},
			ExpectedWinners: nil,
			ExpectedLosers:  nil,
			ExpectedDrawers: []*participant{
				{UserID: "p1", Hands: []string{"rock"}},
				{UserID: "p2", Hands: []string{"rock"}},
			},
		},
		"3 drawers with same hands": {
			Round: 0,
			participants: []*participant{
				{UserID: "p1", Hands: []string{"rock"}},
				{UserID: "p2", Hands: []string{"rock"}},
				{UserID: "p3", Hands: []string{"rock"}},
			},
			ExpectedWinners: nil,
			ExpectedLosers:  nil,
			ExpectedDrawers: []*participant{
				{UserID: "p1", Hands: []string{"rock"}},
				{UserID: "p2", Hands: []string{"rock"}},
				{UserID: "p3", Hands: []string{"rock"}},
			},
		},
		"3 drawers with different hands": {
			Round: 0,
			participants: []*participant{
				{UserID: "p1", Hands: []string{"rock"}},
				{UserID: "p2", Hands: []string{"scissors"}},
				{UserID: "p3", Hands: []string{"paper"}},
			},
			ExpectedWinners: nil,
			ExpectedLosers:  nil,
			ExpectedDrawers: []*participant{
				{UserID: "p1", Hands: []string{"rock"}},
				{UserID: "p2", Hands: []string{"scissors"}},
				{UserID: "p3", Hands: []string{"paper"}},
			},
		},
		"2 winner and 1 loser": {
			Round: 0,
			participants: []*participant{
				{UserID: "p1", Hands: []string{"paper"}},
				{UserID: "p2", Hands: []string{"paper"}},
				{UserID: "p3", Hands: []string{"rock"}},
			},
			ExpectedWinners: []*participant{
				{UserID: "p1", Hands: []string{"paper"}},
				{UserID: "p2", Hands: []string{"paper"}},
			},
			ExpectedLosers: []*participant{
				{UserID: "p3", Hands: []string{"rock"}},
			},
			ExpectedDrawers: nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			winners, losers, drawers := janken(test.participants, test.Round)

			assert.Equal(test.ExpectedWinners, winners)
			assert.Equal(test.ExpectedLosers, losers)
			assert.Equal(test.ExpectedDrawers, drawers)
		})
	}
}
