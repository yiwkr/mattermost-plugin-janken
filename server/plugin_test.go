package main

import (
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	dummySiteURL = "https://mattermost.example.com"
)

func TestPlugin(t *testing.T) {
	t.Run("OnActivate", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupPatch  func() *monkey.PatchGuard
			ShouldError bool
		}{
			"successfully": {
				SetupPatch: func() *monkey.PatchGuard {
					var p *Plugin
					return monkey.PatchInstanceMethod(reflect.TypeOf(p), "InitBundle",
						func(*Plugin) (*i18n.Bundle, error) { return nil, nil })
				},
				ShouldError: false,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				patch := test.SetupPatch()
				defer patch.Unpatch()

				p := &Plugin{}
				err := p.OnActivate()

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})

	t.Run("OnDeactivate", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupAPI    func() *plugintest.API
			ShouldError bool
		}{
			"successfully": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
					return api
				},
				ShouldError: false,
			},
			"failed because UnregisterCommand returns an error": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(errors.New("failed to unregister command"))
					return api
				},
				ShouldError: true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				p := &Plugin{}
				api := test.SetupAPI()
				p.SetAPI(api)

				err := p.OnDeactivate()

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})

	t.Run("getCommand", func(t *testing.T) {
		for name, test := range map[string]struct {
			Trigger         string
			ExpectedCommand *model.Command
		}{
			"successfully": {
				Trigger: "janken",
				ExpectedCommand: &model.Command{
					Trigger:          "janken",
					DisplayName:      "Janken",
					Description:      "Playing janken",
					AutoComplete:     true,
					AutoCompleteDesc: "Create a janken",
					IconURL: "https://mattermost.example.com/plugins/com.github.yiwkr.mattermost-plugin-janken/janken_choki.png",
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				c := getCommand(dummySiteURL, test.Trigger)

				assert.Equal(test.ExpectedCommand, c)
			})
		}
	})

	t.Run("HasPermission", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupAPI       func() *plugintest.API
			game           *game
			UserID         string
			IsAdmin        bool
			ExpectedResult bool
			ShouldError    bool
		}{
			"permitted because userId equal to Creator": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					return api
				},
				game: &game{
					Creator: "p1",
				},
				UserID:         "p1",
				IsAdmin:        false,
				ExpectedResult: true,
				ShouldError:    false,
			},
			// "permitted because userId is in system admin role id": {
			// 	SetupAPI: func() *plugintest.API {
			// 		api := &plugintest.API{}
			// 		api.On("GetUser", mock.AnythingOfType("string")).Return(&model.User{}, nil)
			// 		return api
			// 	},
			// 	game: &game{
			// 		Creator: "p1",
			// 	},
			// 	UserID:         "p2",
			// 	IsAdmin:        true,
			// 	ExpectedResult: true,
			// 	ShouldError:    false,
			// },
			"not permitted because GetUser returns an error": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("GetUser", mock.AnythingOfType("string")).Return(nil, &model.AppError{})
					return api
				},
				game: &game{
					Creator: "p1",
				},
				UserID:         "p2",
				IsAdmin:        false,
				ExpectedResult: false,
				ShouldError:    true,
			},
			"not permitted": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("GetUser", mock.AnythingOfType("string")).Return(&model.User{}, nil)
					return api
				},
				game: &game{
					Creator: "p1",
				},
				UserID:         "p2",
				IsAdmin:        false,
				ExpectedResult: false,
				ShouldError:    false,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				var u *model.User
				patch := monkey.PatchInstanceMethod(reflect.TypeOf(u), "IsInRole", func(*model.User, string) bool {
					return test.IsAdmin
				})
				defer patch.Unpatch()

				p := &Plugin{}
				api := test.SetupAPI()
				p.SetAPI(api)

				result, err := p.HasPermission(test.game, test.UserID)

				assert.Equal(test.ExpectedResult, result)

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})
}
