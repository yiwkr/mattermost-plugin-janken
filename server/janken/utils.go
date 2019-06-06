package janken

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/model"
)

func (p *Plugin) sendEphemeralPost(channelId, userId, message string) *model.Post {
	post := &model.Post{}
	post.ChannelId = channelId
	post.UserId = userId
	post.Message = message
	post.AddProp("sent_by_plugin", true)
	return p.API.SendEphemeralPost(userId, post)
}

func appendMessage(post *model.Post, format string, args ...interface{}) *model.Post {
	message := fmt.Sprintf(format, args...)
	post.Message = fmt.Sprintf("%s\n%s", post.Message, message)
	return post
}

func writePostActionIntegrationResponse(response *model.PostActionIntegrationResponse, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response.ToJson())
}
