package main

type gameImpl1 struct {
	gameBase
}

func newGameImpl1() gameInterface {
	return &gameImpl1{}
}

// getResult は最大maxRoundsのジャンケンの結果を返す．
// 結果は[]interface{}（各要素はinterface{}{int, participant}]．最初の要素は順位、次の要素はparticipant）
func (g *gameImpl1) getResult(game *game) []*participant {
	startRound := 0 // Handsの利用開始番号
	startRank := 1  // 順位の開始番号
	result := g.nextRound(game.Participants, game.MaxRounds, startRound, startRank, nil)
	return result
}

/*
n回戦のジャンケン結果を返すための再帰関数
1位から順に*participantをresultに格納していく
Args:
    participants: 今評価中のジャンケンの参加者
    raound: 現在のラウンド
    rank: 今つけようとしている順位
    result: 結果
Returns:
    []*participant: result
*/
func (g *gameImpl1) nextRound(participants []*participant, maxRounds, round, rank int, result []*participant) []*participant {
	if result == nil {
		result = make([]*participant, 0, len(participants))
	}

	/*
		終了条件1: 勝者or敗者1人になった場合
		participantsに残っている1人をresultに格納する
	*/
	if len(participants) == 1 {
		participants[0].Rank = rank
		participants[0].clearHandsAfter(round)
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
