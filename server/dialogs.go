package main

import (
	"fmt"
	"strconv"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var handMessages = map[string]*i18n.Message{
	"rock": {
		ID:    "joinDialogHandRock",
		Other: "Rock",
	},
	"scissors": {
		ID:    "joinDialogHandScissors",
		Other: "Scissors",
	},
	"paper": {
		ID:    "joinDialogHandPaper",
		Other: "Paper",
	},
}

var (
	joinDialogTitle = &i18n.Message{
		ID:    "joinDialogTitle",
		Other: "Join the janken game",
	}
	joinDialogSubmitLabel = &i18n.Message{
		ID:    "joinDialogSubmitLabel",
		Other: "Save",
	}
	joinDialogCancelLabel = &i18n.Message{
		ID:    "joinDialogCancelLabel",
		Other: "Cancel",
	}
	joinDialogHandElementLabel = &i18n.Message{
		ID:    "joinDialogHandElementLabel",
		Other: "Hand {{.Index}}",
	}
	joinDialogHandElementHelp = &i18n.Message{
		ID:    "joinDialogHandElementHelp",
		Other: "Choose hand {{.Index}}",
	}
	configDialogTitle = &i18n.Message{
		ID:    "configDialogTitle",
		Other: "Config",
	}
	configDialogSubmitLabel = &i18n.Message{
		ID:    "configDialogSubmitLabel",
		Other: "Save",
	}
	configDialogMaxRoundsLabel = &i18n.Message{
		ID:    "configDialogMaxRoundsLabel",
		Other: "Max rounds",
	}
	configDialogDestroyLabel = &i18n.Message{
		ID:    "configDialogDestroyLabel",
		Other: "Destroy this game",
	}
)

// 参加取消の選択肢
var cancelOptions []*model.PostActionOptions = []*model.PostActionOptions{
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
var destroyOptions []*model.PostActionOptions = []*model.PostActionOptions{
	{
		Text:  "-",
		Value: "false",
	},
	{
		Text:  "destroy",
		Value: "true",
	},
}

type dialog struct {
	API      plugin.API
	siteURL  string
	pluginID string
	plugin   *Plugin
}

// joinDialogは"参加"ボタンが押されたときに開くダイアログ
type joinDialog struct{ dialog }

func newJoinDialog(api plugin.API, siteURL, pluginID string, plugin *Plugin) *joinDialog {
	d := &joinDialog{}
	d.API = api
	d.siteURL = siteURL
	d.pluginID = pluginID
	d.plugin = plugin
	return d
}

func (d *joinDialog) Open(triggerID, postID, userID string, game *game) {
	d.API.LogDebug("openJoinDialog is called")

	l := d.plugin.getLocalizer(game.Language)
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

	p := game.GetParticipant(userID)
	if p == nil {
		p = newParticipant(userID)
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

		hand := p.getHand(i)
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
		Options:     cancelOptions,
	})

	dialog := model.Dialog{
		CallbackId:     postID,
		Title:          dialogTitle,
		SubmitLabel:    submitLabel,
		NotifyOnCancel: false,
		State:          game.ID,
		Elements:       elements,
	}

	request := model.OpenDialogRequest{
		TriggerId: triggerID,
		URL:       fmt.Sprintf("%s/plugins/%s/api/v1/janken/join/submit", d.siteURL, d.pluginID),
		Dialog:    dialog,
	}

	d.API.OpenInteractiveDialog(request)
}

// configDialogは"設定"ボタンが押されたときに開くダイアログ
type configDialog struct{ dialog }

func newConfigDialog(api plugin.API, siteURL, pluginID string, plugin *Plugin) *configDialog {
	d := &configDialog{}
	d.API = api
	d.siteURL = siteURL
	d.pluginID = pluginID
	d.plugin = plugin
	return d
}

func (d *configDialog) Open(triggerID, postID string, game *game) {
	d.API.LogDebug("openConfigDialog is called")

	// options for maxRounds
	maxRoundsOptions := []*model.PostActionOptions{}
	for i := 1; i <= maxHands; i++ {
		maxRoundsOptions = append(maxRoundsOptions, &model.PostActionOptions{
			Text: strconv.Itoa(i), Value: strconv.Itoa(i),
		})
	}

	l := d.plugin.getLocalizer(game.Language)
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
			Options:     destroyOptions,
		},
	}

	dialog := model.Dialog{
		CallbackId:     postID,
		Title:          dialogTitle,
		SubmitLabel:    submitLabel,
		NotifyOnCancel: false,
		State:          game.ID,
		Elements:       elements,
	}

	request := model.OpenDialogRequest{
		TriggerId: triggerID,
		URL:       fmt.Sprintf("%s/plugins/%s/api/v1/janken/config/submit", d.siteURL, d.pluginID),
		Dialog:    dialog,
	}

	d.API.OpenInteractiveDialog(request)
}
