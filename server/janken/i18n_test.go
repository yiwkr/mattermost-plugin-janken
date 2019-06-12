package janken

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

type TestFileInfo struct {
	os.FileInfo
	name string
}

func (f TestFileInfo) Name() string {
	return f.name
}

var TestMessages = map[language.Tag][]*i18n.Message{
	language.English: []*i18n.Message{
		&i18n.Message{
			ID:    "NotTemplate",
			Other: "Not template message",
		},
		&i18n.Message{
			ID:    "Template",
			Other: "Template message: {{.Data}}",
		},
	},
	language.Japanese: []*i18n.Message{
		&i18n.Message{
			ID:    "NotTemplate",
			Other: "メッセージ",
		},
		&i18n.Message{
			ID:    "Template",
			Other: "テンプレートメッセージ: {{.Data}}",
		},
	},
}

func TestGetAssetsDir(t *testing.T) {
	for name, test := range map[string]struct {
		SetupPatch  func() *monkey.PatchGuard
		ExpectedDir string
		ShouldError bool
	}{
		"successfully": {
			SetupPatch: func() *monkey.PatchGuard {
				return monkey.Patch(os.Executable, func() (string, error) {
					return "path/to/plugins/server/dist/exe", nil
				})
			},
			ExpectedDir: "path/to/plugins/assets",
			ShouldError: false,
		},
		"failed because os.Executable returns an error": {
			SetupPatch: func() *monkey.PatchGuard {
				return monkey.Patch(os.Executable, func() (string, error) {
					return "", errors.New("error")
				})
			},
			ExpectedDir: "",
			ShouldError: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			patch := test.SetupPatch()
			defer patch.Unpatch()

			d, err := getAssetsDir()

			assert.Equal(test.ExpectedDir, d)

			if test.ShouldError {
				assert.NotNil(err)
			} else {
				assert.Nil(err)
			}
		})
	}
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
			Language:        "",
			DefaultMessage:  &i18n.Message{ID: "NotTemplate", Other: "Not template message"},
			Messages:        TestMessages,
			TemplateData:    nil,
			ExpectedMessage: "Not template message",
		},
		"localize into English": {
			Language:        "en",
			DefaultMessage:  &i18n.Message{ID: "NotTemplate", Other: "Template message:"},
			Messages:        TestMessages,
			TemplateData:    nil,
			ExpectedMessage: "Not template message",
		},
		"localize into English with template": {
			Language:        "en",
			DefaultMessage:  &i18n.Message{ID: "Template", Other: "Template message: {{.Data}}"},
			Messages:        TestMessages,
			TemplateData:    map[string]interface{}{"Data": "template data"},
			ExpectedMessage: "Template message: template data",
		},
		"localize into Japanese": {
			Language:        "ja",
			DefaultMessage:  &i18n.Message{ID: "NotTemplate", Other: "Not template message"},
			Messages:        TestMessages,
			TemplateData:    nil,
			ExpectedMessage: "メッセージ",
		},
		"localize into Japanese with template": {
			Language:        "ja",
			DefaultMessage:  &i18n.Message{ID: "Template", Other: "テストメッセージ 2 {{.Data}}"},
			Messages:        TestMessages,
			TemplateData:    map[string]interface{}{"Data": "テンプレートデータ"},
			ExpectedMessage: "テンプレートメッセージ: テンプレートデータ",
		},
		"unknown language": {
			Language:        "unknown",
			DefaultMessage:  &i18n.Message{ID: "NotTemplate", Other: "Not template message"},
			Messages:        TestMessages,
			TemplateData:    nil,
			ExpectedMessage: "Not template message",
		},
	} {
		b := i18n.NewBundle(language.English)
		for k, v := range test.Messages {
			b.AddMessages(k, v...)
		}

		l := i18n.NewLocalizer(b, test.Language)
		m := Localize(l, test.DefaultMessage, test.TemplateData)

		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(test.ExpectedMessage, m)
		})
	}
}

func TestPluginI18n(t *testing.T) {
	t.Run("InitBundle", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupPatch  func() []*monkey.PatchGuard
			ShouldError bool
		}{
			"successfully": {
				SetupPatch: func() []*monkey.PatchGuard {
					var b *i18n.Bundle
					return []*monkey.PatchGuard{
						monkey.Patch(ioutil.ReadDir, func(string) ([]os.FileInfo, error) {
							return []os.FileInfo{
								TestFileInfo{name: "translate.en.toml"},
								TestFileInfo{name: "active.en.toml"},
							}, nil
						}),
						monkey.PatchInstanceMethod(reflect.TypeOf(b), "LoadMessageFile", func(*i18n.Bundle, string) (*i18n.MessageFile, error) {
							return nil, nil
						}),
					}
				},
				ShouldError: false,
			},
			"failed because getAssetsDir returns an error": {
				SetupPatch: func() []*monkey.PatchGuard {
					return []*monkey.PatchGuard{
						monkey.Patch(getAssetsDir, func() (string, error) {
							return "", errors.New("failed to get assets directory")
						}),
					}
				},
				ShouldError: true,
			},
			"failed because ioutil.ReadDir returns an error": {
				SetupPatch: func() []*monkey.PatchGuard {
					return []*monkey.PatchGuard{
						monkey.Patch(ioutil.ReadDir, func(string) ([]os.FileInfo, error) {
							return nil, errors.New("failed to get files in assets directory")
						}),
					}
				},
				ShouldError: true,
			},
			"failed because bundle.LoadMessageFile returns an error": {
				SetupPatch: func() []*monkey.PatchGuard {
					var b *i18n.Bundle
					return []*monkey.PatchGuard{
						monkey.Patch(ioutil.ReadDir, func(string) ([]os.FileInfo, error) {
							return []os.FileInfo{
								TestFileInfo{name: "translate.en.toml"},
								TestFileInfo{name: "active.en.toml"},
							}, nil
						}),
						monkey.PatchInstanceMethod(reflect.TypeOf(b), "LoadMessageFile", func(*i18n.Bundle, string) (*i18n.MessageFile, error) {
							return nil, errors.New("failed to load message file")
						}),
					}
				},
				ShouldError: true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				patches := test.SetupPatch()
				for _, p := range patches {
					defer p.Unpatch()
				}

				p := &Plugin{}

				_, err := p.InitBundle()

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})

	t.Run("GetLocalizer", func(t *testing.T) {
		for name, test := range map[string]struct {
			Tag string
		}{
			"successfully": {
				Tag: "en",
			},
		} {
			t.Run(name, func(t *testing.T) {
				p := &Plugin{}
				p.bundle = &i18n.Bundle{}

				_ = p.GetLocalizer(test.Tag)
			})
		}
	})
}
