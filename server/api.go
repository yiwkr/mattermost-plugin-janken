package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	iconFilename = "janken_choki.png"
)

var (
	handsRegisteredMessage = &i18n.Message{
		ID:    "HandsRegisteredMessage",
		Other: "Your hands {{.HandsStr}} are registered with janken game ({{.ID}}).",
	}
	resultPermissionErrorMessage = &i18n.Message{
		ID:    "ResultPermissionErrorMessage",
		Other: "Failed to show the result of the janken game. The creator of this game or the administrator can show the result.",
	}
	resultNotEnoughParticipantsErrorMessage = &i18n.Message{
		ID:    "ResultNotEnoughParticipantsErrorMessage",
		Other: "Failed to show the result of the janken game. Least 2 pariticipants are required.",
	}
	resultTableRankLabel = &i18n.Message{
		ID:    "ResultTableRankLabel",
		Other: "Rank",
	}
	resultTableUsernameLabel = &i18n.Message{
		ID:    "ResultTableUsernameLabel",
		Other: "Username",
	}
	resultTableHandsLabel = &i18n.Message{
		ID:    "ResultTableHandsLabel",
		Other: "Hands",
	}
	resultTableTitle = &i18n.Message{
		ID: "ResultTableTitle",
		Other: `**Janken game ({{.ID}})**
Result
`,
	}
	configPermissionErrorMessage = &i18n.Message{
		ID:    "ConfigPermissionErrorMessage",
		Other: "Failed to open the configration dialog. The creator of this game or the administrator can configure the game.",
	}
	jankenGameDestroyedMessage = &i18n.Message{
		ID:    "gameDestroyedMessage",
		Other: "This janken game was destroyed by @{{.Username}}.",
	}
	failedToGetStoredGameErrorMessage = &i18n.Message{
		ID:    "FailedToGetStoredGameErrorMessage",
		Other: "Failed to get stored game data. Try to create another game.",
	}
)

func (p *Plugin) initAPI() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/"+iconFilename, p.handleIcon).Methods(http.MethodGet)

	apiV1 := r.PathPrefix("/api/v1").Subrouter()
	schedulesRouter := apiV1.PathPrefix("/janken").Subrouter()
	schedulesRouter.HandleFunc("/join", p.handleJoin).Methods(http.MethodPost)
	schedulesRouter.HandleFunc("/join/submit", p.handleJoinSubmit).Methods(http.MethodPost)
	schedulesRouter.HandleFunc("/result", p.handleResult).Methods(http.MethodPost)
	schedulesRouter.HandleFunc("/config", p.handleConfig).Methods(http.MethodPost)
	schedulesRouter.HandleFunc("/config/submit", p.handleConfigSubmit).Methods(http.MethodPost)

	return r
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("New request:", "Host", r.Host, "RequestURI", r.RequestURI, "Method", r.Method)
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) handleJoin(w http.ResponseWriter, r *http.Request) {
	req := model.PostActionIntegrationRequestFromJson(r.Body)

	postID := req.PostId
	userID := req.UserId

	gameID := req.Context["id"].(string)
	game, err := p.store.jankenStore.Get(gameID)
	if err != nil {
		p.API.LogError(err.Error())
		l := p.getLocalizer(p.configuration.DefaultLanguage)
		message := Localize(l, failedToGetStoredGameErrorMessage, nil)
		p.sendEphemeralPost(req.ChannelId, userID, message)
		return
	}

	d := newJoinDialog(p.API, *p.ServerConfig.ServiceSettings.SiteURL, PluginID, p)
	d.Open(req.TriggerId, postID, userID, game)

	response := &model.PostActionIntegrationResponse{}
	writePostActionIntegrationResponse(response, w, r)
}

func (p *Plugin) handleJoinSubmit(w http.ResponseWriter, r *http.Request) {
	req := model.SubmitDialogRequestFromJson(r.Body)

	p.API.LogDebug("handleJoinSubmit", "Submission", fmt.Sprintf("%#v", req.Submission))

	if req.Cancelled {
		return
	}

	userID := req.UserId
	postID := req.CallbackId
	post, _ := p.API.GetPost(postID)

	// get stored data
	gameID := req.State
	game, err := p.store.jankenStore.Get(gameID)
	if err != nil {
		p.API.LogError(err.Error())
		l := p.getLocalizer(p.configuration.DefaultLanguage)
		message := Localize(l, failedToGetStoredGameErrorMessage, nil)
		p.sendEphemeralPost(req.ChannelId, userID, message)
		return
	}

	// submitされたデータの取得
	var cancel bool
	var hands []string
	handsTmp := make([][]string, 0)
	for k, v := range req.Submission {
		if k == "cancel" {
			// cancel
			v := req.Submission[k].(string)
			cancel, _ = strconv.ParseBool(v)
		} else if strings.HasPrefix(k, "hand") {
			// hands
			handsTmp = append(handsTmp, []string{k, v.(string)})
		}
	}

	if cancel {
		// Participantを削除
		game.RemoveParticipant(userID)
	} else {
		// handsTmpをキーでソート
		sort.Slice(handsTmp, func(i, j int) bool {
			return handsTmp[i][0] < handsTmp[j][0]
		})
		hands = make([]string, 0)
		for _, v := range handsTmp {
			hands = append(hands, v[1])
		}
		// Handsを更新
		game.UpdateHands(userID, hands)
	}
	p.API.LogDebug("JoinSubmission", "cancel", cancel, "hands", hands, "userID", userID)

	// save data to store
	p.store.jankenStore.Save(game)

	// update post
	p.attachGameToPost(post, *p.ServerConfig.ServiceSettings.SiteURL, PluginID, game)
	p.API.UpdatePost(post)

	if !cancel {
		// show registered hands
		participant := game.GetParticipant(userID)
		handsEmoji := make([]string, game.MaxRounds)
		for i := 0; i < game.MaxRounds; i++ {
			handsEmoji[i] = handIcons[participant.getHand(i)]
		}
		handsStr := strings.Join(handsEmoji, " ")
		id := game.getShortID()

		l := p.getLocalizer(game.Language)
		message := Localize(l, handsRegisteredMessage, map[string]interface{}{
			"HandsStr": handsStr,
			"ID":       id,
		})

		p.sendEphemeralPost(post.ChannelId, userID, message)
	}
}

func (p *Plugin) handleResult(w http.ResponseWriter, r *http.Request) {
	req := model.PostActionIntegrationRequestFromJson(r.Body)

	userID := req.UserId
	postID := req.PostId
	post, _ := p.API.GetPost(postID)

	// データ取得
	gameID := req.Context["id"].(string)
	game, err := p.store.jankenStore.Get(gameID)
	if err != nil {
		p.API.LogError(err.Error())
		l := p.getLocalizer(p.configuration.DefaultLanguage)
		message := Localize(l, failedToGetStoredGameErrorMessage, nil)
		p.sendEphemeralPost(req.ChannelId, userID, message)
		return
	}

	// localizer
	l := p.getLocalizer(game.Language)

	// 権限チェック
	permission, _ := p.HasPermission(game, userID)
	if !permission {
		message := Localize(l, resultPermissionErrorMessage, nil)
		p.sendEphemeralPost(post.ChannelId, userID, message)
		return
	}

	// 最低人数2人を満たしているかチェック
	if len(game.Participants) < 2 {
		message := Localize(l, resultNotEnoughParticipantsErrorMessage, nil)
		p.sendEphemeralPost(post.ChannelId, req.UserId, message)
		return
	}

	// データ削除
	p.store.jankenStore.Delete(gameID)

	// Attachmentを削除
	model.ParseSlackAttachment(post, nil)

	// 結果取得
	result := game.getResult()
	p.API.LogDebug("Result", "game", fmt.Sprintf("%#v", game), "result", fmt.Sprintf("%#v", result))

	rankLabel := Localize(l, resultTableRankLabel, nil)
	userNameLabel := Localize(l, resultTableUsernameLabel, nil)
	handsLabel := Localize(l, resultTableHandsLabel, nil)

	resultStr := Localize(l, resultTableTitle, map[string]interface{}{
		"ID": game.getShortID(),
	})
	resultStr = fmt.Sprintf("%s\n%s", resultStr, fmt.Sprintf("|%s|%s|%s|", rankLabel, userNameLabel, handsLabel))
	resultStr = fmt.Sprintf("%s\n%s", resultStr, "|:---:|:---|:---|")
	for _, participant := range result {
		username := participant.UserID
		u, err := p.API.GetUser(participant.UserID)
		if err == nil {
			username = u.Username
		}

		hands := make([]string, 0, len(participant.Hands))
		for _, h := range participant.Hands {
			hands = append(hands, handIcons[h])
		}
		handsStr := strings.Join(hands, " ")

		text := fmt.Sprintf("|%d|@%s|%s|", participant.Rank, username, handsStr)
		resultStr = fmt.Sprintf("%s\n%s", resultStr, text)
	}

	// 結果を追加
	appendMessage(post, resultStr)

	response := &model.PostActionIntegrationResponse{}
	response.Update = post
	writePostActionIntegrationResponse(response, w, r)
}

func (p *Plugin) handleConfig(w http.ResponseWriter, r *http.Request) {
	req := model.PostActionIntegrationRequestFromJson(r.Body)

	userID := req.UserId
	postID := req.PostId
	post, _ := p.API.GetPost(postID)

	gameID := req.Context["id"].(string)
	game, err := p.store.jankenStore.Get(gameID)
	if err != nil {
		p.API.LogError(err.Error())
		l := p.getLocalizer(p.configuration.DefaultLanguage)
		message := Localize(l, failedToGetStoredGameErrorMessage, nil)
		p.sendEphemeralPost(req.ChannelId, userID, message)
		return
	}

	// 権限チェック
	permission, _ := p.HasPermission(game, userID)
	if !permission {
		l := p.getLocalizer(game.Language)
		message := Localize(l, configPermissionErrorMessage, nil)
		p.sendEphemeralPost(post.ChannelId, userID, message)
		return
	}

	d := newConfigDialog(p.API, *p.ServerConfig.ServiceSettings.SiteURL, PluginID, p)
	d.Open(req.TriggerId, postID, game)

	response := &model.PostActionIntegrationResponse{}
	writePostActionIntegrationResponse(response, w, r)
}

func (p *Plugin) handleConfigSubmit(w http.ResponseWriter, r *http.Request) {
	req := model.SubmitDialogRequestFromJson(r.Body)

	p.API.LogDebug("handleConfigSubmit", "Submission", fmt.Sprintf("%#v", req.Submission))

	if req.Cancelled {
		return
	}

	userID := req.UserId
	postID := req.CallbackId
	post, _ := p.API.GetPost(postID)

	gameID := req.State
	game, err := p.store.jankenStore.Get(gameID)
	if err != nil {
		p.API.LogError(err.Error())
		l := p.getLocalizer(p.configuration.DefaultLanguage)
		message := Localize(l, failedToGetStoredGameErrorMessage, nil)
		p.sendEphemeralPost(req.ChannelId, userID, message)
		return
	}

	destroy, _ := strconv.ParseBool(req.Submission["destroy"].(string))
	maxRounds, _ := strconv.Atoi(req.Submission["max_rounds"].(string))
	p.API.LogDebug("submission", "destroy", destroy, "maxRounds", maxRounds)

	if destroy {
		p.store.jankenStore.Delete(gameID)
		// Attachmentを削除
		model.ParseSlackAttachment(post, nil)

		// メッセージを追加
		l := p.getLocalizer(game.Language)
		user, _ := p.API.GetUser(req.UserId)
		message := Localize(l, jankenGameDestroyedMessage, map[string]interface{}{
			"Username": user.Username,
		})
		appendMessage(post, message)

		// 更新
		p.API.UpdatePost(post)
		return
	}

	game.MaxRounds = maxRounds

	p.store.jankenStore.Save(game)

	p.attachGameToPost(post, *p.ServerConfig.ServiceSettings.SiteURL, PluginID, game)
	p.API.UpdatePost(post)
}

func (p *Plugin) handleIcon(w http.ResponseWriter, r *http.Request) {
	bundlePath, err := p.API.GetBundlePath()
	if err != nil {
		p.API.LogWarn("failed to get bundle path", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=604800")
	http.ServeFile(w, r, filepath.Join(bundlePath, "assets", iconFilename))
}

func writePostActionIntegrationResponse(response *model.PostActionIntegrationResponse, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response.ToJson())
}
