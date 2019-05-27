package janken

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	COMMAND_RESPONSE_USERNAME = "mattermost-plugin-janken"
)

var (
	jankenGameTitle = &i18n.Message{
		ID:    "JankenGameTitle",
		Other: "Janken Game ({{.ID}}) created by @{{.Username}}",
	}
	jankenGameDescription = &i18n.Message{
		ID: "JankenGameDescription",
		Other: `Please join this janken game.
Participants ({{.ParticipantsNum}}): {{.ParticipantsStr}}`,
	}
	jankenGameJoinButtonLabel = &i18n.Message{
		ID:    "JankenGameJoinButtonLabel",
		Other: "Join",
	}
	jankenGameConfigButtonLabel = &i18n.Message{
		ID:    "JankenGameConfigButtonLabel",
		Other: "Config",
	}
	jankenGameResultButtonLabel = &i18n.Message{
		ID:    "JankenGameResultButtonLabel",
		Other: "Result",
	}
)

type ParsedArgs struct {
	Language string
}

func NewParsedArgs() *ParsedArgs {
	return &ParsedArgs{
		Language: defaultLanguage.String(),
	}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.API.LogDebug("ExecuteCommand", "Context", fmt.Sprintf("%#v", c), "args", fmt.Sprintf("%#v", args))

	parsedArgs, err := p.parseArgs(args.Command)
	if err != nil {
		usage := p.getCommandUsage()
		errmsg := fmt.Sprintf("Failed to parse arguments.: %s", err.Error())
		message := fmt.Sprintf("%s\n\n%s", usage, errmsg)
		response := NewCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, message, nil)
		return response, nil
	}

	game := NewJankenGame(&JankenGameImpl1{})
	game.Creator = args.UserId
	game.Language = parsedArgs.Language
	err = p.store.jankenStore.Save(game)
	if err != nil {
		errmsg := fmt.Sprintf("Failed to store game data.: %s", err.Error())
		response := NewCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, errmsg, nil)
		return response, nil
	}

	if !p.isValidLanguage(game.Language) {
		defaultLanguageStr := defaultLanguage.String()
		message := fmt.Sprintf(`Language "%s" is not available. "%s" is used instead.`, game.Language, defaultLanguageStr)
		p.sendEphemeralPost(args.ChannelId, args.UserId, message)
		game.Language = defaultLanguageStr
	}

	attachments := p.getJankenGameAttachments(*p.ServerConfig.ServiceSettings.SiteURL, PluginId, game)
	response := NewCommandResponse(model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, "", attachments)
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

func (p *Plugin) parseArgs(command string) (*ParsedArgs, error) {
	// split command string like shell arguments
	args, err := shellquote.Split(command)
	if err != nil {
		return nil, err
	}
	args = args[1:]

	parsedArgs := NewParsedArgs()
	positionalArgs := make([]string, 0)
	unknownOptions := make([]string, 0)

	for i := 0; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "-"):
			option := strings.TrimPrefix(args[i], "-")
			switch {
			case option == "l":
				if i >= len(args)-1 {
					return nil, errors.New(fmt.Sprintf(`"-l" option requires a value, "en" or "ja".`))
				}
				parsedArgs.Language = args[i+1]
				i++
			default:
				unknownOptions = append(unknownOptions, args[i])
			}
		default:
			positionalArgs = append(positionalArgs, args[i])
		}
	}

	if len(unknownOptions) > 0 {
		return nil, errors.New(fmt.Sprintf("Invalid arguments: %s", unknownOptions))
	}
	if len(positionalArgs) > 0 {
		return nil, errors.New(fmt.Sprintf("Invalid arguments: %s", positionalArgs))
	}

	return parsedArgs, nil
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

	l := p.GetLocalizer(game.Language)
	// get localized messages
	title := Localize(l, jankenGameTitle, map[string]interface{}{
		"ID":       game.GetShortId(),
		"Username": username,
	})
	description := Localize(l, jankenGameDescription, map[string]interface{}{
		"ParticipantsNum": len(participants),
		"ParticipantsStr": participants_str,
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
					URL:     fmt.Sprintf("%s/plugins/%s/api/v1/janken/result", siteURL, pluginId),
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

func (p *Plugin) getCommandUsage() string {
	template := `
	Usage: /%s [-l en|ja]

	Optional arguments
	  -l en|ja   Language (default: en)
	`
	return fmt.Sprintf(template, p.configuration.Trigger)
}

func NewCommandResponse(responseType, text string, attachments []*model.SlackAttachment) *model.CommandResponse {
	response := &model.CommandResponse{
		ResponseType: responseType,
		Text:         text,
		Username:     COMMAND_RESPONSE_USERNAME,
		Attachments:  attachments,
	}
	return response
}
