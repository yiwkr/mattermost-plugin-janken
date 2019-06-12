package janken

import (
	"testing"

	"github.com/bouk/monkey"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestAPI struct {
	plugin.API
}

func TestJoinDialog(t *testing.T) {
	t.Run("NewJoinDialog", func(t *testing.T) {
		for name, test := range map[string]struct {
			SiteURL          string
			PluginId         string
			ExpectedSiteURL  string
			ExpectedPluginId string
		}{
			"successfully": {
				SiteURL:          "http://example.com/siteurl",
				PluginId:         "com.example.test",
				ExpectedSiteURL:  "http://example.com/siteurl",
				ExpectedPluginId: "com.example.test",
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				d := NewJoinDialog(TestAPI{}, test.SiteURL, test.PluginId, &Plugin{})

				assert.Equal(test.ExpectedSiteURL, d.siteURL)
				assert.Equal(test.ExpectedPluginId, d.pluginId)
			})
		}
	})

	t.Run("Open", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupAPI   func() *plugintest.API
			SetupPatch func() *monkey.PatchGuard
			TriggerId  string
			PostId     string
			UserId     string
			Game       *JankenGame
		}{
			"successfully": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", mock.AnythingOfType("string")).Return()
					api.On("OpenInteractiveDialog", mock.Anything).Return(nil)
					return api
				},
				SetupPatch: func() *monkey.PatchGuard {
					return monkey.Patch(Localize, func(l *i18n.Localizer, defaultMessage *i18n.Message, templateData map[string]interface{}) string {
						return ""
					})
				},
				TriggerId: "t1",
				PostId:    "p1",
				UserId:    "u1",
				Game:      &JankenGame{MaxRounds: 5},
			},
		} {
			t.Run(name, func(t *testing.T) {
				patch := test.SetupPatch()
				defer patch.Unpatch()

				api := test.SetupAPI()

				d := NewJoinDialog(api, "", "", &Plugin{})
				d.Open(test.TriggerId, test.PostId, test.UserId, test.Game)
			})
		}
	})
}

func TestConfigDialog(t *testing.T) {
	t.Run("NewConfigDialog", func(t *testing.T) {
		for name, test := range map[string]struct {
			SiteURL          string
			PluginId         string
			ExpectedSiteURL  string
			ExpectedPluginId string
		}{
			"successfully": {
				SiteURL:          "http://example.com/siteurl",
				PluginId:         "com.example.test",
				ExpectedSiteURL:  "http://example.com/siteurl",
				ExpectedPluginId: "com.example.test",
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				d := NewJoinDialog(TestAPI{}, test.SiteURL, test.PluginId, &Plugin{})

				assert.Equal(test.ExpectedSiteURL, d.siteURL)
				assert.Equal(test.ExpectedPluginId, d.pluginId)
			})
		}
	})

	t.Run("Open", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupAPI   func() *plugintest.API
			SetupPatch func() *monkey.PatchGuard
			TriggerId  string
			PostId     string
			Game       *JankenGame
		}{
			"successfully": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", mock.AnythingOfType("string")).Return()
					api.On("OpenInteractiveDialog", mock.Anything).Return(nil)
					return api
				},
				SetupPatch: func() *monkey.PatchGuard {
					return monkey.Patch(Localize, func(l *i18n.Localizer, defaultMessage *i18n.Message, templateData map[string]interface{}) string {
						return ""
					})
				},
				TriggerId: "t1",
				PostId:    "p1",
				Game:      &JankenGame{},
			},
		} {
			t.Run(name, func(t *testing.T) {
				patch := test.SetupPatch()
				defer patch.Unpatch()

				api := test.SetupAPI()

				d := NewConfigDialog(api, "", "", &Plugin{})
				d.Open(test.TriggerId, test.PostId, test.Game)
			})
		}
	})

}
