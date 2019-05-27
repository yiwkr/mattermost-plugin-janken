package janken

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestJankenGameImpl struct {
	JankenGameBase
}
func (g *TestJankenGameImpl) GetResult(game *JankenGame) []*Participant {
	return []*Participant{
		&Participant{UserId: "p1", Rank: 1},
		&Participant{UserId: "p2", Rank: 2},
	}
}

func TestParticipant(t *testing.T) {
	t.Run("NewParticipant", func(t *testing.T){
		for name, test := range map[string]struct {
			UserId              string
			ExpectedParticipant *Participant
		}{
			"create NewParticipant successfully": {
				UserId: "p1",
				ExpectedParticipant: &Participant{
					UserId: "p1",
					Hands: make([]string, 10),
					Rank: 0,
				},
			},
		}{
			t.Run(name, func(t *testing.T){
				assert := assert.New(t)

				p := NewParticipant(test.UserId)

				assert.Equal(test.ExpectedParticipant, p)
			})
		}
	})

	t.Run("GetHand", func(t *testing.T){
		for name, test := range map[string]struct {
			Index         int
			Participant   *Participant
			ExpectedHands []string
		}{
			"get a hand successfully": {
				Index: 0,
				Participant: &Participant{
					UserId: "p1",
					Hands: []string{"rock"},
				},
				ExpectedHands: []string{"rock"},
			},
			"get a random hand successfully": {
				Index: 0,
				Participant: &Participant{
					UserId: "p1",
					Hands: []string{""},
				},
				ExpectedHands: []string{"rock", "scissors", "paper"},
			},
		}{
			t.Run(name, func(t *testing.T){
				p := test.Participant

				h := p.GetHand(test.Index)

				assert := assert.New(t)
				assert.Contains(test.ExpectedHands, h)
			})
		}
	})
}

func TestJankenGame(t *testing.T) {
	t.Run("NewJankenGame", func(t *testing.T){
		t.Run("generate a new JankenGame successfully", func(t *testing.T){
			assert := assert.New(t)

			g := NewJankenGame(&TestJankenGameImpl{})

			assert.Equal("", g.Creator)
			assert.Equal(5, g.MaxRounds)
			assert.Equal(make([]*Participant, 0), g.Participants)
			assert.Equal("en", g.Language)
			assert.Equal("TestJankenGameImpl", g.GameType)
		})
	})

	t.Run("GetResult", func(t *testing.T){
		for name, test := range map[string]struct {
			ExpectedParticipants []*Participant
		}{
			"pariticipant not found": {
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Rank: 1},
					{UserId: "p2", Rank: 2},
				},
			},
		}{
			t.Run(name, func(t *testing.T){
				assert := assert.New(t)
				g := NewJankenGame(&TestJankenGameImpl{})

				r := g.GetResult()

				assert.Equal(test.ExpectedParticipants, r)
			})
		}
	})

	t.Run("ToBytes", func(t *testing.T){
		t.Run("convert to bytes successfully", func(t *testing.T){
			assert := assert.New(t)

			g := JankenGame{}

			b, _ := g.ToBytes()
			expect, _ := json.Marshal(g)
			assert.Equal(expect, b)
		})
	})

	t.Run("GetShortId", func(t *testing.T){
		assert := assert.New(t)

		g := JankenGame{}
		g.Id = "abcdefghijklmnopqrstuvwxyz"
		assert.Equal("abcdefg", g.GetShortId())
	})

	t.Run("GetParticipant", func(t *testing.T){
		for name, test := range map[string]struct {
			UserId              string
			Participants        []*Participant
			ExpectedParticipant *Participant
		}{
			"pariticipant not found": {
				UserId: "unknown",
				Participants: []*Participant{
					{UserId: "p1"},
					{UserId: "p2"},
					{UserId: "p3"},
				},
				ExpectedParticipant: nil,
			},
			"get pariticipant successfully": {
				UserId: "p2",
				Participants: []*Participant{
					{UserId: "p1"},
					{UserId: "p2"},
					{UserId: "p3"},
				},
				ExpectedParticipant: &Participant{UserId: "p2"},
			},
		}{
			t.Run(name, func(t *testing.T){
				g := JankenGame{}
				g.Participants = test.Participants

				p := g.GetParticipant(test.UserId)

				assert := assert.New(t)
				assert.Equal(test.ExpectedParticipant, p)
			})
		}
	})

	t.Run("RemoveParticipant", func(t *testing.T){
		for name, test := range map[string]struct {
			UserId               string
			Participants         []*Participant
			ExpectedParticipants []*Participant
		}{
			"participant not found": {
				UserId: "unknown",
				Participants: []*Participant{
					{UserId: "p1"},
					{UserId: "p2"},
					{UserId: "p3"},
				},
				ExpectedParticipants: []*Participant{
					{UserId: "p1"},
					{UserId: "p2"},
					{UserId: "p3"},
				},
			},
			"remove participant successufully": {
				UserId: "p2",
				Participants: []*Participant{
					{UserId: "p1"},
					{UserId: "p2"},
					{UserId: "p3"},
				},
				ExpectedParticipants: []*Participant{
					{UserId: "p1"},
					{UserId: "p3"},
				},
			},
		}{
			t.Run(name, func(t *testing.T){
				g := JankenGame{}
				g.Participants = test.Participants

				g.RemoveParticipant(test.UserId)

				assert := assert.New(t)
				assert.Equal(test.ExpectedParticipants, g.Participants)
			})
		}
	})

	t.Run("UpdateHands", func(t *testing.T){
		for name, test := range map[string]struct {
			UserId               string
			Hands                []string
			Participants         []*Participant
			ExpectedParticipants []*Participant
		}{
			"update hands successfully": {
				UserId: "p2",
				Hands: []string{"paper", "scissors"},
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "scissors"}},
					{UserId: "p2", Hands: []string{"rock", "scissors"}},
					{UserId: "p3", Hands: []string{"rock", "scissors"}},
				},
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "scissors"}},
					{UserId: "p2", Hands: []string{"paper", "scissors"}},
					{UserId: "p3", Hands: []string{"rock", "scissors"}},
				},
			},
			"update by short length hands successfully": {
				UserId: "p2",
				Hands: []string{"paper"},
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "scissors"}},
					{UserId: "p2", Hands: []string{"rock", "scissors"}},
					{UserId: "p3", Hands: []string{"rock", "scissors"}},
				},
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "scissors"}},
					{UserId: "p2", Hands: []string{"paper"}},
					{UserId: "p3", Hands: []string{"rock", "scissors"}},
				},
			},
			"update by long length hands successfully": {
				UserId: "p2",
				Hands: []string{"paper", "scissors", "rock"},
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "scissors"}},
					{UserId: "p2", Hands: []string{"rock", "scissors"}},
					{UserId: "p3", Hands: []string{"rock", "scissors"}},
				},
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "scissors"}},
					{UserId: "p2", Hands: []string{"paper", "scissors", "rock"}},
					{UserId: "p3", Hands: []string{"rock", "scissors"}},
				},
			},
			"add new participant": {
				UserId: "p2",
				Hands: []string{"paper", "scissors"},
				Participants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "scissors"}},
				},
				ExpectedParticipants: []*Participant{
					{UserId: "p1", Hands: []string{"rock", "scissors"}},
					{UserId: "p2", Hands: []string{"paper", "scissors", "", "", "", "", "", "", "", ""}},
				},
			},
		}{
			t.Run(name, func(t *testing.T){
				g := JankenGame{}
				g.Participants = test.Participants

				g.UpdateHands(test.UserId, test.Hands)

				assert := assert.New(t)
				assert.Equal(test.ExpectedParticipants, g.Participants)
			})
		}
	})
}

func TestJanken(t *testing.T) {
	for name, test := range map[string]struct {
		Round                int
		Participants         []*Participant
		ExpectedWinners []*Participant
		ExpectedLosers  []*Participant
		ExpectedDrawers []*Participant
	}{
		"2 drawers": {
			Round: 0,
			Participants: []*Participant{
				{UserId: "p1", Hands: []string{"rock"}},
				{UserId: "p2", Hands: []string{"rock"}},
			},
			ExpectedWinners: nil,
			ExpectedLosers: nil,
			ExpectedDrawers: []*Participant{
				{UserId: "p1", Hands: []string{"rock"}},
				{UserId: "p2", Hands: []string{"rock"}},
			},
		},
		"3 drawers with same hands": {
			Round: 0,
			Participants: []*Participant{
				{UserId: "p1", Hands: []string{"rock"}},
				{UserId: "p2", Hands: []string{"rock"}},
				{UserId: "p3", Hands: []string{"rock"}},
			},
			ExpectedWinners: nil,
			ExpectedLosers: nil,
			ExpectedDrawers: []*Participant{
				{UserId: "p1", Hands: []string{"rock"}},
				{UserId: "p2", Hands: []string{"rock"}},
				{UserId: "p3", Hands: []string{"rock"}},
			},
		},
		"3 drawers with different hands": {
			Round: 0,
			Participants: []*Participant{
				{UserId: "p1", Hands: []string{"rock"}},
				{UserId: "p2", Hands: []string{"scissors"}},
				{UserId: "p3", Hands: []string{"paper"}},
			},
			ExpectedWinners: nil,
			ExpectedLosers: nil,
			ExpectedDrawers: []*Participant{
				{UserId: "p1", Hands: []string{"rock"}},
				{UserId: "p2", Hands: []string{"scissors"}},
				{UserId: "p3", Hands: []string{"paper"}},
			},
		},
		"2 winner and 1 loser": {
			Round: 0,
			Participants: []*Participant{
				{UserId: "p1", Hands: []string{"paper"}},
				{UserId: "p2", Hands: []string{"paper"}},
				{UserId: "p3", Hands: []string{"rock"}},
			},
			ExpectedWinners: []*Participant{
				{UserId: "p1", Hands: []string{"paper"}},
				{UserId: "p2", Hands: []string{"paper"}},
			},
			ExpectedLosers: []*Participant{
				{UserId: "p3", Hands: []string{"rock"}},
			},
			ExpectedDrawers: nil,
		},
	}{
		t.Run(name, func(t *testing.T){
			assert := assert.New(t)

			winners, losers, drawers := janken(test.Participants, test.Round)

			assert.Equal(test.ExpectedWinners, winners)
			assert.Equal(test.ExpectedLosers, losers)
			assert.Equal(test.ExpectedDrawers, drawers)
		})
	}
}
