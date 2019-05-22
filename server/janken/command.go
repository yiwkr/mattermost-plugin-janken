package janken

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.API.LogDebug("ExecuteCommand", "Context", fmt.Sprintf("%#v", c), "args", fmt.Sprintf("%#v", args))

	game := NewJankenGame()
	// 作成者を記録
	game.Creator = args.UserId
	p.store.jankenStore.Save(game)

	response := &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL,
		Text: "",
		Username: "mattermost-plugin-janken",
		ChannelId: args.ChannelId,
		Attachments: p.getJankenGameAttachments(*p.ServerConfig.ServiceSettings.SiteURL, PluginId, game),
	}

	return response, nil
}

func (p *Plugin) getJankenGameAttachments(siteURL, pluginId string, game *JankenGame) []*model.SlackAttachment {
	// 現在の参加者を取得
	participants := make([]string, len(game.Participants))
	for i, pp := range game.Participants {
		user, err := p.API.GetUser(pp.UserId)
		if err != nil {
			p.API.LogError(fmt.Sprintf("User %s is not found.", pp.UserId))
			continue
		}
		participants[i] = user.Username
	}
	// カンマ区切りの文字列に変換
	participants_str := strings.Join(participants, ", ")

	context := map[string]interface{}{
		"id": game.Id,
	}

	var username string
	user, err := p.API.GetUser(game.Creator)
	if err != nil {
		username = "anonymous"
	}
	username = user.Username

	title := fmt.Sprintf("Janken Game (%s) created by @%s", game.GetShortId(), username)
	description := fmt.Sprintf("Please join this janken game.\nParticipants (%d): %s", len(participants), participants_str)
	joinButtonLabel := "Join"
	configButtonLabel := "Config"
	resultButtonLabel := "Result"

	attachments := []*model.SlackAttachment{{
		Title:      title,
		Text:       description,
		Actions:    []*model.PostAction{
			{
				Name: joinButtonLabel,
				Type: model.POST_ACTION_TYPE_BUTTON,
				Integration: &model.PostActionIntegration{
					URL:     fmt.Sprintf("%s/plugins/%s/api/v1/janken/join", siteURL, pluginId),
					Context: context,
				},
			},
			{
				Name: configButtonLabel,
				Type: model.POST_ACTION_TYPE_BUTTON,
				Integration: &model.PostActionIntegration{
					URL:     fmt.Sprintf("%s/plugins/%s/api/v1/janken/config", siteURL, pluginId),
					Context: context,
				},
			},
			{
				Name: resultButtonLabel,
				Type: model.POST_ACTION_TYPE_BUTTON,
				Integration: &model.PostActionIntegration{
					URL: fmt.Sprintf("%s/plugins/%s/api/v1/janken/result", siteURL, pluginId),
					Context: context,
				},
			},
		},
	}}
	return attachments
}

func (p *Plugin) attachJankenGameToPost(post *model.Post, siteURL, pluginId string, game *JankenGame) *model.Post {
	attachments := p.getJankenGameAttachments(siteURL, pluginId, game)

	model.ParseSlackAttachment(post, attachments)
	return post
}
