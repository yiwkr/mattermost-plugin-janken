package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	commandResponseUsername = "janken"
)

var (
	jankenGameTitle = &i18n.Message{
		ID:    "gameTitle",
		Other: "Janken game ({{.ID}}) created by @{{.Username}}",
	}
	jankenGameDescription = &i18n.Message{
		ID: "gameDescription",
		Other: `Please join this janken game.
participants ({{.participantsNum}}): {{.participantsStr}}`,
	}
	jankenGameJoinButtonLabel = &i18n.Message{
		ID:    "gameJoinButtonLabel",
		Other: "Join",
	}
	jankenGameConfigButtonLabel = &i18n.Message{
		ID:    "gameConfigButtonLabel",
		Other: "Config",
	}
	jankenGameResultButtonLabel = &i18n.Message{
		ID:    "gameResultButtonLabel",
		Other: "Result",
	}
)

type parsedArgs struct {
	Language *string
}

// ExecuteCommand executes a command that has been previously registered via the RegisterCommand API.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.API.LogDebug("executeCommand", "Context", fmt.Sprintf("%#v", c), "args", fmt.Sprintf("%#v", args))

	siteURL := *p.ServerConfig.ServiceSettings.SiteURL

	parsedArgs, err := p.parseArgs(args.Command)
	if err != nil {
		message := p.getCommandUsage()
		if err.Error() != "" {
			errmsg := fmt.Sprintf("Failed to parse arguments.: %s", err.Error())
			message = fmt.Sprintf("%s\n\n%s", message, errmsg)
		}
		response := newCommandResponse(siteURL, model.COMMAND_RESPONSE_TYPE_EPHEMERAL, message, nil)
		return response, nil
	}

	game := newGame(&gameImpl1{})
	game.Creator = args.UserId
	game.Language = *parsedArgs.Language
	err = p.store.jankenStore.Save(game)
	if err != nil {
		errmsg := fmt.Sprintf("Failed to store game data.: %s", err.Error())
		response := newCommandResponse(siteURL, model.COMMAND_RESPONSE_TYPE_EPHEMERAL, errmsg, nil)
		return response, nil
	}

	if !p.isValidLanguage(game.Language) {
		defaultLanguageStr := p.configuration.DefaultLanguage
		if game.Language != "" {
			message := fmt.Sprintf(`Language "%s" is not available. "%s" is used instead.`, game.Language, defaultLanguageStr)
			p.sendEphemeralPost(args.ChannelId, args.UserId, message)
		}
		game.Language = defaultLanguageStr
	}

	attachments := p.getGameAttachments(siteURL, PluginID, game)
	response := newCommandResponse(siteURL, model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, "", attachments)
	return response, nil
}

func (p *Plugin) isValidLanguage(language string) bool {
	for _, t := range p.bundle.LanguageTags() {
		if language == t.String() {
			return true
		}
	}
	return false
}

func (p *Plugin) parseArgs(command string) (*parsedArgs, error) {
	parsedArgs := &parsedArgs{Language: &p.configuration.DefaultLanguage}

	fs := flag.NewFlagSet("janken", flag.ContinueOnError)
	parsedArgs.Language = fs.String("l", "", `Language option. Available values are "en" or "ja".`)
	flag.ErrHelp = errors.New("")

	// split command string like shell arguments
	args, err := shellquote.Split(command)
	if err != nil {
		return nil, err
	}
	if err := fs.Parse(args[1:]); err != nil {
		return nil, err
	}

	positionalArgs := fs.Args()
	if len(positionalArgs) > 0 {
		return nil, fmt.Errorf("Invalid arguments: %s", positionalArgs)
	}

	return parsedArgs, nil
}

func (p *Plugin) getGameAttachments(siteURL, pluginID string, game *game) []*model.SlackAttachment {
	// 現在の参加者を取得
	participants := make([]string, len(game.Participants))
	for i, pp := range game.Participants {
		user, err := p.API.GetUser(pp.UserID)
		if err != nil {
			p.API.LogError(fmt.Sprintf("User %s is not found.", pp.UserID))
			continue
		}
		participants[i] = user.Username
	}
	// カンマ区切りの文字列に変換
	participantsStr := strings.Join(participants, ", ")

	context := map[string]interface{}{
		"id": game.ID,
	}

	var username string
	user, err := p.API.GetUser(game.Creator)
	if err != nil {
		username = "anonymous"
	}
	username = user.Username

	l := p.getLocalizer(game.Language)
	// get localized messages
	title := Localize(l, jankenGameTitle, map[string]interface{}{
		"ID":       game.getShortID(),
		"Username": username,
	})
	description := Localize(l, jankenGameDescription, map[string]interface{}{
		"participantsNum": len(participants),
		"participantsStr": participantsStr,
	})
	joinButtonLabel := Localize(l, jankenGameJoinButtonLabel, nil)
	configButtonLabel := Localize(l, jankenGameConfigButtonLabel, nil)
	resultButtonLabel := Localize(l, jankenGameResultButtonLabel, nil)

	attachments := []*model.SlackAttachment{{
		Title: title,
		Text:  description,
		Actions: []*model.PostAction{
			{
				Name: joinButtonLabel,
				Type: model.POST_ACTION_TYPE_BUTTON,
				Integration: &model.PostActionIntegration{
					URL:     fmt.Sprintf("%s/plugins/%s/api/v1/janken/join", siteURL, pluginID),
					Context: context,
				},
			},
			{
				Name: configButtonLabel,
				Type: model.POST_ACTION_TYPE_BUTTON,
				Integration: &model.PostActionIntegration{
					URL:     fmt.Sprintf("%s/plugins/%s/api/v1/janken/config", siteURL, pluginID),
					Context: context,
				},
			},
			{
				Name: resultButtonLabel,
				Type: model.POST_ACTION_TYPE_BUTTON,
				Integration: &model.PostActionIntegration{
					URL:     fmt.Sprintf("%s/plugins/%s/api/v1/janken/result", siteURL, pluginID),
					Context: context,
				},
			},
		},
	}}
	return attachments
}

func (p *Plugin) attachGameToPost(post *model.Post, siteURL, pluginID string, game *game) *model.Post {
	attachments := p.getGameAttachments(siteURL, pluginID, game)

	model.ParseSlackAttachment(post, attachments)
	return post
}

func (p *Plugin) getCommandUsage() string {
	template := `
	Usage: /%s [-l en|ja]

	Optional arguments
	  -l en|ja   Language
	`
	return fmt.Sprintf(template, p.configuration.Trigger)
}

func newCommandResponse(siteURL, responseType, text string, attachments []*model.SlackAttachment) *model.CommandResponse {
	response := &model.CommandResponse{
		ResponseType: responseType,
		Text:         text,
		Username:     commandResponseUsername,
		Attachments:  attachments,
		IconURL:      fmt.Sprintf("%s/plugins/%s/%s", siteURL, PluginID, iconFilename),
	}
	return response
}
