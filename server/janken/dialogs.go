package janken

import (
	"fmt"
	"strconv"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var handMessages = map[string]*i18n.Message{
	"rock": &i18n.Message{
		ID:    "JoinDialogHandRock",
		Other: "Rock",
	},
	"scissors": &i18n.Message{
		ID:    "JoinDialogHandScissors",
		Other: "Scissors",
	},
	"paper": &i18n.Message{
		ID:    "JoinDialogHandPaper",
		Other: "Paper",
	},
}

var (
	joinDialogTitle = &i18n.Message{
		ID:    "JoinDialogTitle",
		Other: "Join the janken game",
	}
	joinDialogSubmitLabel = &i18n.Message{
		ID:    "JoinDialogSubmitLabel",
		Other: "Save",
	}
	joinDialogCancelLabel = &i18n.Message{
		ID:    "JoinDialogCancelLabel",
		Other: "Cancel",
	}
	joinDialogHandElementLabel = &i18n.Message{
		ID:    "JoinDialogHandElementLabel",
		Other: "Hand {{.Index}}",
	}
	joinDialogHandElementHelp = &i18n.Message{
		ID:    "JoinDialogHandElementHelp",
		Other: "Choose hand {{.Index}}",
	}
	configDialogTitle = &i18n.Message{
		ID:    "ConfigDialogTitle",
		Other: "Config",
	}
	configDialogSubmitLabel = &i18n.Message{
		ID:    "ConfigDialogSubmitLabel",
		Other: "Save",
	}
	configDialogMaxRoundsLabel = &i18n.Message{
		ID:    "ConfigDialogMaxRoundsLabel",
		Other: "Max rounds",
	}
	configDialogDestroyLabel = &i18n.Message{
		ID:    "ConfigDialogDestroyLabel",
		Other: "Destroy this game",
	}
)

// 参加取消の選択肢
var CancelOptions []*model.PostActionOptions = []*model.PostActionOptions{
	{
		Text:  "-",
		Value: "false",
	},
	{
		Text:  "cancel",
		Value: "true",
	},
}

// ゲーム削除の選択肢
var DestroyOptions []*model.PostActionOptions = []*model.PostActionOptions{
	{
		Text:  "-",
		Value: "false",
	},
	{
		Text:  "destroy",
		Value: "true",
	},
}

type Dialog struct {
	API      plugin.API
	siteURL  string
	pluginId string
	plugin   *Plugin
}

// JoinDialogは"参加"ボタンが押されたときに開くダイアログ
type JoinDialog struct{ Dialog }

func NewJoinDialog(api plugin.API, siteURL, pluginId string, plugin *Plugin) *JoinDialog {
	d := &JoinDialog{}
	d.API = api
	d.siteURL = siteURL
	d.pluginId = pluginId
	d.plugin = plugin
	return d
}

func (d *JoinDialog) Open(triggerId, postId, userId string, game *JankenGame) {
	d.API.LogDebug("openJoinDialog is called")

	l := d.plugin.GetLocalizer(game.Language)
	dialogTitle := Localize(l, joinDialogTitle, nil)
	submitLabel := Localize(l, joinDialogSubmitLabel, nil)
	cancelLabel := Localize(l, joinDialogCancelLabel, nil)

	// ジャンケンで出せる手
	var HandsOptions []*model.PostActionOptions = []*model.PostActionOptions{
		{
			Text:  Localize(l, handMessages["rock"], nil),
			Value: "rock",
		},
		{
			Text:  Localize(l, handMessages["scissors"], nil),
			Value: "scissors",
		},
		{
			Text:  Localize(l, handMessages["paper"], nil),
			Value: "paper",
		},
	}

	participant := game.GetParticipant(userId)
	if participant == nil {
		participant = NewParticipant(userId)
	}

	// 手の入力フォームを追加
	elements := []model.DialogElement{}
	for i := 0; i < game.MaxRounds; i++ {

		i1 := i + 1 // 1-base index
		displayName := Localize(l, joinDialogHandElementLabel, map[string]interface{}{
			"Index": i1,
		})
		name := fmt.Sprintf("hand%d", i1)
		helpText := Localize(l, joinDialogHandElementHelp, map[string]interface{}{
			"Index": i1,
		})

		hand := participant.GetHand(i)
		localizedHand := Localize(l, handMessages[hand], nil)

		elements = append(elements, model.DialogElement{
			DisplayName: displayName,
			Name:        name,
			Type:        "select",
			Placeholder: localizedHand,
			Default:     hand,
			Optional:    true,
			Options:     HandsOptions,
			HelpText:    helpText,
		})
	}

	elements = append(elements, model.DialogElement{
		DisplayName: cancelLabel,
		Name:        "cancel",
		Type:        "select",
		Placeholder: "-",
		Default:     "false",
		Optional:    true,
		Options:     CancelOptions,
	})

	dialog := model.Dialog{
		CallbackId:     postId,
		Title:          dialogTitle,
		SubmitLabel:    submitLabel,
		NotifyOnCancel: false,
		State:          game.Id,
		Elements:       elements,
	}

	request := model.OpenDialogRequest{
		TriggerId: triggerId,
		URL:       fmt.Sprintf("%s/plugins/%s/api/v1/janken/join/submit", d.siteURL, d.pluginId),
		Dialog:    dialog,
	}

	_ = d.API.OpenInteractiveDialog(request)
}

// ConfigDialogは"設定"ボタンが押されたときに開くダイアログ
type ConfigDialog struct{ Dialog }

func NewConfigDialog(api plugin.API, siteURL, pluginId string, plugin *Plugin) *ConfigDialog {
	d := &ConfigDialog{}
	d.API = api
	d.siteURL = siteURL
	d.pluginId = pluginId
	d.plugin = plugin
	return d
}

func (d *ConfigDialog) Open(triggerId, postId string, game *JankenGame) {
	d.API.LogDebug("openConfigDialog is called")

	// options for maxRounds
	maxRoundsOptions := []*model.PostActionOptions{}
	for i := 1; i <= MAX_HANDS; i++ {
		maxRoundsOptions = append(maxRoundsOptions, &model.PostActionOptions{
			Text: strconv.Itoa(i), Value: strconv.Itoa(i),
		})
	}

	l := d.plugin.GetLocalizer(game.Language)
	dialogTitle := Localize(l, configDialogTitle, nil)
	submitLabel := Localize(l, configDialogSubmitLabel, nil)
	maxRoundsLabel := Localize(l, configDialogMaxRoundsLabel, nil)
	destroyLabel := Localize(l, configDialogDestroyLabel, nil)

	elements := []model.DialogElement{
		{
			DisplayName: maxRoundsLabel,
			Name:        "max_rounds",
			Type:        "select",
			Placeholder: strconv.Itoa(game.MaxRounds),
			Default:     strconv.Itoa(game.MaxRounds),
			Options:     maxRoundsOptions,
		},
		{
			DisplayName: destroyLabel,
			Name:        "destroy",
			Type:        "select",
			Placeholder: "-",
			Default:     "false",
			Optional:    true,
			Options:     DestroyOptions,
		},
	}

	dialog := model.Dialog{
		CallbackId:     postId,
		Title:          dialogTitle,
		SubmitLabel:    submitLabel,
		NotifyOnCancel: false,
		State:          game.Id,
		Elements:       elements,
	}

	request := model.OpenDialogRequest{
		TriggerId: triggerId,
		URL:       fmt.Sprintf("%s/plugins/%s/api/v1/janken/config/submit", d.siteURL, d.pluginId),
		Dialog:    dialog,
	}

	_ = d.API.OpenInteractiveDialog(request)
}
