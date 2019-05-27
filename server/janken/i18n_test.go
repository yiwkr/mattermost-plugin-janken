package janken

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var TestMessages = map[language.Tag][]*i18n.Message{
	language.English: []*i18n.Message{
		&i18n.Message{
			ID: "NotTemplate",
			Other: "Not template message",
		},
		&i18n.Message{
			ID: "Template",
			Other: "Template message: {{.Data}}",
		},
	},
	language.Japanese: []*i18n.Message{
		&i18n.Message{
			ID: "NotTemplate",
			Other: "メッセージ",
		},
		&i18n.Message{
			ID: "Template",
			Other: "テンプレートメッセージ: {{.Data}}",
		},
	},
}

func TestLocalize(t *testing.T) {
	for name, test := range map[string]struct {
		Language        string
		DefaultMessage  *i18n.Message
		Messages        map[language.Tag][]*i18n.Message
		TemplateData    map[string]interface{}
		ExpectedMessage string
	}{
		"use default language": {
			Language: "",
			DefaultMessage: &i18n.Message{ID: "NotTemplate", Other: "Not template message"},
			Messages: TestMessages,
			TemplateData: nil,
			ExpectedMessage: "Not template message",
		},
		"localize into English": {
			Language: "en",
			DefaultMessage: &i18n.Message{ID: "NotTemplate", Other: "Template message:"},
			Messages: TestMessages,
			TemplateData: nil,
			ExpectedMessage: "Not template message",
		},
		"localize into English with template": {
			Language: "en",
			DefaultMessage: &i18n.Message{ID: "Template", Other: "Template message: {{.Data}}"},
			Messages: TestMessages,
			TemplateData: map[string]interface{}{"Data": "template data"},
			ExpectedMessage: "Template message: template data",
		},
		"localize into Japanese": {
			Language: "ja",
			DefaultMessage: &i18n.Message{ID: "NotTemplate", Other: "Not template message"},
			Messages: TestMessages,
			TemplateData: nil,
			ExpectedMessage: "メッセージ",
		},
		"localize into Japanese with template": {
			Language: "ja",
			DefaultMessage: &i18n.Message{ID: "Template", Other: "テストメッセージ 2 {{.Data}}"},
			Messages: TestMessages,
			TemplateData: map[string]interface{}{"Data": "テンプレートデータ"},
			ExpectedMessage: "テンプレートメッセージ: テンプレートデータ",
		},
		"unknown language": {
			Language: "unknown",
			DefaultMessage: &i18n.Message{ID: "NotTemplate", Other: "Not template message"},
			Messages: TestMessages,
			TemplateData: nil,
			ExpectedMessage: "Not template message",
		},
	}{
		b := i18n.NewBundle(language.English)
		for k, v := range test.Messages {
			b.AddMessages(k, v...)
		}
		l := i18n.NewLocalizer(b, test.Language)
		m := Localize(l, test.DefaultMessage, test.TemplateData)

		t.Run(name, func(t *testing.T){
			assert := assert.New(t)
			assert.Equal(test.ExpectedMessage, m)
		})
	}
}
