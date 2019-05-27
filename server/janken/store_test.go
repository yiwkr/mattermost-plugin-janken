package janken

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bouk/monkey"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
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
			Id                  string
			SetupAPI            func() *plugintest.API
			JankenGameFromBytes func([]byte) (*JankenGame, error)
			ExpectedGame        *JankenGame
			ShouldError         bool
		}{
			"successfully": {
				Id: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("KVGet", mock.AnythingOfType("string")).Return(nil, nil)
					api.On("LogDebug", "Get", "id", mock.AnythingOfType("string"), "game", mock.AnythingOfType("string")).Return()
					return api
				},
				JankenGameFromBytes: func([]byte) (*JankenGame, error) {
					return NewJankenGame(&TestJankenGameImpl{}), nil
				},
				ExpectedGame: NewJankenGame(&TestJankenGameImpl{}),
				ShouldError:  false,
			},
			"failed with invalid data": {
				Id: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("KVGet", mock.AnythingOfType("string")).Return(nil, nil)
					return api
				},
				JankenGameFromBytes: func([]byte) (*JankenGame, error) {
					return nil, errors.New("error")
				},
				ExpectedGame: nil,
				ShouldError:  true,
			},
			"failed because KVGet returns model.AppError": {
				Id: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("KVGet", mock.AnythingOfType("string")).Return(nil, &model.AppError{})
					return api
				},
				JankenGameFromBytes: func([]byte) (*JankenGame, error) {
					return NewJankenGame(&TestJankenGameImpl{}), nil
				},
				ExpectedGame: nil,
				ShouldError:  true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)
				api := test.SetupAPI()
				s := JankenStore{API: api}

				patch := monkey.Patch(JankenGameFromBytes, test.JankenGameFromBytes)
				defer patch.Unpatch()

				g, err := s.Get(test.Id)

				if g != nil {
					g.Id = ""
					g.CreatedAt = 0
				}
				if test.ExpectedGame != nil {
					test.ExpectedGame.Id = ""
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
					api.On("LogDebug", "Save", "id", mock.AnythingOfType("string"), "game", mock.AnythingOfType("string"))
					// mock.AnythingOfType("[]byte") doesn't quite work.
					// A work around is using mock.AnythingOfType("[]uint8") instead.
					// https://github.com/stretchr/testify/issues/387
					api.On("KVSetWithExpiry", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("int64")).Return(nil)
					return api
				},
				ShouldError: false,
			},
			"failed to convert []byte": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", "Save", "id", mock.AnythingOfType("string"), "game", mock.AnythingOfType("string"))
					return api
				},
				SetupPatch: func() *monkey.PatchGuard {
					return monkey.PatchInstanceMethod(reflect.TypeOf(&JankenGame{}), "ToBytes", func(g *JankenGame) ([]byte, error) {
						return nil, errors.New("error")
					})
				},
				ShouldError: true,
			},
			"failed because KVSetWithExpiry returns model.AppError": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", "Save", "id", mock.AnythingOfType("string"), "game", mock.AnythingOfType("string"))
					api.On("KVSetWithExpiry", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("int64")).Return(&model.AppError{})
					return api
				},
				ShouldError: true,
			},
		} {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)
				api := test.SetupAPI()
				s := JankenStore{API: api}
				g := NewJankenGame(&TestJankenGameImpl{})

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
			Id          string
			SetupAPI    func() *plugintest.API
			ShouldError bool
		}{
			"successfully": {
				Id: "testId",
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("LogDebug", "Delete", "id", mock.AnythingOfType("string"))
					api.On("KVDelete", mock.AnythingOfType("string")).Return(nil)
					return api
				},
				ShouldError: false,
			},
			"failed because KVDelete retunrs model.AppError": {
				Id: "testId",
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
				s := JankenStore{API: api}

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
