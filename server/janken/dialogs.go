package janken

import (
	"fmt"
	"strconv"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

// ジャンケンで出せる手
var HandsOptions []*model.PostActionOptions = []*model.PostActionOptions{
	{
		Text:  Hands["rock"],
		Value: "rock",
	},
	{
		Text:  Hands["scissors"],
		Value: "scissors",
	},
	{
		Text:  Hands["paper"],
		Value: "paper",
	},
}

// 参加取消の選択肢
var CancelOptions []*model.PostActionOptions = []*model.PostActionOptions{
	{
		Text: "-",
		Value: "false",
	},
	{
		Text: "cancel",
		Value: "true",
	},
}

// ゲーム削除の選択肢
var DestroyOptions []*model.PostActionOptions = []*model.PostActionOptions{
	{
		Text: "-",
		Value: "false",
	},
	{
		Text: "destroy",
		Value: "true",
	},
}

type Dialog struct {
	API      plugin.API
	siteURL  string
	pluginId string
}

// JoinDialogは"参加"ボタンが押されたときに開くダイアログ
type JoinDialog struct { Dialog }

func NewJoinDialog(api plugin.API, siteURL, pluginId string) *JoinDialog {
	d := &JoinDialog{}
	d.API = api
	d.siteURL = siteURL
	d.pluginId = pluginId
	return d
}

func (d *JoinDialog) Open(triggerId, postId, userId string, game *JankenGame) {
	d.API.LogDebug("openJoinDialog is called")

	cancelLabel := "Cancel"
	dialogTitle := "To join the janken game"
	submitLabel := "Save"

	participant := game.GetParticipant(userId)
	if participant == nil {
		participant = NewParticipant(userId)
	}

	// 手の入力フォームを追加
	elements := []model.DialogElement{}
	for i := 0; i < game.MaxRounds; i++ {
		hand := participant.GetHand(i)

		i1 := i + 1  // 1-base index
		elements = append(elements, model.DialogElement{
			DisplayName: fmt.Sprintf("Hand %d", i1),
			Name:        fmt.Sprintf("hand%d", i1),
			Type:        "select",
			Placeholder: Hands[hand],
			Default:     hand,
			Optional:    true,
			Options:     HandsOptions,
			HelpText:    fmt.Sprintf("Choose hand %d", i1),
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
type ConfigDialog struct { Dialog }

func NewConfigDialog(api plugin.API, siteURL, pluginId string) *ConfigDialog {
	d := &ConfigDialog{}
	d.API = api
	d.siteURL = siteURL
	d.pluginId = pluginId
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

	maxRoundsLabel := "Max rounds"
	destroyLabel := "Destroy this game"
	dialogTitle := "Config"
	submitLabel := "Save"

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
