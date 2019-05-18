package janken

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

func (p *Plugin) InitAPI() *mux.Router {
	r := mux.NewRouter()

	apiV1 := r.PathPrefix("/api/v1").Subrouter()

	schedulesRouter := apiV1.PathPrefix("/janken").Subrouter()
	schedulesRouter.HandleFunc("/join", p.handleJoin).Methods("POST")
	schedulesRouter.HandleFunc("/join/submit", p.handleJoinSubmit).Methods("POST")
	schedulesRouter.HandleFunc("/result", p.handleResult).Methods("POST")
	schedulesRouter.HandleFunc("/config", p.handleConfig).Methods("POST")
	schedulesRouter.HandleFunc("/config/submit", p.handleConfigSubmit).Methods("POST")

	return r
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("New request:", "Host", r.Host, "RequestURI", r.RequestURI, "Method", r.Method)
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) handleJoin(w http.ResponseWriter, r *http.Request) {
	req := model.PostActionIntegrationRequestFromJson(r.Body)

	postId := req.PostId
	userId := req.UserId

	gameId := req.Context["id"].(string)
	game := p.store.jankenStore.Get(gameId)

	d := NewJoinDialog(p.API, *p.ServerConfig.ServiceSettings.SiteURL, PluginId)
	d.Open(req.TriggerId, postId, userId, game)

	response := &model.PostActionIntegrationResponse{}
	p.writePostActionIntegrationResponse(response, w, r)
}

func (p *Plugin) handleJoinSubmit(w http.ResponseWriter, r *http.Request) {
	req := model.SubmitDialogRequestFromJson(r.Body)

	p.API.LogDebug("handleJoinSubmit", "Submission", fmt.Sprintf("%#v", req.Submission))

	if req.Cancelled {
		return
	}

	userId := req.UserId
	postId := req.CallbackId
	post, _ := p.API.GetPost(postId)

	// get stored data
	gameId := req.State
	game := p.store.jankenStore.Get(gameId)

	// submitされたデータの取得
	var cancel bool
	var hands []string
	hands_tmp := make([][]string, 0)
	for k, v := range req.Submission {
		if k == "cancel" {
			// cancel
			v := req.Submission[k].(string)
			cancel, _ = strconv.ParseBool(v)
		} else if strings.HasPrefix(k, "hand") {
			// hands
			hands_tmp = append(hands_tmp, []string{k, v.(string)})
		}
	}

	if cancel {
		// Participantを削除
		game.RemoveParticipant(userId)
	} else {
		// hands_tmpをキーでソート
		sort.Slice(hands_tmp, func(i, j int) bool {
			return hands_tmp[i][0] < hands_tmp[j][0]
		})
		hands := make([]string, 0)
		for _, v := range hands_tmp {
			hands = append(hands, v[1])
		}
		// Handsを更新
		game.UpdateHands(userId, hands)
	}
	p.API.LogDebug("JoinSubmission", "cancel", cancel, "hands", hands, "userId", userId)

	// save data to store
	p.store.jankenStore.Save(game)

	// update post
	p.attachJankenGameToPost(post, *p.ServerConfig.ServiceSettings.SiteURL, PluginId, game)
	p.API.UpdatePost(post)

	if !cancel {
		// show registered hands
		participant := game.GetParticipant(userId)
		hands_emoji := make([]string, game.MaxRounds)
		for i := 0; i<game.MaxRounds; i++ {
			hands_emoji[i] = HandIcons[participant.GetHand(i)]
		}
		hands_str := strings.Join(hands_emoji, " ")
		id := game.GetShortId()
		message := fmt.Sprintf(`あなたの手 %s をジャンケンゲーム(%s)に登録しました`, hands_str, id)
		p.SendEphemeralPost(post.ChannelId, userId, message)
	}
}

func (p *Plugin) handleResult(w http.ResponseWriter, r *http.Request) {
	req := model.PostActionIntegrationRequestFromJson(r.Body)

	userId := req.UserId
	postId := req.PostId
	post, _ := p.API.GetPost(postId)

	// データ取得
	gameId := req.Context["id"].(string)
	game := p.store.jankenStore.Get(gameId)

	// 権限チェック
	permission, _ := p.HasPermission(game, userId)
	if !permission {
		message := fmt.Sprint("ジャンケンゲームの結果を表示できませんでした。結果を表示できるのは作成者か管理者のみです。")
		p.SendEphemeralPost(post.ChannelId, userId, message)
		return
	}

	// 最低人数2人を満たしているかチェック
	if len(game.Participants) < 2 {
		message := "ジャンケンゲームの結果を表示できませんでした。結果を表示するには2人以上の参加者が必要です"
		p.SendEphemeralPost(post.ChannelId, req.UserId, message)
		return
	}

	// データ削除
	p.store.jankenStore.Delete(gameId)

	// Attachmentを削除
	model.ParseSlackAttachment(post, nil)

	// 結果取得
	result := game.GetResult()
	p.API.LogDebug("Result", "game", fmt.Sprintf("%#v", game), "result", fmt.Sprintf("%#v", result))

	result_str := fmt.Sprintf("**ジャンケンゲーム (%s)**\n結果\n", game.GetShortId())
	result_str = fmt.Sprintf("%s\n%s", result_str, "|順位|名前|手|")
	result_str = fmt.Sprintf("%s\n%s", result_str, "|:---|:---|:---|")
	for _, participant := range result {
		username := participant.UserId
		u, err := p.API.GetUser(participant.UserId)
		if err == nil {
			username = u.Username
		}

		hands := make([]string, 0, len(participant.Hands))
		for _, h := range participant.Hands {
			hands = append(hands, HandIcons[h])
		}
		hands_str := strings.Join(hands, " ")

		text := fmt.Sprintf("|%d位|@%s|%s|", participant.Rank, username, hands_str)
		result_str = fmt.Sprintf("%s\n%s", result_str, text)
	}

	// 結果を追加
	p.AppendMessage(post, result_str)

	response := &model.PostActionIntegrationResponse{}
	response.Update = post
	p.writePostActionIntegrationResponse(response, w, r)
}

func (p *Plugin) handleConfig(w http.ResponseWriter, r *http.Request) {
	req := model.PostActionIntegrationRequestFromJson(r.Body)

	userId := req.UserId
	postId := req.PostId
	post, _ := p.API.GetPost(postId)

	gameId := req.Context["id"].(string)
	game := p.store.jankenStore.Get(gameId)

	// 権限チェック
	permission, _ := p.HasPermission(game, userId)
	if !permission {
		message := fmt.Sprint("ジャンケンゲームの設定を開けませんでした。設定は作成者か管理者のみが実行できます。")
		p.SendEphemeralPost(post.ChannelId, userId, message)
		return
	}

	d := NewConfigDialog(p.API, *p.ServerConfig.ServiceSettings.SiteURL, PluginId)
	d.Open(req.TriggerId, postId, game)

	response := &model.PostActionIntegrationResponse{}
	p.writePostActionIntegrationResponse(response, w, r)
}

func (p *Plugin) handleConfigSubmit(w http.ResponseWriter, r *http.Request) {
	req := model.SubmitDialogRequestFromJson(r.Body)

	p.API.LogDebug("handleConfigSubmit", "Submission", fmt.Sprintf("%#v", req.Submission))

	if req.Cancelled {
		return
	}

	postId := req.CallbackId
	post, _ := p.API.GetPost(postId)

	gameId := req.State
	game := p.store.jankenStore.Get(gameId)

	destroy, _ := strconv.ParseBool(req.Submission["destroy"].(string))
	maxRounds, _ := strconv.Atoi(req.Submission["max_rounds"].(string))
	p.API.LogDebug("submission", "destroy", destroy, "maxRounds", maxRounds)

	if destroy {
		p.store.jankenStore.Delete(gameId)
		// Attachmentを削除
		model.ParseSlackAttachment(post, nil)

		// メッセージを追加
		user, _ := p.API.GetUser(req.UserId)
		message := fmt.Sprintf("このジャンケンゲームは@%sによって削除されました", user.Username)
		p.AppendMessage(post, message)

		// 更新
		p.API.UpdatePost(post)
		return
	}

	game.MaxRounds = maxRounds

	p.store.jankenStore.Save(game)

	p.attachJankenGameToPost(post, *p.ServerConfig.ServiceSettings.SiteURL, PluginId, game)
	p.API.UpdatePost(post)
}
