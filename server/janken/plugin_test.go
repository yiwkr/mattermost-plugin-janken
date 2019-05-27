package janken

//import (
//	"testing"
//
//	"github.com/mattermost/mattermost-server/plugin/plugintest"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestPlugin(t *testing.T) {
//	t.Run("OnActivate", func(t *testing.T){
//		for name, test := range map[string]struct {
//			SetupAPI    func() *plugintest.API
//			ShouldError bool
//		}{
//			"successfully": {
//				SetupAPI: func() *plugintest.API {
//					api := &plugintest.API{}
//					return api
//				},
//				ShouldError: false,
//			},
//		}{
//			t.Run(name, func(t *testing.T){
//				assert := assert.New(t)
//				api := test.SetupAPI()
//				p := &Plugin{}
//				p.SetAPI(api)
//
//				err := p.OnActivate()
//
//				if test.ShouldError {
//					assert.NotNil(err)
//				} else {
//					assert.Nil(err)
//				}
//			})
//		}
//	})
//}
