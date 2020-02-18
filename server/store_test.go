package main

import (
	"errors"
	"testing"

	"bou.ke/monkey"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStore(t *testing.T) {
	t.Run("NewStore", func(t *testing.T) {
		assert := assert.New(t)
		api := &plugintest.API{}

		s := NewStore(api)

		assert.Equal(api, s.API)
	})
}

func TestJankenStore(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		for name, test := range map[string]struct {
			ID                  string
			SetupAPI            func() *plugintest.API
			gameFromBytes func([]byte) (*game, error)
			ExpectedGame        *game
			ShouldError         bool
		}{
			"successfully": {
				ID: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("KVGet", mock.AnythingOfType("string")).Return(nil, nil)
					api.On("LogDebug", "Get", "id", mock.AnythingOfType("string"), "game", mock.AnythingOfType("string")).Return()
					return api
				},
				gameFromBytes: func([]byte) (*game, error) {
					return newGame(&TestGameImpl{}), nil
				},
				ExpectedGame: newGame(&TestGameImpl{}),
				ShouldError:  false,
			},
			"failed with invalid data": {
				ID: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("KVGet", mock.AnythingOfType("string")).Return(nil, nil)
					return api
				},
				gameFromBytes: func([]byte) (*game, error) {
					return nil, errors.New("error")
				},
				ExpectedGame: nil,
				ShouldError:  true,
			},
			"failed because KVGet returns model.AppError": {
				ID: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("KVGet", mock.AnythingOfType("string")).Return(nil, &model.AppError{})
					return api
				},
				gameFromBytes: func([]byte) (*game, error) {
					return newGame(&TestGameImpl{}), nil
				},
				ExpectedGame: nil,
				ShouldError:  true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)
				api := test.SetupAPI()
				s := jankenStore{API: api}

				patch := monkey.Patch(gameFromBytes, test.gameFromBytes)
				defer patch.Unpatch()

				g, err := s.Get(test.ID)

				if g != nil {
					g.ID = ""
					g.CreatedAt = 0
				}
				if test.ExpectedGame != nil {
					test.ExpectedGame.ID = ""
					test.ExpectedGame.CreatedAt = 0
				}

				assert.Equal(test.ExpectedGame, g)
				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})

	t.Run("Save", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupAPI    func() *plugintest.API
			SetupPatch  func() *monkey.PatchGuard
			ShouldError bool
		}{
			"successfully": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", "Save", "id", mock.AnythingOfType("string"),
						"game", mock.AnythingOfType("string"))
					// mock.AnythingOfType("[]byte") doesn't quite work.
					// A work around is using mock.AnythingOfType("[]uint8") instead.
					// https://github.com/stretchr/testify/issues/387
					api.On("KVSetWithExpiry",
						mock.AnythingOfType("string"),
						mock.AnythingOfType("[]uint8"),
						mock.AnythingOfType("int64")).Return(nil)
					return api
				},
				ShouldError: false,
			},
			"failed because KVSetWithExpiry returns model.AppError": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", "Save", "id", mock.AnythingOfType("string"),
						"game", mock.AnythingOfType("string"))
					api.On("KVSetWithExpiry", mock.AnythingOfType("string"),
						mock.AnythingOfType("[]uint8"),
						mock.AnythingOfType("int64"),
					).Return(&model.AppError{})
					return api
				},
				ShouldError: true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)
				api := test.SetupAPI()
				s := jankenStore{API: api}
				g := newGame(&TestGameImpl{})

				if test.SetupPatch != nil {
					patch := test.SetupPatch()
					defer patch.Unpatch()
				}

				err := s.Save(g)

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		for name, test := range map[string]struct {
			ID          string
			SetupAPI    func() *plugintest.API
			ShouldError bool
		}{
			"successfully": {
				ID: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", "Delete", "id", mock.AnythingOfType("string"))
					api.On("KVDelete", mock.AnythingOfType("string")).Return(nil)
					return api
				},
				ShouldError: false,
			},
			"failed because KVDelete retunrs model.AppError": {
				ID: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", "Delete", "id", mock.AnythingOfType("string"))
					api.On("KVDelete", mock.AnythingOfType("string")).Return(&model.AppError{})
					return api
				},
				ShouldError: true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)
				api := test.SetupAPI()
				s := jankenStore{API: api}

				err := s.Delete("testId")

				if test.ShouldError {
					assert.NotNil(err)
				} else {
					assert.Nil(err)
				}
			})
		}
	})
}
