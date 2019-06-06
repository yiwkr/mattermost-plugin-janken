package janken

import (
	"net/http"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestResponseWriter struct {
	http.ResponseWriter
	header        http.Header
	writtenHeader []int
	writtenBytes  [][]byte
}

func NewTestResponseWriter() *TestResponseWriter {
	writtenHeader := make([]int, 0)
	writtenBytes := make([][]byte, 0)
	return &TestResponseWriter{
		header:        http.Header{},
		writtenHeader: writtenHeader,
		writtenBytes:  writtenBytes,
	}
}

func (w TestResponseWriter) Header() http.Header {
	return w.header
}

func (w TestResponseWriter) WriteHeader(statusCode int) {
	w.writtenHeader = append(w.writtenHeader, statusCode)
}

func (w TestResponseWriter) Write(b []byte) (int, error) {
	w.writtenBytes = append(w.writtenBytes, b)
	return len(b), nil
}

func TestPluginUtils(t *testing.T) {
	t.Run("sendEphemeralPost", func(t *testing.T) {
		for name, test := range map[string]struct {
			SetupAPI     func() *plugintest.API
			ChannelId    string
			UserId       string
			Message      string
			ExpectedPost *model.Post
		}{
			"successfully": {
				SetupAPI: func() *plugintest.API {
					api := &plugintest.API{}
					api.On("SendEphemeralPost", mock.AnythingOfType("string"), mock.AnythingOfType("*model.Post")).Return(&model.Post{})
					return api
				},
				ChannelId: "test_channel_id",
				UserId:    "test_user_id",
				Message:   "test_message",
			},
		} {
			t.Run(name, func(t *testing.T) {
				api := test.SetupAPI()
				p := &Plugin{}
				p.SetAPI(api)

				p.sendEphemeralPost(test.ChannelId, test.UserId, test.Message)
			})
		}
	})
}

func TestAppendMessage(t *testing.T) {
	for name, test := range map[string]struct {
		Post            *model.Post
		AppendedMessage string
		Args            []interface{}
		ExpectedMessage string
	}{
		"append message": {
			Post: &model.Post{
				Message: "original message",
			},
			AppendedMessage: "appended message",
			Args:            nil,
			ExpectedMessage: "original message\nappended message",
		},
		"append message with args": {
			Post: &model.Post{
				Message: "original message",
			},
			AppendedMessage: "appended message: %s=%d",
			Args:            []interface{}{"value", 1},
			ExpectedMessage: "original message\nappended message: value=1",
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			p := appendMessage(test.Post, test.AppendedMessage, test.Args...)

			assert.Equal(test.ExpectedMessage, p.Message)
		})
	}
}

func TestWritePostActionIntegrationResponse(t *testing.T) {
	for name, test := range map[string]struct {
		ExpectedHeader     http.Header
		ExpectedStatusCode []int
	}{
		"successfully": {
			ExpectedHeader: http.Header{
				"Content-Type": []string{"application/json"},
			},
			ExpectedStatusCode: []int{http.StatusOK},
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			response := &model.PostActionIntegrationResponse{}
			w := NewTestResponseWriter()
			r := &http.Request{}
			writePostActionIntegrationResponse(response, w, r)

			assert.Equal(test.ExpectedHeader, w.header)
		})
	}
}
