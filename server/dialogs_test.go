package main

import (
	"testing"

//	"bou.ke/monkey"
	"github.com/mattermost/mattermost-server/v5/plugin"
//	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
//	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
)

type TestAPI struct {
	plugin.API
}

func TestJoinDialog(t *testing.T) {
	t.Run("newJoinDialog", func(t *testing.T) {
		for name, test := range map[string]struct {
			SiteURL          string
			PluginID         string
			ExpectedSiteURL  string
			ExpectedPluginID string
		}{
			"successfully": {
				SiteURL:          "http://example.com/siteurl",
				PluginID:         "com.example.test",
				ExpectedSiteURL:  "http://example.com/siteurl",
				ExpectedPluginID: "com.example.test",
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				d := newJoinDialog(TestAPI{}, test.SiteURL, test.PluginID, &Plugin{})

				assert.Equal(test.ExpectedSiteURL, d.siteURL)
				assert.Equal(test.ExpectedPluginID, d.pluginID)
			})
		}
	})

	// t.Run("Open", func(t *testing.T) {
	// 	for name, test := range map[string]struct {
	// 		SetupAPI   func() *plugintest.API
	// 		SetupPatch func() *monkey.PatchGuard
	// 		TriggerID  string
	// 		PostID     string
	// 		UserID     string
	// 		game       *game
	// 	}{
	// 		"successfully": {
	// 			SetupAPI: func() *plugintest.API {
	// 				api := &plugintest.API{}
	// 				api.On("LogDebug", mock.AnythingOfType("string")).Return()
	// 				api.On("OpenInteractiveDialog", mock.Anything).Return(nil)
	// 				return api
	// 			},
	// 			SetupPatch: func() *monkey.PatchGuard {
	// 				return monkey.Patch(Localize, func(l *i18n.Localizer, defaultMessage *i18n.Message, templateData map[string]interface{}) string {
	// 					return ""
	// 				})
	// 			},
	// 			TriggerID: "t1",
	// 			PostID:    "p1",
	// 			UserID:    "u1",
	// 			game:      &game{MaxRounds: 5},
	// 		},
	// 	} {
	// 		t.Run(name, func(t *testing.T) {
	// 			patch := test.SetupPatch()
	// 			defer patch.Unpatch()

	// 			api := test.SetupAPI()

	// 			d := newJoinDialog(api, "", "", &Plugin{})
	// 			d.Open(test.TriggerID, test.PostID, test.UserID, test.game)
	// 		})
	// 	}
	// })
}

func TestConfigDialog(t *testing.T) {
	t.Run("newConfigDialog", func(t *testing.T) {
		for name, test := range map[string]struct {
			SiteURL          string
			PluginID         string
			ExpectedSiteURL  string
			ExpectedPluginID string
		}{
			"successfully": {
				SiteURL:          "http://example.com/siteurl",
				PluginID:         "com.example.test",
				ExpectedSiteURL:  "http://example.com/siteurl",
				ExpectedPluginID: "com.example.test",
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				d := newJoinDialog(TestAPI{}, test.SiteURL, test.PluginID, &Plugin{})

				assert.Equal(test.ExpectedSiteURL, d.siteURL)
				assert.Equal(test.ExpectedPluginID, d.pluginID)
			})
		}
	})

	// t.Run("Open", func(t *testing.T) {
	// 	for name, test := range map[string]struct {
	// 		SetupAPI   func() *plugintest.API
	// 		SetupPatch func() *monkey.PatchGuard
	// 		TriggerID  string
	// 		PostID     string
	// 		game       *game
	// 	}{
	// 		"successfully": {
	// 			SetupAPI: func() *plugintest.API {
	// 				api := &plugintest.API{}
	// 				api.On("LogDebug", mock.AnythingOfType("string")).Return()
	// 				api.On("OpenInteractiveDialog", mock.Anything).Return(nil)
	// 				return api
	// 			},
	// 			SetupPatch: func() *monkey.PatchGuard {
	// 				return monkey.Patch(Localize, func(l *i18n.Localizer, defaultMessage *i18n.Message, templateData map[string]interface{}) string {
	// 					return ""
	// 				})
	// 			},
	// 			TriggerID: "t1",
	// 			PostID:    "p1",
	// 			game:      &game{},
	// 		},
	// 	} {
	// 		t.Run(name, func(t *testing.T) {
	// 			patch := test.SetupPatch()
	// 			defer patch.Unpatch()

	// 			api := test.SetupAPI()

	// 			d := newConfigDialog(api, "", "", &Plugin{})
	// 			d.Open(test.TriggerID, test.PostID, test.game)
	// 		})
	// 	}
	// })
}
