package janken

import (
)

type JankenGameImpl1 struct {
	JankenGameBase
}

func NewJankenGameImpl1() JankenGameInterface {
	return &JankenGameImpl1{}
}

// GetResultは最大maxRoundsのジャンケンの結果を返す．
// 結果は[]interface{}（各要素はinterface{}{int, Participant}]．最初の要素は順位、次の要素はParticipant）
func (g *JankenGameImpl1) GetResult(game *JankenGame) []*Participant {
	start_round := 0  // Handsの利用開始番号
	start_rank := 1  // 順位の開始番号
	result := g.nextRound(game.Participants, game.MaxRounds, start_round, start_rank, nil)
	return result
}

/*
n回戦のジャンケン結果を返すための再帰関数
1位から順に*Participantをresultに格納していく
Args:
    participants: 今評価中のジャンケンの参加者
    raound: 現在のラウンド
    rank: 今つけようとしている順位
    result: 結果
Returns:
    []*Participant: result
*/
func (g *JankenGameImpl1) nextRound(participants []*Participant, maxRounds, round, rank int, result []*Participant) []*Participant {
	if result == nil {
		result = make([]*Participant, 0, len(participants))
	}

	/*
	終了条件1: 勝者or敗者1人になった場合
	participantsに残っている1人をresultに格納する
	*/
	if len(participants) == 1 {
		participants[0].Rank = rank
		participants[0].ClearHandsAfter(round)
		result = append(result, participants[0])
		return result
	}

	/*
	終了条件2: 最大対戦回数に達した場合
	今残っているparticipantsをすべて同じ順位でresultに格納する
	*/
	if round >= maxRounds {
		for _, p := range participants {
			p.Rank = rank
			result = append(result, p)
		}
		return result
	}

	// ジャンケンを1回実行
	winner, loser, drawer := janken(participants, round)

	if drawer != nil {
		// あいこの処理
		result = g.nextRound(drawer, maxRounds, round+1, rank, result)
	} else {
		// 勝者の処理
		result = g.nextRound(winner, maxRounds, round+1, rank, result)
		// 敗者の処理
		result = g.nextRound(loser, maxRounds, round+1, rank+len(winner), result)
	}
	return result
}

