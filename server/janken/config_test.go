package janken

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPluginConfig(t *testing.T) {
	t.Run("OnConfigurationChange", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupAPI         func() *plugintest.API
			SetupPatch       func() *monkey.PatchGuard
			PreConfiguration *configuration
			ShouldError      bool
		}{
			"successfully": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LoadPluginConfiguration", mock.Anything).Return(nil)
					api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.Anything).Return(nil)
					api.On("RegisterCommand", mock.Anything).Return(nil)
					api.On("GetConfig").Return(nil)
					return api
				},
				SetupPatch: func() *monkey.PatchGuard {
					var p *Plugin
					return monkey.PatchInstanceMethod(reflect.TypeOf(p), "InitBundle", func(*Plugin) (*i18n.Bundle, error) {
						return &i18n.Bundle{}, nil
					})
				},
				PreConfiguration: &configuration{Trigger: "janken"},
				ShouldError:      false,
			},
			"failed because LoadPluginConfiguration returns an error": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LoadPluginConfiguration", mock.Anything).Return(errors.New("failed to load configuration"))
					return api
				},
				PreConfiguration: &configuration{},
				ShouldError:      true,
			},
			"failed because UnregisterCommand returns an error": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LoadPluginConfiguration", mock.Anything).Return(nil)
					api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.Anything).Return(errors.New("failed to unregister command"))
					return api
				},
				PreConfiguration: &configuration{Trigger: "janken"},
				ShouldError:      true,
			},
			"failed because RegisterCommand returns an error": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LoadPluginConfiguration", mock.Anything).Return(nil)
					api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.Anything).Return(nil)
					api.On("RegisterCommand", mock.Anything).Return(errors.New("failed to register command"))
					return api
				},
				PreConfiguration: &configuration{Trigger: "janken"},
				ShouldError:      true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				if test.SetupPatch != nil {
					patch := test.SetupPatch()
					defer patch.Unpatch()
				}

				p := &Plugin{}
				api := test.SetupAPI()
				p.SetAPI(api)

				p.configuration = test.PreConfiguration
				err := p.OnConfigurationChange()

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})

	t.Run("setConfiguration", func(t *testing.T) {
		c := &configuration{Trigger: "janken"}

		for name, test := range map[string]struct {
			PreConfiguration *configuration
			PanicMessage     string
			ShouldPanic      bool
		}{
			"same configuration": {
				PreConfiguration: c,
				PanicMessage:     "setConfiguration called with the existing configuration",
				ShouldPanic:      true,
			},
			"different configuration": {
				PreConfiguration: &configuration{Trigger: "differentConfig"},
				PanicMessage:     "",
				ShouldPanic:      false,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				p := &Plugin{}

				p.configuration = test.PreConfiguration

				if test.ShouldPanic {
					defer func() {
						err := recover()
						assert.Equal(test.PanicMessage, err)
					}()
				} else {
					defer func() {
						assert.Equal(p.configuration, c)
					}()
				}

				p.setConfiguration(c)
			})
		}
	})
}
